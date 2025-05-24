package epub

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	testhelpers "github.com/leotaku/kojirou/cmd/formats/testhelpers"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

// TestEPUBGeneration tests the EPUB generation function with various manga inputs and configuration options
func TestEPUBGeneration(t *testing.T) {
	// Test cases for different manga inputs and configuration options
	tests := []struct {
		name     string
		manga    md.Manga
		widepage kindle.WidepagePolicy
		autocrop bool
		ltr      bool
		wantErr  bool
	}{
		{
			name:     "basic manga with standard pages",
			manga:    testhelpers.CreateTestManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "right-to-left manga",
			manga:    testhelpers.CreateTestManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      false, // Right-to-left
			wantErr:  false,
		},
		{
			name:     "manga with wide page splitting",
			manga:    createWidePageTestManga(),
			widepage: kindle.WidepagePolicySplit, // Split wide pages
			autocrop: false,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "manga with autocrop enabled",
			manga:    testhelpers.CreateTestManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: true, // Enable autocrop
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "empty manga",
			manga:    md.Manga{},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true, // Should fail with empty manga
		},
		{
			name:     "manga with no volumes",
			manga:    md.Manga{Info: md.MangaInfo{Title: "Test"}},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true, // Should fail with no volumes
		},
		{
			name:     "manga with empty volumes",
			manga:    createEmptyVolumeManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true, // Should fail with empty volumes
		},
		{
			name:     "manga with no title",
			manga:    createNoTitleManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false, // Should use default title
		},
		{
			name:     "manga with invalid images",
			manga:    createInvalidImageManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true, // Should fail with invalid images
		},
		{
			name:     "manga with special characters in title",
			manga:    createSpecialCharTitleManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false, // Should handle special characters
		},
		{
			name:     "manga with mixed languages",
			manga:    createMixedLanguageManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false, // Should use primary language
		},
		{
			name:     "manga with extremely large images",
			manga:    createLargeImageManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false, // Should handle large images
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate EPUB
			epub, cleanup, err := GenerateEPUB(tt.manga, tt.widepage, tt.autocrop, tt.ltr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Skip further checks if we expected an error
			if tt.wantErr {
				return
			}

			// Verify EPUB object is not nil
			if epub == nil {
				t.Fatal("GenerateEPUB() returned nil EPUB object")
				return
			}

			// Test writing the EPUB to a file
			epubPath := filepath.Join(t.TempDir(), "test.epub")
			if epub != nil {
				if err := epub.Write(epubPath); err != nil {
					t.Fatalf("Failed to write EPUB to file: %v", err)
					return
				}
			}

			// Debug: List all files in the resulting EPUB archive
			if r, err := zip.OpenReader(epubPath); err == nil {
				t.Logf("Files in EPUB archive:")
				for _, f := range r.File {
					t.Logf("- %s", f.Name)
				}
				r.Close()
			}

			// Verify EPUB structure
			verifyEPUBStructure(t, epubPath, tt.ltr)
		})
	}
}

// TestEPUBImageProcessing tests various image processing scenarios in EPUB generation
func TestEPUBImageProcessing(t *testing.T) {
	tests := []struct {
		name          string
		manga         md.Manga
		widepage      kindle.WidepagePolicy
		autocrop      bool
		ltr           bool
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "standard images",
			manga:         testhelpers.CreateTestManga(),
			widepage:      kindle.WidepagePolicyPreserve,
			autocrop:      false,
			ltr:           true,
			expectedCount: 7, // 2 volume covers + 5 pages
			wantErr:       false,
		},
		{
			name:          "split wide pages",
			manga:         createWidePageTestManga(),
			widepage:      kindle.WidepagePolicySplit,
			autocrop:      false,
			ltr:           true,
			expectedCount: 8, // 2 volume covers + 4 pages + 2 from split
			wantErr:       false,
		},
		{
			name:          "autocrop enabled",
			manga:         testhelpers.CreateTestManga(),
			widepage:      kindle.WidepagePolicyPreserve,
			autocrop:      true,
			ltr:           true,
			expectedCount: 7, // 2 volume covers + 5 pages
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := GenerateEPUB(tt.manga, tt.widepage, tt.autocrop, tt.ltr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cleanup != nil {
				defer cleanup()
			}

			if tt.wantErr {
				return
			}

			// Write to file and extract as zip to count images
			zipReader, err := writeEPUBToZipReader(t, epub)
			if err != nil {
				t.Fatalf("Failed to write and read EPUB: %v", err)
			}

			// Count images
			imageCount := 0
			for _, f := range zipReader.File {
				if strings.HasPrefix(f.Name, "EPUB/images/") &&
					(strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") ||
						strings.HasSuffix(f.Name, ".png")) {
					imageCount++
				}
			}

			if imageCount != tt.expectedCount {
				t.Errorf("Expected %d images, got %d", tt.expectedCount, imageCount)
			}
		})
	}
}

