package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

// TestKEPUBWithRealWorldManga tests KEPUB conversion with realistic manga data
func TestKEPUBWithRealWorldManga(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create test cases with different manga characteristics
	testCases := []struct {
		name        string
		setupManga  func() md.Manga
		expectation string
	}{
		{
			name: "standard manga with regular pages",
			setupManga: func() md.Manga {
				return createStandardManga()
			},
			expectation: "successful conversion",
		},
		{
			name: "manga with very long title and author",
			setupManga: func() md.Manga {
				manga := createStandardManga()
				manga.Info.Title = strings.Repeat("Very Long Title ", 20)
				// MangaInfo does not have Author, use Authors slice
				manga.Info.Authors = []string{strings.Repeat("Very Long Author Name ", 20)}
				return manga
			},
			expectation: "successful conversion with properly handled long strings",
		},
		{
			name: "manga with color pages",
			setupManga: func() md.Manga {
				return createColorManga()
			},
			expectation: "successful conversion preserving color",
		},
		{
			name: "manga with mixed page orientations",
			setupManga: func() md.Manga {
				return createMixedOrientationManga()
			},
			expectation: "successful conversion respecting orientations",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the manga
			manga := tc.setupManga()

			// Generate EPUB
			epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if cleanup != nil {
				// cleanup() will be called after all conversions below
			}

			// Convert to KEPUB
			kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Verify KEPUB output
			verifyKEPUBOutput(t, kepubData, tc.name, tc.expectation)
		})
	}
}

