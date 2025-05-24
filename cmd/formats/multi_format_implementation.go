// Package formats implements multi-format output for manga conversion
package formats

import (
	"fmt"
	"os"
	"strings"

	"github.com/bmaupin/go-epub"
	epub_format "github.com/leotaku/kojirou/cmd/formats/epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	"github.com/leotaku/kojirou/cmd/formats/output"
	"github.com/leotaku/kojirou/cmd/formats/progress"
	md "github.com/leotaku/kojirou/mangadex"
)

// HandleVolumeMultiFormat processes a manga volume and generates requested formats
func HandleVolumeMultiFormat(
	skeleton md.Manga,
	volume md.Volume,
	dir kindle.NormalizedDirectory,
	formatsArg string,
	forceArg bool,
	autocropArg bool,
	leftToRightArg bool,
	widepageArg string,
	fillVolumeNumberArg int,
) error {
	p := progress.TitledProgress(fmt.Sprintf("Volume: %v", volume.Info.Identifier))

	// Get selected formats
	selectedFormats, err := ParseFormats(formatsArg)
	if err != nil {
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
			p.SetFormatMessage("", fmt.Sprintf("All formats already exist: %s", formatsArg))
			p.Done()
			return nil
		}
	}

	// Track which formats succeeded and failed
	formatStatus := make(map[FormatType]string)

	// Common parameters for all formats
	var widepagePolicy kindle.WidepagePolicy
	switch widepageArg {
	case "preserve":
		widepagePolicy = kindle.WidepagePolicyPreserve
	case "split":
		widepagePolicy = kindle.WidepagePolicySplit
	case "preserve-and-split":
		widepagePolicy = kindle.WidepagePolicyPreserveAndSplit
	case "split-and-preserve":
		widepagePolicy = kindle.WidepagePolicySplitAndPreserve
	default:
		widepagePolicy = kindle.WidepagePolicyPreserve
	}

	// Process each format with format-specific progress reporting
	for _, format := range selectedFormats {
		// Skip if the format already exists and we're not forcing regeneration
		if !forceArg && dir.HasWithExtension(volume.Info.Identifier, string(format)) {
			formatStatus[format] = "Skipped (already exists)"
			continue
		}

		formatProgress := progress.TitledProgress(fmt.Sprintf("Writing %s...", format))
		var err error

		var output output.FormatOutput
		switch format {
		case FormatEpub:
			// Assign all 3 return values and handle cleanup
			epubObj, cleanup, err := epub_format.GenerateEPUB(md.Manga{
				Info:    md.MangaInfo{Title: skeleton.Info.Title},
				Volumes: map[md.Identifier]md.Volume{volume.Info.Identifier: volume},
			}, widepagePolicy, autocropArg, leftToRightArg)
			if err == nil {
				output = &EPUBFormatOutput{
					epub:     epubObj,
					filePath: dir.Path(volume.Info.Identifier, "epub"),
				}
				if cleanup != nil {
					defer cleanup()
				}
			}
		case FormatKepub:
			// Assign all 3 return values and handle cleanup
			epubObj, cleanup, err := epub_format.GenerateEPUB(md.Manga{
				Info:    md.MangaInfo{Title: skeleton.Info.Title},
				Volumes: map[md.Identifier]md.Volume{volume.Info.Identifier: volume},
			}, widepagePolicy, autocropArg, leftToRightArg)
			if err == nil {
				output = &KEPUBFormatOutput{
					epub:     epubObj,
					filePath: dir.Path(volume.Info.Identifier, "kepub.epub"),
				}
				if cleanup != nil {
					defer cleanup()
				}
			}
		default:
			err = fmt.Errorf("unsupported format: %s", format)
		}

		if err != nil {
			formatStatus[format] = fmt.Sprintf("Error: %v", err)
			formatProgress.Cancel("Error")
			continue
		}

		err = dir.WriteFormat(volume.Info.Identifier, output, formatProgress)
		if err != nil {
			formatStatus[format] = fmt.Sprintf("Error: %v", err)
			formatProgress.Cancel("Error")
		} else {
			formatStatus[format] = "Success"
			formatProgress.Done()
		}
	}

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
	p.Done()
	return nil
}

// EPUBFormatOutput implements output.FormatOutput for EPUB format
type EPUBFormatOutput struct {
	epub     *epub.Epub
	filePath string
}

// GetBytes returns the EPUB as a byte array
func (e *EPUBFormatOutput) GetBytes() ([]byte, error) {
	// Write the EPUB to a temporary file first
	tempPath := e.filePath + ".tmp"
	err := e.epub.Write(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to write EPUB: %w", err)
	}
	defer os.Remove(tempPath)

	// Read the file back as bytes
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read EPUB data: %w", err)
	}

	return data, nil
}

// Write writes the EPUB to a file
func (e *EPUBFormatOutput) Write() error {
	return e.epub.Write(e.filePath)
}

// Extension returns the file extension for this format
func (e *EPUBFormatOutput) Extension() string {
	return "epub"
}

// KEPUBFormatOutput implements output.FormatOutput for KEPUB format
type KEPUBFormatOutput struct {
	epub     *epub.Epub
	filePath string
}

// GetBytes returns the KEPUB as a byte array
func (k *KEPUBFormatOutput) GetBytes() ([]byte, error) {
	// Convert EPUB to KEPUB
	data, err := kepubconv.ConvertToKEPUB(k.epub)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to KEPUB: %w", err)
	}

	return data, nil
}

// Write writes the KEPUB to a file
func (k *KEPUBFormatOutput) Write() error {
	data, err := k.GetBytes()
	if err != nil {
		return err
	}

	return os.WriteFile(k.filePath, data, 0644)
}

// Extension returns the file extension for this format
func (k *KEPUBFormatOutput) Extension() string {
	return "kepub.epub"
}
