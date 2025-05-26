package cmd

import (
	"fmt"
	"strings"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/filter"
	"github.com/leotaku/kojirou/cmd/formats"
	"github.com/leotaku/kojirou/cmd/formats/disk"
	"github.com/leotaku/kojirou/cmd/formats/download"
	epubpkg "github.com/leotaku/kojirou/cmd/formats/epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	"github.com/leotaku/kojirou/cmd/formats/output"
	"github.com/leotaku/kojirou/cmd/formats/progress"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

func run() error {
	manga, err := download.MangadexSkeleton(identifierArg)
	if err != nil {
		return fmt.Errorf("skeleton: %w", err)
	}

	chapters, err := getChapters(*manga)
	if err != nil {
		return fmt.Errorf("chapters: %w", err)
	}
	*manga = manga.WithChapters(chapters)

	// Parse formats early to validate user input
	selectedFormats, err := formats.ParseFormats(FormatsArg)
	if err != nil {
		return fmt.Errorf("invalid formats: %w", err)
	}

	// Print summary and exit if dry run
	formats.PrintSummary(manga)
	if dryRunArg {
		return nil
	}

	// Log format selection
	formatStrings := make([]string, len(selectedFormats))
	for i, format := range selectedFormats {
		formatStrings[i] = string(format)
	}
	fmt.Printf("Generating formats: %s\n", strings.Join(formatStrings, ", "))

	covers, err := getCovers(manga)
	if err != nil {
		return fmt.Errorf("covers: %w", err)
	}
	*manga = manga.WithCovers(covers)

	dir := kindle.NewNormalizedDirectory(outArg, manga.Info.Title, kindleFolderModeArg)
	for _, volume := range manga.Sorted() {
		if err := HandleVolume(*manga, volume, dir); err != nil {
			return fmt.Errorf("volume %v: %w", volume.Info.Identifier, err)
		}
	}

	return nil
}

// 6. Report consolidated status at the end
func HandleVolume(skeleton md.Manga, volume md.Volume, dir kindle.NormalizedDirectory) error {
	// Create a titled progress bar with volume information
	p := progress.TitledProgress(fmt.Sprintf("Volume: %v", volume.Info.Identifier))

	// Get selected formats
	selectedFormats, err := formats.ParseFormats(FormatsArg)
	if err != nil {
		p.Cancel(fmt.Sprintf("Format selection error: %v", err))
		return fmt.Errorf("parse formats: %w", err)
	}

	// Check if we can skip the entire volume processing
	if !forceArg {
		allExist := true
		for _, format := range selectedFormats {
			if !dir.HasWithExtension(volume.Info.Identifier, string(format)) {
				allExist = false
				break
			}
		}

		if allExist {
			p.Cancel("Skipped (all formats exist)")
			return nil
		}
	}

	// Load pages (shared operation for all formats)
	p.SetFormat("pages")
	pages, err := getPages(volume, p)
	if err != nil {
		return fmt.Errorf("pages: %w", err)
	}
	p.SetFormat("")

	mangaForVolume := skeleton.WithChapters(volume.Sorted()).WithPages(pages)

	// Common formatting for title
	title := fmt.Sprintf("%v: %v",
		skeleton.Info.Title,
		volume.Info.Identifier.StringFilled(fillVolumeNumberArg, 0, false),
	)

	// Track which formats succeeded and failed
	formatStatus := make(map[formats.FormatType]string)

	// Common parameters for all formats
	widepagePolicy := kindle.WidepagePolicy(widepageArg)

	// Create a shared EPUB for both EPUB and KEPUB formats
	var sharedEpub *epub.Epub
	needsEpub := false
	for _, format := range selectedFormats {
		if format == formats.FormatEpub || format == formats.FormatKepub {
			needsEpub = true
			break
		}
	}

	if needsEpub {
		var epubErr error
		var cleanup func()
		sharedEpub, cleanup, epubErr = epubpkg.GenerateEPUBProd(
			mangaForVolume,
			widepagePolicy,
			autocropArg,
			leftToRightArg,
		)
		if epubErr != nil {
			p.Cancel("Error generating EPUB base")
			return fmt.Errorf("generate epub base: %w", epubErr)
		}
		if cleanup != nil {
			defer cleanup()
		}
		p.SetFormat("")
	}

	// Create a multi-format progress tracker for summary
	formatStrings := make([]string, len(selectedFormats))
	for i, f := range selectedFormats {
		formatStrings[i] = string(f)
	}
	summaryProgress := progress.MultiFormatStatusProgress(
		fmt.Sprintf("Formats - %v", volume.Info.Identifier),
		formatStrings,
	)
	defer summaryProgress.Done()

	// Process each format with format-specific progress reporting
	for _, format := range selectedFormats {
		// Skip if the format already exists and we're not forcing regeneration
		if !forceArg && dir.HasWithExtension(volume.Info.Identifier, string(format)) {
			formatStatus[format] = "Skipped (already exists)"
			summaryProgress.FormatCompleted(string(format), "Skipped")
			continue
		}

		// Update the main progress to show which format is being processed
		p.SetFormat(string(format))

		// Create format-specific progress
		formatProgress := progress.FormatVanishingProgress("Writing", string(format))
		var outputFormat output.FormatOutput
		var formatErr error

		switch format {
		case formats.FormatMobi:
			mobi := kindle.GenerateMOBI(
				mangaForVolume,
				widepagePolicy,
				autocropArg,
				leftToRightArg,
			)
			mobi.RightToLeft = !leftToRightArg
			mobi.Title = title
			outputFormat = &output.MobiOutput{Book: &mobi}

		case formats.FormatEpub:
			// We already generated the EPUB above
			outputFormat = &output.EpubOutput{Epub: sharedEpub}

		case formats.FormatKepub:
			// We already generated the EPUB above, use it for KEPUB
			outputFormat = &output.KepubOutput{Epub: sharedEpub}
		}

		// Write the format to disk
		if err := dir.WriteFormat(volume.Info.Identifier, outputFormat, formatProgress); err != nil {
			formatStatus[format] = fmt.Sprintf("Error: %v", err)
			formatProgress.CancelWithFormat(string(format), "Error")
			summaryProgress.FormatCompleted(string(format), "Error")
			formatErr = err
		} else {
			formatStatus[format] = "Success"
			formatProgress.Done()
			summaryProgress.FormatCompleted(string(format), "Success")
		}

		// We don't fail immediately on format errors to allow other formats to be processed
		// Instead, we track the error and report it at the end
		if formatErr != nil {
			formatStatus[format] = fmt.Sprintf("Error: %v", formatErr)
		}
	}

	// Reset format indicator before final status
	p.SetFormat("")

	// Check if any format failed
	var errorFormats []string
	for format, status := range formatStatus {
		if strings.HasPrefix(status, "Error") {
			errorFormats = append(errorFormats, fmt.Sprintf("%s (%s)", format, status))
		}
	}

	if len(errorFormats) > 0 {
		p.Cancel(fmt.Sprintf("Errors: %s", strings.Join(errorFormats, ", ")))
		return fmt.Errorf("errors processing formats: %s", strings.Join(errorFormats, ", "))
	}

	// All formats succeeded
	p.Cancel("All formats completed")
	return nil
}

func getChapters(manga md.Manga) (md.ChapterList, error) {
	chapters, err := download.MangadexChapters(manga.Info.ID)
	if err != nil {
		return nil, fmt.Errorf("mangadex: %w", err)
	}

	if diskArg != "" {
		p := progress.VanishingProgress("Disk...")
		diskChapters, err := disk.LoadChapters(diskArg, language.Make(languageArg), p)
		if err != nil {
			p.Cancel("Error")
			return nil, fmt.Errorf("disk: %w", err)
		}
		p.Done()
		chapters = append(chapters, diskChapters...)
	}

	chapters, err = filterAndSortFromFlags(chapters)
	if err != nil {
		return nil, fmt.Errorf("filter: %w", err)
	}

	// Ensure chapters from disk are preferred
	if diskArg != "" {
		chapters = chapters.SortBy(func(a md.ChapterInfo, b md.ChapterInfo) bool {
			return a.GroupNames.String() == "Filesystem" && b.GroupNames.String() != "Filesystem"
		})
	}

	return filter.RemoveDuplicates(chapters), nil
}

func getCovers(manga *md.Manga) (md.ImageList, error) {
	p := progress.VanishingProgress("Covers")
	covers, err := download.MangadexCovers(manga, p)
	if err != nil {
		p.Cancel("Error")
		return nil, fmt.Errorf("mangadex: %w", err)
	}
	p.Done()

	// Covers from disk should automatically be preferred, because
	// they appear later in the list and thus should override the
	// earlier downloaded covers.
	if diskArg != "" {
		p := progress.VanishingProgress("Disk...")
		diskCovers, err := disk.LoadCovers(diskArg, p)
		if err != nil {
			p.Cancel("Error")
			return nil, fmt.Errorf("disk: %w", err)
		}
		p.Done()
		covers = append(covers, diskCovers...)
	}

	return covers, nil
}

func getPages(volume md.Volume, p progress.CliProgress) (md.ImageList, error) {
	mangadexPages, err := download.MangadexPages(volume.Sorted().FilterBy(func(ci md.ChapterInfo) bool {
		return ci.GroupNames.String() != "Filesystem"
	}), download.DataSaverPolicy(dataSaverArg), p)
	if err != nil {
		p.Cancel("Error")
		return nil, fmt.Errorf("mangadex: %w", err)
	}
	diskPages, err := disk.LoadPages(volume.Sorted().FilterBy(func(ci md.ChapterInfo) bool {
		return ci.GroupNames.String() == "Filesystem"
	}), p)
	if err != nil {
		p.Cancel("Error")
		return nil, fmt.Errorf("disk: %w", err)
	}
	p.Done()

	return append(mangadexPages, diskPages...), nil
}

func filterAndSortFromFlags(cl md.ChapterList) (md.ChapterList, error) {
	if languageArg != "" {
		lang := language.Make(languageArg)
		cl = filter.FilterByLanguage(cl, lang)
	}
	if groupsFilter != "" {
		cl = filter.FilterByRegex(cl, "GroupNames", groupsFilter)
	}
	if volumesFilter != "" {
		ranges := filter.ParseRanges(volumesFilter)
		cl = filter.FilterByIdentifier(cl, "VolumeIdentifier", ranges)
	}
	if chaptersFilter != "" {
		ranges := filter.ParseRanges(chaptersFilter)
		cl = filter.FilterByIdentifier(cl, "Identifier", ranges)
	}

	switch rankArg {
	case "newest":
		cl = filter.SortByNewest(cl)
	case "newest-total":
		cl = filter.SortByNewestGroup(cl)
	case "views":
		cl = filter.SortByViews(cl)
	case "views-total":
		cl = filter.SortByGroupViews(cl)
	case "most":
		cl = filter.SortByMost(cl)
	default:
		return nil, fmt.Errorf(`not a valid ranking algorithm: "%v"`, rankArg)
	}

	return cl, nil
}
