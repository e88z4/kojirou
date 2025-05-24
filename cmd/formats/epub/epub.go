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
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"sort"

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
func GenerateEPUB(manga mangadex.Manga, widepage kindle.WidepagePolicy, crop bool, ltr bool) (*epub.Epub, func(), error) {
	// Basic validation
	if manga.Info.Title == "" {
		// Instead of error, use a default title to match test expectations
		manga.Info.Title = "Untitled Manga"
	}
	if len(manga.Volumes) == 0 {
		return nil, nil, fmt.Errorf("manga must have at least one volume")
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
	// Always add a default stylesheet as a real file
	cssContent := "body { margin: 0; padding: 0; } img { display: block; max-width: 100%; height: auto; }"
	cssTempPath := filepath.Join(os.TempDir(), "kojirou-style.css")
	err := os.WriteFile(cssTempPath, []byte(cssContent), 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write temp CSS file: %w", err)
	}
	cssHref, _ := e.AddCSS(cssTempPath, "style.css")
	cssHref = "css/style.css" // go-epub always puts CSS in css/ subdir

	var tempImagePaths []string
	// Track temp CSS for cleanup
	tempImagePaths = append(tempImagePaths, cssTempPath)

	// Add covers for each volume as images
	coverIndex := 1
	for volID, vol := range manga.Volumes {
		if vol.Cover != nil {
			coverName := fmt.Sprintf("cover-%v.jpg", volID)
			imgPath := filepath.Join(os.TempDir(), coverName)
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

	// For each volume and chapter, add pages with deterministic image names
	for volID, vol := range manga.Volumes {
		// Check for empty chapters in volume
		if len(vol.Chapters) == 0 {
			return nil, nil, fmt.Errorf("volume %v has no chapters", volID)
		}
		// Sort chapter keys to ensure deterministic chapter order
		var chapKeys []mangadex.Identifier
		for k := range vol.Chapters {
			chapKeys = append(chapKeys, k)
		}
		sort.Slice(chapKeys, func(i, j int) bool {
			return chapKeys[i].Less(chapKeys[j])
		})
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
			var htmlContent string
			// Sort page keys to ensure deterministic order
			var pageKeys []int
			for k := range chap.Pages {
				pageKeys = append(pageKeys, k)
			}
			sort.Ints(pageKeys)
			for _, k := range pageKeys {
				img := chap.Pages[k]
				if img == nil {
					return nil, nil, fmt.Errorf("chapter %q has nil image page", sectionTitle)
				}
				// Use CropAndSplit for wide page handling
				processedImages := kindle.CropAndSplit(img, widepage, crop, ltr)
				for splitIdx, splitImg := range processedImages {
					imgName := fmt.Sprintf("page-%v-%v-%d", volID, chapKey, k)
					if len(processedImages) > 1 {
						imgName = fmt.Sprintf("%s-%d.jpg", imgName, splitIdx)
					} else {
						imgName = imgName + ".jpg"
					}
					imgPath := filepath.Join(os.TempDir(), imgName)
					f, err := os.Create(imgPath)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to create temp image: %w", err)
					}
					err = jpeg.Encode(f, splitImg, nil)
					f.Close()
					if err != nil {
						return nil, nil, fmt.Errorf("failed to encode image: %w", err)
					}
					// Add image to EPUB for consistency
					imgHref, err := e.AddImage(imgPath, imgName)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to add image: %w", err)
					}
					htmlContent += fmt.Sprintf("<div><img src=\"%s\" alt=\"Page image\"/></div>", imgHref)
					tempImagePaths = append(tempImagePaths, imgPath)
				}
			}
			if htmlContent == "" {
				htmlContent = "<p>(No images in this chapter)</p>"
			}
			// Prepend stylesheet link in a full XHTML document structure
			sectionHTML := `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>` + sectionTitle + `</title>
  <link rel="stylesheet" type="text/css" href="` + cssHref + `"/>
</head>
<body>
<h1>` + sectionTitle + `</h1>` + htmlContent + `
</body>
</html>`
			_, err := e.AddSection(sectionHTML, sectionTitle, "", "")
			if err != nil {
				return nil, nil, fmt.Errorf("failed to add section: %w", err)
			}
		}
	}

	// Only clean up temp images after EPUB is written
	cleanup := func() {
		for _, path := range tempImagePaths {
			_ = os.Remove(path)
		}
	}

	return e, cleanup, nil
}