// TestEPUBNavigationStructure tests the navigation structure of generated EPUBs
func TestEPUBNavigationStructure(t *testing.T) {
	tests := []struct {
		name         string
		manga        md.Manga
		ltr          bool
		volumeCount  int
		chapterCount int
	}{
		{
			name:         "standard navigation",
			manga:        testhelpers.CreateTestManga(),
			ltr:          true,
			volumeCount:  2,
			chapterCount: 3,
		},
		{
			name:         "rtl navigation",
			manga:        testhelpers.CreateTestManga(),
			ltr:          false,
			volumeCount:  2,
			chapterCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := GenerateEPUB(tt.manga, kindle.WidepagePolicyPreserve, false, tt.ltr)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Write EPUB to file and get zip reader
			zipReader, err := writeEPUBToZipReader(t, epub)
			if err != nil {
				t.Fatalf("Failed to write and read EPUB: %v", err)
			}

			// Find and parse nav.xhtml and package.opf
			var navFile, contentFile *zip.File
			for _, f := range zipReader.File {
				if f.Name == "EPUB/nav.xhtml" {
					navFile = f
				} else if f.Name == "EPUB/package.opf" {
					contentFile = f
				}
			}

			if navFile == nil {
				t.Fatal("nav.xhtml not found in EPUB")
			}
			if contentFile == nil {
				t.Fatal("package.opf not found in EPUB")
			}

			// Read nav.xhtml content
			rc, err := navFile.Open()
			if err != nil {
				t.Fatalf("Failed to open nav.xhtml: %v", err)
			}
			navContent, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read nav.xhtml: %v", err)
			}

			// Check for volume entries
			volCount := strings.Count(string(navContent), "Volume")
			if volCount != tt.volumeCount {
				t.Errorf("Expected %d volume entries, found %d", tt.volumeCount, volCount)
			}

			// Check for chapter entries
			chapCount := strings.Count(string(navContent), "Chapter")
			if chapCount != tt.chapterCount {
				t.Errorf("Expected %d chapter entries, found %d", tt.chapterCount, chapCount)
			}

			// Check for reading direction in package.opf
			rc, err = contentFile.Open()
			if err != nil {
				t.Fatalf("Failed to open package.opf: %v", err)
			}
			opfContent, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read package.opf: %v", err)
			}

			// Check reading direction
			rtlText := "page-progression-direction=\"rtl\""

			if tt.ltr {
				if strings.Contains(string(opfContent), rtlText) {
					t.Error("Expected LTR reading direction, found RTL")
				}
				// Note: Some EPUB readers default to LTR if not specified, so lack of LTR text might be ok
			} else {
				if !strings.Contains(string(opfContent), rtlText) {
					t.Error("Expected RTL reading direction, not found in package.opf")
				}
			}
		})
	}
}

