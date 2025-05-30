// Package epub provides functionality for generating EPUB ebook files from manga data.
//
// This package contains the core generators for EPUB files and utilities for converting
// to Kobo-specific KEPUB format. It handles manga data processing, image transformation,
// and proper navigation structure for manga reading.
//
// The main functions provided are:
//   - GenerateEPUB: Creates an EPUB object from manga data
//   - ConvertToKEPUB: Transforms a standard EPUB into a Kobo-compatible KEPUB
//
// Usage example:
//
//	// Generate EPUB
//	epubObj, err := epub.GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
//	if err != nil {
//	    return err
//	}
//
//	// Write to file
//	err = epubObj.Write("output.epub")
//	if err != nil {
//	    return err
//	}
//
//	// Convert to KEPUB
//	kepubData, err := epub.ConvertToKEPUB(epubObj)
//	if err != nil {
//	    return err
//	}
//
//	// Write KEPUB data to file
//	err = os.WriteFile("output.kepub.epub", kepubData, 0644)
package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/image/draw"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	"github.com/leotaku/kojirou/mangadex"
)

// GenerateEPUB creates an EPUB file from manga data
//
// This function processes manga data and converts it into a structured EPUB document,
// with proper chapter organization, navigation, and image processing.
//
// Parameters:
//   - manga: The manga data to be converted
//   - widepage: Policy for handling wide pages (preserve, split, etc.)
//   - crop: Whether to automatically crop images
//   - ltr: Reading direction (true for left-to-right, false for right-to-left)
//
// Returns:
//   - *epub.Epub: A pointer to the generated EPUB object
//   - func(): A cleanup function to delete temporary images
//   - error: Any error encountered during generation
//
// The function handles various aspects of EPUB creation:
//   - Setting proper metadata (title, author, language)
//   - Creating a hierarchical structure with volumes and chapters
//   - Processing images according to specified policies
//   - Setting correct reading direction
//   - Generating navigation elements
func GenerateEPUB(tempDir string, manga mangadex.Manga, widepage kindle.WidepagePolicy, crop bool, ltr bool) (*epub.Epub, func(), error) {
	// Basic validation
	if manga.Info.Title == "" {
		// Instead of error, use a default title to match test expectations
		manga.Info.Title = "Untitled Manga"
	}
	if len(manga.Volumes) == 0 {
		return nil, nil, fmt.Errorf("manga has no volumes")
	}

	e := epub.NewEpub(manga.Info.Title)
	if len(manga.Info.Authors) > 0 {
		e.SetAuthor(manga.Info.Authors[0])
	}
	// Set identifier if present
	if manga.Info.ID != "" {
		e.SetIdentifier(manga.Info.ID)
	}
	// Always set language to en (default)
	e.SetLang("en")
	cssContent := "body { margin: 0; padding: 0; } img { display: block; max-width: 100%; height: auto; }"
	cssTempPath := filepath.Join(tempDir, "style.css")
	err := os.WriteFile(cssTempPath, []byte(cssContent), 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write temp CSS file: %w", err)
	}
	cssHref, _ := e.AddCSS(cssTempPath, "style.css")

	var tempImagePaths []string
	// Track temp CSS for cleanup
	tempImagePaths = append(tempImagePaths, cssTempPath)

	// Add covers for each volume as images
	coverIndex := 1
	for volID, vol := range manga.Volumes {
		// Validate cover dimensions
		if vol.Cover != nil {
			bounds := vol.Cover.Bounds()
			if bounds.Dx() <= 0 || bounds.Dy() <= 0 || bounds.Min.X < 0 || bounds.Min.Y < 0 || bounds.Max.X <= bounds.Min.X || bounds.Max.Y <= bounds.Min.Y {
				return nil, nil, fmt.Errorf("invalid cover image dimensions: %+v", bounds)
			}
			coverName := fmt.Sprintf("cover-%v.jpg", volID)
			imgPath := filepath.Join(tempDir, coverName)
			f, err := os.Create(imgPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create temp cover image: %w", err)
			}
			err = jpeg.Encode(f, vol.Cover, nil)
			f.Close()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to encode cover image: %w", err)
			}
			// Add cover image to EPUB and manifest
			imgHref, err := e.AddImage(imgPath, coverName)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to add cover image: %w", err)
			}
			// Set as cover if first volume
			if coverIndex == 1 {
				e.SetCover(imgHref, "")
			}
			tempImagePaths = append(tempImagePaths, imgPath)
			coverIndex++
		}
	}

	// Parallel image processing worker pool
	type imgJob struct {
		img      image.Image
		imgName  string
		imgPath  string
		resultCh chan error
	}

	const maxWorkers = 4 // Tune for your CPU
	imgJobs := make(chan imgJob, maxWorkers*2)
	var wg sync.WaitGroup
	jpegBuf := &bytes.Buffer{}
	jpegMu := &sync.Mutex{} // Protect jpegBuf

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range imgJobs {
				jpegMu.Lock()
				jpegBuf.Reset()
				err := jpeg.Encode(jpegBuf, job.img, nil)
				jpegMu.Unlock()
				if err == nil {
					f, ferr := os.Create(job.imgPath)
					if ferr == nil {
						_, werr := f.Write(jpegBuf.Bytes())
						f.Close()
						if werr != nil {
							err = werr
						}
					} else {
						err = ferr
					}
				}
				job.resultCh <- err
			}
		}()
	}

	// Track chapters that actually had a section created
	type chapterKey struct {
		volID   mangadex.Identifier
		chapKey mangadex.Identifier
	}
	addedChapters := make(map[chapterKey]bool)

	// For each volume and chapter, add pages with deterministic image names
	for volID, vol := range manga.Volumes {
		// Add a section for the volume at the start of the volume loop
		volNum := volID.StringFilled(1, 0, false)
		volTitle := "Volume " + volNum
		volSectionHTML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>%s</title>
  <link rel="stylesheet" type="text/css" href="%s"/>
</head>
<body><h1>%s</h1></body>
</html>`, volTitle, cssHref, volTitle)
		_, _ = e.AddSection(volSectionHTML, volTitle, fmt.Sprintf("volume-%v.xhtml", volID), "volume")

		// Check for empty chapters in volume
		if len(vol.Chapters) == 0 {
			return nil, nil, fmt.Errorf("volume %v has no chapters", volID)
		}
		// Sort chapter keys to ensure deterministic chapter order
		chapKeys := make([]mangadex.Identifier, 0, len(vol.Chapters))
		for k := range vol.Chapters {
			chapKeys = append(chapKeys, k)
		}
		sort.Slice(chapKeys, func(i, j int) bool { return chapKeys[i].Less(chapKeys[j]) })
		for _, chapKey := range chapKeys {
			chap := vol.Chapters[chapKey]
			sectionTitle := chap.Info.Title
			if sectionTitle == "" {
				sectionTitle = "Untitled Chapter"
			}
			// Check for empty pages in chapter
			if len(chap.Pages) == 0 {
				return nil, nil, fmt.Errorf("chapter %q has no pages", sectionTitle)
			}
			// Build HTML for this chapter with all images, in sorted order
			var htmlBuilder strings.Builder
			// Sort page keys to ensure deterministic order
			pageKeys := make([]int, 0, len(chap.Pages))
			for k := range chap.Pages {
				pageKeys = append(pageKeys, k)
			}
			sort.Ints(pageKeys)
			imgIdx := 0
			for _, k := range pageKeys {
				img := chap.Pages[k]
				if img == nil {
					// Return an error for nil images instead of skipping
					return nil, nil, fmt.Errorf("nil image found in chapter %q, page %d", sectionTitle, k)
				}
				bounds := img.Bounds()
				if bounds.Dx() <= 0 || bounds.Dy() <= 0 || bounds.Min.X < 0 || bounds.Min.Y < 0 || bounds.Max.X <= bounds.Min.X || bounds.Max.Y <= bounds.Min.Y {
					return nil, nil, fmt.Errorf("invalid image dimensions in chapter %q: %+v", sectionTitle, bounds)
				}
				// Use CropAndSplit for wide page handling
				processedImages := kindle.CropAndSplit(img, widepage, crop, ltr)
				// Release reference to original image
				chap.Pages[k] = nil
				for splitIdx, splitImg := range processedImages {
					bounds := splitImg.Bounds()
					if bounds.Dx() <= 0 || bounds.Dy() <= 0 || bounds.Min.X < 0 || bounds.Min.Y < 0 || bounds.Max.X <= bounds.Min.X || bounds.Max.Y <= bounds.Min.Y {
						return nil, nil, fmt.Errorf("invalid split image dimensions in chapter %q: %+v", sectionTitle, bounds)
					}
					// Scale image if wider than 1600px
					if splitImg.Bounds().Dx() > 1600 {
						splitImg = scaleImageToMaxWidth(splitImg, 1600)
					}
					imgName := fmt.Sprintf("page-%v-%v-%d", volID, chapKey, k)
					if len(processedImages) > 1 {
						imgName = fmt.Sprintf("%s-%d.jpg", imgName, splitIdx)
					} else {
						imgName = imgName + ".jpg"
					}
					imgPath := filepath.Join(tempDir, imgName)
					resultCh := make(chan error, 1)
					imgJobs <- imgJob{img: splitImg, imgName: imgName, imgPath: imgPath, resultCh: resultCh}
					err := <-resultCh
					if err != nil {
						return nil, nil, fmt.Errorf("failed to encode/write image: %w", err)
					}
					imgHref, err := e.AddImage(imgPath, imgName)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to add image: %w", err)
					}
					htmlBuilder.WriteString(fmt.Sprintf("<div><img src=\"%s\" alt=\"Page image\"/></div>", imgHref))
					tempImagePaths = append(tempImagePaths, imgPath)
					// Release reference to split image
					processedImages[splitIdx] = nil
					imgIdx++
				}
			}
			if htmlBuilder.Len() == 0 {
				htmlBuilder.WriteString("<p>(No images in this chapter)</p>")
			}
			// Prepend stylesheet link in a full XHTML document structure
			sectionHTML := `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>` + sectionTitle + `</title>
  <link rel="stylesheet" type="text/css" href="` + cssHref + `"/>
</head>
<body>
<h1>` + sectionTitle + `</h1>` + htmlBuilder.String() + `
</body>
</html>`
			sectionID := fmt.Sprintf("chapter-%v-%v.xhtml", volID, chapKey)
			sectionPath, err := e.AddSection(sectionHTML, sectionTitle, sectionID, "chapter")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to add section %s: %v\n", sectionID, err)
				return nil, nil, fmt.Errorf("failed to add section: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Added section: %s at %s\n", sectionID, sectionPath)
			// Mark this chapter as added
			addedChapters[chapterKey{volID, chapKey}] = true
			// Encourage GC after each chapter
			runtime.GC()
		}
		// Encourage GC after each volume
		runtime.GC()
	}
	close(imgJobs)
	wg.Wait()

	// After all chapters are added, generate nav.xhtml
	// Always use nested structure for all manga (even single-volume)
	navHTML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
  <head>
    <title>` + manga.Info.Title + `</title>
  </head>
  <body>
    <nav epub:type="toc">
      <h1>Table of Contents</h1>
      <ol>
`
	// Volumes and chapters (always nested)
	volKeys := make([]mangadex.Identifier, 0, len(manga.Volumes))
	for k := range manga.Volumes {
		volKeys = append(volKeys, k)
	}
	sort.Slice(volKeys, func(i, j int) bool { return volKeys[i].Less(volKeys[j]) })
	// Always use nested structure for navigation
	for _, volID := range volKeys {
		vol := manga.Volumes[volID]
		volNum := volID.StringFilled(1, 0, false)
		volTitle := "Volume " + volNum
		// Emit <li>Volume N<ol>...</ol></li> with NO indentation or newline between <li> and volume title
		navHTML += "        <li>" + volTitle + "<ol>\n"
		chapKeys := make([]mangadex.Identifier, 0, len(vol.Chapters))
		for k := range vol.Chapters {
			chapKeys = append(chapKeys, k)
		}
		sort.Slice(chapKeys, func(i, j int) bool { return chapKeys[i].Less(chapKeys[j]) })
		chapterCount := 0
		for _, chapKey := range chapKeys {
			if !addedChapters[chapterKey{volID, chapKey}] {
				continue
			}
			chap := vol.Chapters[chapKey]
			chapTitle := chap.Info.Title
			if chapTitle == "" {
				chapTitle = "Untitled Chapter"
			}
			sectionID := fmt.Sprintf("chapter-%v-%v.xhtml", volID, chapKey)
			navHTML += "            <li><a href=\"xhtml/" + sectionID + "\">" + chapTitle + "</a></li>\n"
			chapterCount++
		}
		navHTML += "          </ol>\n"
		navHTML += "        </li>\n"
	}
	// Optionally add navigation link at the end
	navHTML += "        <li><a href=\"nav.xhtml\">Navigation</a></li>\n"
	navHTML += `      </ol>
    </nav>
  </body>
</html>
`
	// Add nav.xhtml as a section with nav property
	fmt.Fprintf(os.Stderr, "[DEBUG] nav.xhtml about to be added:\n%s\n", navHTML)
	_, _ = e.AddSection(navHTML, "Navigation", "nav.xhtml", "nav")
	fmt.Fprintf(os.Stderr, "[DEBUG] nav.xhtml AddSection complete\n")

	/*
	   Cleanup function: Must be called only after the EPUB is fully written.
	   If called before e.Write(), temp image files will be deleted too early and EPUB writing will fail.
	*/
	cleanup := func() {
		for _, path := range tempImagePaths {
			_ = os.Remove(path)
		}
	}

	return e, cleanup, nil
}