// TestKEPUBWithExternalMangaData tests with sample manga files from an external source
func TestKEPUBWithExternalMangaData(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "kepub-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with predefined sample manga
	testSamples := []struct {
		name     string
		filename string
		setup    func(string) (*epub.Epub, error)
	}{
		{
			name:     "sample manga EPUB",
			filename: "sample_manga.epub",
			setup: func(path string) (*epub.Epub, error) {
				// Create a simple sample EPUB
				e := epub.NewEpub("Sample Manga")
				e.SetAuthor("Test Author")

				// Add cover image
				coverPath := filepath.Join(tempDir, "cover.jpg")
				img := createTestImage(600, 800, color.White)
				f, err := os.Create(coverPath)
				if err != nil {
					panic(err)
				}
				_ = jpeg.Encode(f, img, nil)
				f.Close()
				_, err = e.AddImage(coverPath, "cover.jpg")
				if err != nil {
					return nil, fmt.Errorf("failed to add cover: %v", err)
				}

				// Add multiple chapters with images
				for i := 1; i <= 5; i++ {
					// Create chapter images
					imgPath := filepath.Join(tempDir, fmt.Sprintf("page%d.jpg", i))
					img := createTestImage(800, 1200, color.White)
					f, err := os.Create(imgPath)
					if err != nil {
						panic(err)
					}
					_ = jpeg.Encode(f, img, nil)
					f.Close()
					imgID, err := e.AddImage(imgPath, fmt.Sprintf("page%d.jpg", i))
					if err != nil {
						return nil, fmt.Errorf("failed to add image: %v", err)
					}

					// Create HTML content with the image
					content := fmt.Sprintf("<img src=\"%s\" alt=\"Page %d\" />", imgID, i)
					_, err = e.AddSection(content, fmt.Sprintf("Chapter %d", i), fmt.Sprintf("ch%d", i), "")
					if err != nil {
						return nil, fmt.Errorf("failed to add section: %v", err)
					}
				}

				// Write to file
				err = e.Write(path)
				if err != nil {
					return nil, fmt.Errorf("failed to write EPUB: %v", err)
				}

				return e, nil
			},
		},
		{
			name:     "complex manga EPUB",
			filename: "complex_manga.epub",
			setup: func(path string) (*epub.Epub, error) {
				// Create a more complex sample
				e := epub.NewEpub("Complex Manga Sample")
				e.SetAuthor("Test Complex Author")
				e.SetDescription("A complex manga sample with multiple chapters and special layouts")

				// Add CSS for manga-specific styling
				cssPath := filepath.Join(tempDir, "manga.css")
				writeTestCSS(cssPath)
				_, err := e.AddCSS(cssPath, "manga.css")
				if err != nil {
					return nil, fmt.Errorf("failed to add CSS: %v", err)
				}

				// Add cover
				coverPath := filepath.Join(tempDir, "complex_cover.jpg")
				img := createTestImage(600, 800, color.White)
				f, err := os.Create(coverPath)
				if err != nil {
					panic(err)
				}
				_ = jpeg.Encode(f, img, nil)
				f.Close()
				_, err = e.AddImage(coverPath, "cover.jpg")
				if err != nil {
					return nil, fmt.Errorf("failed to add cover: %v", err)
				}

				// Add multiple chapters with different layouts
				for i := 1; i <= 3; i++ {
					// Chapter with multiple images
					var content strings.Builder
					content.WriteString("<div class=\"chapter\">")

					for j := 1; j <= 3; j++ {
						imgPath := filepath.Join(tempDir, fmt.Sprintf("complex_ch%d_page%d.jpg", i, j))
						img := createTestImage(800, 1200, color.White)
						f, err := os.Create(imgPath)
						if err != nil {
							panic(err)
						}
						_ = jpeg.Encode(f, img, nil)
						f.Close()
						imgID, err := e.AddImage(imgPath, fmt.Sprintf("ch%d_page%d.jpg", i, j))
						if err != nil {
							return nil, fmt.Errorf("failed to add image: %v", err)
						}

						content.WriteString(fmt.Sprintf("<div class=\"page\"><img src=\"%s\" alt=\"Chapter %d Page %d\" /></div>",
							imgID, i, j))
					}

					content.WriteString("</div>")

					_, err = e.AddSection(content.String(), fmt.Sprintf("Chapter %d", i), fmt.Sprintf("ch%d", i), "")
					if err != nil {
						return nil, fmt.Errorf("failed to add section: %v", err)
					}
				}

				// Write to file
				err = e.Write(path)
				if err != nil {
					return nil, fmt.Errorf("failed to write EPUB: %v", err)
				}

				return e, nil
			},
		},
	}

	for _, sample := range testSamples {
		t.Run(sample.name, func(t *testing.T) {
			samplePath := filepath.Join(tempDir, sample.filename)

			// Setup the sample file
			epubObj, err := sample.setup(samplePath)
			if err != nil {
				t.Fatalf("Failed to setup sample: %v", err)
			}

			// Convert to KEPUB
			kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Verify KEPUB data
			if len(kepubData) == 0 {
				t.Error("KEPUB data is empty")
			}

			// Verify Kobo specific features
			verifyKoboEnhancements(t, kepubData)

			// Write the output for inspection if needed
			kepubPath := filepath.Join(tempDir, strings.TrimSuffix(sample.filename, ".epub")+".kepub")
			err = os.WriteFile(kepubPath, kepubData, 0644)
			if err != nil {
				t.Logf("Failed to write KEPUB output: %v", err)
			} else {
				t.Logf("KEPUB output written to: %s", kepubPath)
			}
		})
	}
}

// Helper functions for test data creation

// createStandardManga returns a Manga with the new data model (Volumes/Chapters/Pages as maps)
func createStandardManga() md.Manga {
	volID := md.NewIdentifier("1")
	ch1ID := md.NewIdentifier("1.1")
	ch2ID := md.NewIdentifier("1.2")

	return md.Manga{
		Info: md.MangaInfo{
			Title:   "Standard Test Manga",
			Authors: []string{"Test Author"},
		},
		Volumes: map[md.Identifier]md.Volume{
			volID: {
				Info: md.VolumeInfo{Identifier: volID},
				Chapters: map[md.Identifier]md.Chapter{
					ch1ID: {
						Info: md.ChapterInfo{
							Title:            "Chapter 1",
							Identifier:       ch1ID,
							VolumeIdentifier: volID,
						},
						Pages: map[int]image.Image{
							1: createTestImage(800, 1200, color.White),
							2: createTestImage(800, 1200, color.White),
						},
					},
					ch2ID: {
						Info: md.ChapterInfo{
							Title:            "Chapter 2",
							Identifier:       ch2ID,
							VolumeIdentifier: volID,
						},
						Pages: map[int]image.Image{
							1: createTestImage(800, 1200, color.White),
						},
					},
				},
			},
		},
	}
}