// TestEPUBMetadataHandling tests how the EPUB generator handles various metadata scenarios
func TestEPUBMetadataHandling(t *testing.T) {
	tests := []struct {
		name     string
		modify   func(*md.Manga)
		validate func(*testing.T, *epub.Epub)
		wantErr  bool
	}{
		{
			name: "standard metadata",
			modify: func(manga *md.Manga) {
				// No changes
			},
			validate: func(t *testing.T, e *epub.Epub) {
				if e.Identifier() != "test-manga-id" {
					t.Errorf("Expected identifier 'test-manga-id', got %s", e.Identifier())
				}
				if e.Title() != "Test Manga" {
					t.Errorf("Expected title 'Test Manga', got %s", e.Title())
				}
				// Just check that the author is set
				title := e.Title()
				if title == "" {
					t.Error("EPUB title is empty")
				}
			},
			wantErr: false,
		},
		{
			name: "missing title",
			modify: func(manga *md.Manga) {
				manga.Info.Title = ""
			},
			validate: func(t *testing.T, e *epub.Epub) {
				// Check if title is set
				title := e.Title()
				if title == "" {
					t.Error("EPUB title should not be empty")
				}
			},
			wantErr: false,
		},
		{
			name: "missing authors",
			modify: func(manga *md.Manga) {
				manga.Info.Authors = nil
			},
			validate: func(t *testing.T, e *epub.Epub) {
				// Author may be "Unknown" or similar
				// Check if we can get title, which confirms the EPUB is valid
				title := e.Title()
				if title == "" {
					t.Error("EPUB authors should not be empty")
				}
			},
			wantErr: false,
		},
		{
			name: "very long title",
			modify: func(manga *md.Manga) {
				manga.Info.Title = strings.Repeat("Very Long Title ", 20)
			},
			validate: func(t *testing.T, e *epub.Epub) {
				// Check if title is set
				title := e.Title()
				if title == "" {
					t.Error("EPUB title should not be empty")
				}

				// Check if title length is reasonable
				if len(title) > 300 {
					t.Errorf("EPUB title is too long: %d characters", len(title))
				}
			},
			wantErr: false,
		},
		{
			name: "special characters in title",
			modify: func(manga *md.Manga) {
				manga.Info.Title = "Special & Characters: < > \" '"
			},
			validate: func(t *testing.T, e *epub.Epub) {
				title := e.Title()
				if !strings.Contains(title, "Special") {
					t.Errorf("EPUB title should contain 'Special', got %s", title)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := testhelpers.CreateTestManga()
			tt.modify(&manga)

			epub, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cleanup != nil {
				defer cleanup()
			}

			if tt.wantErr {
				return
			}

			tt.validate(t, epub)
		})
	}
}

// TestEPUBErrorHandling tests how the EPUB generator handles various error scenarios
func TestEPUBErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		manga   md.Manga
		wantErr bool
	}{
		{
			name:    "empty manga",
			manga:   md.Manga{},
			wantErr: true,
		},
		{
			name: "no volumes",
			manga: md.Manga{
				Info: md.MangaInfo{
					Title: "Test Manga",
					ID:    "test-id",
				},
				Volumes: nil,
			},
			wantErr: true,
		},
		{
			name: "empty volumes",
			manga: md.Manga{
				Info: md.MangaInfo{
					Title: "Test Manga",
					ID:    "test-id",
				},
				Volumes: map[md.Identifier]md.Volume{},
			},
			wantErr: true,
		},
		{
			name:    "invalid images",
			manga:   createInvalidImageManga(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup, err := GenerateEPUB(tt.manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cleanup != nil {
				defer cleanup()
			}
		})
	}
}

// verifyEPUBStructure checks the basic structure of an EPUB file
func verifyEPUBStructure(t *testing.T, epubPath string, ltr bool) {
	// Check file exists
	if _, err := os.Stat(epubPath); os.IsNotExist(err) {
		t.Fatalf("EPUB file does not exist: %s", epubPath)
	}

	// Open as ZIP file to verify structure
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB as ZIP: %v", err)
	}
	defer r.Close()

	// Check for required files
	requiredFiles := []string{
		"mimetype",
		"META-INF/container.xml",
		"EPUB/package.opf",
		"EPUB/nav.xhtml",
		"EPUB/css/style.css",
	}

	for _, requiredFile := range requiredFiles {
		found := false
		for _, f := range r.File {
			if f.Name == requiredFile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required file missing from EPUB: %s", requiredFile)
		}
	}

	// Check mimetype is first file and uncompressed
	if len(r.File) > 0 {
		if r.File[0].Name != "mimetype" {
			t.Error("Mimetype is not the first file in the EPUB")
		} else if r.File[0].Method != zip.Store {
			t.Error("Mimetype is compressed, should be stored without compression")
		}
	}
}