func GenerateEPUBProd(manga mangadex.Manga, widepage kindle.WidepagePolicy, crop bool, ltr bool) (*epub.Epub, func(), error) {
	tempDir, err := os.MkdirTemp("", "epub-prod-*")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	epubObj, cleanup, err := GenerateEPUB(tempDir, manga, widepage, crop, ltr)
	prodCleanup := func() {
		cleanup()
		_ = os.RemoveAll(tempDir)
	}
	return epubObj, prodCleanup, err
}

func scaleImageToMaxWidth(src image.Image, maxWidth int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= maxWidth {
		return src
	}
	newWidth := maxWidth
	newHeight := int(float64(height) * float64(maxWidth) / float64(width))
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

// PatchEPUBNavManifest ensures nav.xhtml is listed with properties="nav" in the OPF manifest inside the EPUB file.
func PatchEPUBNavManifest(epubPath string) error {
	// Open the EPUB as a zip archive
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find the OPF file and read all files into memory
	var opfName string
	files := make(map[string][]byte)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		files[f.Name] = data
		if strings.HasSuffix(f.Name, ".opf") {
			opfName = f.Name
		}
	}

	if opfName == "" {
		return nil // No OPF found, nothing to patch
	}

	// Patch the OPF manifest
	orig := string(files[opfName])
	lines := strings.Split(orig, "\n")
	for i, line := range lines {
		if strings.Contains(line, "nav.xhtml") && !strings.Contains(line, "properties=\"nav\"") {
			lines[i] = strings.Replace(line, "/>", " properties=\"nav\"/>", 1)
		}
	}
	files[opfName] = []byte(strings.Join(lines, "\n"))

	// Write a new EPUB file
	tmpPath := epubPath + ".patched"
	w, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	zipw := zip.NewWriter(w)
	for name, data := range files {
		fh := &zip.FileHeader{Name: name, Method: zip.Deflate}
		fh.SetMode(0644)
		fw, err := zipw.CreateHeader(fh)
		if err != nil {
			zipw.Close()
			w.Close()
			return err
		}
		_, err = fw.Write(data)
		if err != nil {
			zipw.Close()
			w.Close()
			return err
		}
	}
	zipw.Close()
	w.Close()

	// Replace the original EPUB
	return os.Rename(tmpPath, epubPath)
}