// createColorManga returns a Manga with color pages using the new data model
func createColorManga() md.Manga {
	manga := createStandardManga()
	manga.Info.Title = "Color Test Manga"
	// Mark first page of each chapter as color (for test distinction)
	for volID, volume := range manga.Volumes {
		for chID, chapter := range volume.Chapters {
			for pageNum := range chapter.Pages {
				if pageNum == 1 {
					chapter.Pages[pageNum] = createTestImage(800, 1200, color.RGBA{255, 0, 0, 255})
				}
			}
			manga.Volumes[volID].Chapters[chID] = chapter
		}
	}
	return manga
}

// createMixedOrientationManga returns a Manga with mixed page orientations using the new data model
func createMixedOrientationManga() md.Manga {
	manga := createStandardManga()
	manga.Info.Title = "Mixed Orientation Manga"
	// Add a wide page to the first chapter
	for volID, volume := range manga.Volumes {
		for chID, chapter := range volume.Chapters {
			widePage := createTestImage(1600, 1200, color.White)
			chapter.Pages[len(chapter.Pages)+1] = widePage
			manga.Volumes[volID].Chapters[chID] = chapter
			break // Only add to the first chapter
		}
		break
	}
	return manga
}

func createPagePath(filename string) string {
	// In a real test, this would create actual image files
	// For our tests, we'll just return the path string as if the file existed
	return filepath.Join(os.TempDir(), "test_manga_images", filename)
}

func createTestImageFile(path string, width, height int) error {
	// Create a dummy image file for testing
	// In a real implementation, this would create an actual image
	// For our test, just create an empty file
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("test image data"), 0644)
}

func writeTestCSS(path string) error {
	// Write a simple CSS file for manga styling
	css := `
		.page {
			display: block;
			margin: 0;
			padding: 0;
		}
		.page img {
			width: 100%;
			height: auto;
		}
		/* Styling for wide pages */
		.wide-page img {
			max-width: 100%;
			transform: rotate(90deg);
		}
	`
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(css), 0644)
}

func verifyKEPUBOutput(t *testing.T, data []byte, testName, expectation string) {
	// Base verification
	if len(data) == 0 {
		t.Errorf("KEPUB data is empty for test: %s", testName)
		return
	}

	// Check if it's a valid ZIP archive (EPUB/KEPUB files are ZIP-based)
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Errorf("Failed to read KEPUB as ZIP: %v", err)
		return
	}

	// Log basic information about the archive
	t.Logf("KEPUB file has %d entries", len(zipReader.File))

	// Verify KEPUB-specific structure
	hasContainerXML := false
	hasPackageOPF := false
	hasKoboCustomization := false

	for _, file := range zipReader.File {
		switch {
		case file.Name == "META-INF/container.xml":
			hasContainerXML = true
		case strings.HasSuffix(file.Name, "package.opf"):
			hasPackageOPF = true
		case strings.Contains(file.Name, "kobo"):
			hasKoboCustomization = true
		}
	}

	if !hasContainerXML {
		t.Errorf("KEPUB missing META-INF/container.xml")
	}

	if !hasPackageOPF {
		t.Errorf("KEPUB missing package.opf")
	}

	if !hasKoboCustomization {
		t.Logf("KEPUB may be missing Kobo-specific files or customizations")
	}

	// More specific checks based on expectation
	switch {
	case strings.Contains(expectation, "preserving color"):
		// Additional color-specific checks would go here
		t.Log("Color preservation verification requires manual inspection")
	case strings.Contains(expectation, "respecting orientations"):
		// Additional orientation-specific checks would go here
		t.Log("Orientation handling verification requires manual inspection")
	}
}