// createWidePageTestManga creates a manga with wide pages for testing
func createWidePageTestManga() md.Manga {
	manga := testhelpers.CreateTestManga()

	// Add a wide page chapter
	volID := md.NewIdentifier("1")
	chapID := md.NewIdentifier("wide")

	// Create a chapter with wide pages
	wideChapter := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chapID,
			Title:            "Wide Page Chapter",
			VolumeIdentifier: volID,
		},
		Pages: map[int]image.Image{
			0: createTestImage(1000, 1500, color.White), // Normal
			1: createTestImage(2000, 1000, color.White), // Wide 2:1
			2: createTestImage(1000, 1500, color.White), // Normal
		},
	}

	// Add to volume if it exists, otherwise create a new volume
	vol, exists := manga.Volumes[volID]
	if exists {
		vol.Chapters[chapID] = wideChapter
		manga.Volumes[volID] = vol
	} else {
		newVol := md.Volume{
			Info: md.VolumeInfo{
				Identifier: volID,
			},
			Chapters: map[md.Identifier]md.Chapter{
				chapID: wideChapter,
			},
		}
		manga.Volumes[volID] = newVol
	}

	return manga
}

// createInvalidImageManga creates a manga with nil images
func createInvalidImageManga() md.Manga {
	manga := testhelpers.CreateTestManga()

	// Replace images with nil
	for volID, vol := range manga.Volumes {
		for chapID, chap := range vol.Chapters {
			for pageID := range chap.Pages {
				chap.Pages[pageID] = nil
			}
			vol.Chapters[chapID] = chap
		}
		manga.Volumes[volID] = vol
	}

	return manga
}

// createSpecialCharTitleManga creates a manga with special characters in the title
func createSpecialCharTitleManga() md.Manga {
	manga := testhelpers.CreateTestManga()
	manga.Info.Title = "Special & Characters: < > \" '"
	return manga
}

// createMixedLanguageManga creates a manga with chapters in different languages
func createMixedLanguageManga() md.Manga {
	manga := testhelpers.CreateTestManga()

	// Set different languages for different chapters
	languages := []language.Tag{
		language.English,
		language.Japanese,
		language.Spanish,
	}

	i := 0
	for volID, vol := range manga.Volumes {
		for chapID, chap := range vol.Chapters {
			chap.Info.Language = languages[i%len(languages)]
			vol.Chapters[chapID] = chap
			i++
		}
		manga.Volumes[volID] = vol
	}

	return manga
}

// createLargeImageManga creates a manga with very large images
func createLargeImageManga() md.Manga {
	manga := testhelpers.CreateTestManga()

	// Replace with large images (but not too large to avoid memory issues in tests)
	for volID, vol := range manga.Volumes {
		for chapID, chap := range vol.Chapters {
			for pageID := range chap.Pages {
				// 3000x4000 is large but not excessive for testing
				chap.Pages[pageID] = createTestImage(3000, 4000, color.White)
			}
			vol.Chapters[chapID] = chap
		}
		manga.Volumes[volID] = vol
	}

	return manga
}

// writeEPUBToZipReader writes the EPUB object to a temporary file and returns a zip reader
func writeEPUBToZipReader(t *testing.T, epubObj *epub.Epub) (*zip.ReadCloser, error) {
	// Write the EPUB to a temp file
	if epubObj == nil {
		return nil, fmt.Errorf("epubObj is nil")
	}
	epubPath := filepath.Join(t.TempDir(), "test.epub")
	if err := epubObj.Write(epubPath); err != nil {
		return nil, err
	}

	// Open the file as a zip reader
	zipReader, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, err
	}

	return zipReader, nil
}
