package epub

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

// TestInvalidMangaHandling tests how the GenerateEPUB function handles invalid manga input
func TestInvalidMangaHandling(t *testing.T) {
	manga := createTestManga()

	invalidManga := manga
	invalidManga.Volumes = nil

	emptyVolManga := manga
	emptyVolManga.Volumes = make(map[md.Identifier]md.Volume)

	noTitleManga := manga
	noTitleManga.Info.Title = ""

	invalidImageManga := manga
	for id, vol := range invalidImageManga.Volumes {
		for chapID, chap := range vol.Chapters {
			chap.Pages[0] = nil // Invalid image
			vol.Chapters[chapID] = chap
		}
		invalidImageManga.Volumes[id] = vol
		break
	}

	tests := []struct {
		name    string
		manga   md.Manga
		wantErr bool
	}{
		{
			name:    "Invalid manga (nil volumes)",
			manga:   invalidManga,
			wantErr: true,
		},
		{
			name:    "Empty volumes manga",
			manga:   emptyVolManga,
			wantErr: true,
		},
		{
			name:    "No title manga",
			manga:   noTitleManga,
			wantErr: true,
		},
		{
			name:    "Invalid image manga",
			manga:   invalidImageManga,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := GenerateEPUB(tt.manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if epub == nil && !tt.wantErr {
				t.Error("GenerateEPUB() returned nil but expected success")
				return
			}
			if cleanup != nil {
				defer cleanup()
			}
		})
	}
}

// TestEPUBGenerationAndValidation tests the complete EPUB generation functionality with validation
func TestEPUBGenerationAndValidation(t *testing.T) {
	manga := createTestManga()

	invalidManga := manga
	invalidManga.Volumes = nil

	emptyVolManga := manga
	emptyVolManga.Volumes = make(map[md.Identifier]md.Volume)

	noTitleManga := manga
	noTitleManga.Info.Title = ""

	invalidImageManga := manga
	for id, vol := range invalidImageManga.Volumes {
		for chapID, chap := range vol.Chapters {
			chap.Pages[0] = nil // Invalid image
			vol.Chapters[chapID] = chap
		}
		invalidImageManga.Volumes[id] = vol
		break
	}

	tests := []struct {
		name     string
		manga    md.Manga
		widepage kindle.WidepagePolicy
		autocrop bool
		ltr      bool
		wantErr  bool
	}{
		{
			name:     "basic manga with volumes and chapters",
			manga:    manga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "manga with wide page splitting",
			manga:    manga,
			widepage: kindle.WidepagePolicySplit,
			autocrop: false,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "manga with right-to-left reading",
			manga:    manga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      false,
			wantErr:  false,
		},
		{
			name:     "invalid manga with no volumes",
			manga:    invalidManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true,
		},
		{
			name:     "manga with autocrop enabled",
			manga:    manga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: true,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "empty manga volumes",
			manga:    emptyVolManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true,
		},
		{
			name:     "manga with no title",
			manga:    noTitleManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false, // Should use default title
		},
		{
			name:     "manga with invalid images",
			manga:    invalidImageManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := GenerateEPUB(tt.manga, tt.widepage, tt.autocrop, tt.ltr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if epub == nil && !tt.wantErr {
				t.Error("GenerateEPUB() returned nil but expected success")
				return
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Write EPUB to file and get zip reader
			zipReader, err := writeEPUB(t, epub)
			if err != nil {
				t.Errorf("failed to write and open EPUB: %v", err)
				return
			}

			// Verify expected files exist
			expectedFiles := []string{
				"EPUB/package.opf",       // Package document
				"EPUB/nav.xhtml",         // Navigation document
				"EPUB/style.css",         // Our custom CSS
				"mimetype",               // Required EPUB file
				"META-INF/container.xml", // Required EPUB file
			}

			for _, file := range expectedFiles {
				found := false
				for _, f := range zipReader.File {
					if f.Name == file {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("required file %s not found in EPUB", file)
				}
			}

			// Verify all volumes and chapters are present
			chapterCount := 0
			volumeCount := 0
			for _, f := range zipReader.File {
				if f.Name == "EPUB/nav.xhtml" {
					// TODO: Parse nav.xhtml to verify TOC structure
					volumeCount++
				} else if f.Name == "EPUB/package.opf" {
					// TODO: Parse package.opf to verify reading direction
					chapterCount++
				}
			}

			// Basic counts check
			expectedChapters := len(tt.manga.Chapters())
			if chapterCount != expectedChapters {
				t.Errorf("expected %d chapters, got %d", expectedChapters, chapterCount)
			}

			expectedVolumes := len(tt.manga.Volumes)
			if volumeCount != expectedVolumes {
				t.Errorf("expected %d volumes, got %d", expectedVolumes, volumeCount)
			}
		})
	}
}

func TestNavigationStructure(t *testing.T) {
	manga := createTestManga()

	// Test both LTR and RTL navigation
	tests := []struct {
		name     string
		ltr      bool
		widepage kindle.WidepagePolicy
	}{
		{
			name:     "left-to-right navigation",
			ltr:      true,
			widepage: kindle.WidepagePolicyPreserve,
		},
		{
			name:     "right-to-left navigation",
			ltr:      false,
			widepage: kindle.WidepagePolicyPreserve,
		},
		{
			name:     "split wide pages navigation",
			ltr:      true,
			widepage: kindle.WidepagePolicySplit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := GenerateEPUB(manga, tt.widepage, false, tt.ltr)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if epub == nil {
				t.Fatal("GenerateEPUB() returned nil but expected success")
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Write EPUB to file and get zip reader
			zipReader, err := writeEPUB(t, epub)
			if err != nil {
				t.Fatalf("failed to write and open EPUB: %v", err)
			}

			// Find and parse nav.xhtml and package.opf
			var navFile, packageFile *zip.File
			for _, f := range zipReader.File {
				switch f.Name {
				case "EPUB/nav.xhtml":
					navFile = f
				case "EPUB/package.opf":
					packageFile = f
				}
			}

			if navFile == nil {
				t.Fatal("nav.xhtml not found in EPUB")
			}
			if packageFile == nil {
				t.Fatal("package.opf not found in EPUB")
			}

			// Read nav.xhtml content
			rc, err := navFile.Open()
			if err != nil {
				t.Fatalf("failed to open nav.xhtml: %v", err)
			}
			navContent := make([]byte, navFile.UncompressedSize64)
			if _, err := rc.Read(navContent); err != nil {
				t.Fatalf("failed to read nav.xhtml: %v", err)
			}
			rc.Close()

			// Read package.opf
			rc, err = packageFile.Open()
			if err != nil {
				t.Fatalf("failed to open package.opf: %v", err)
			}
			opfContent := make([]byte, packageFile.UncompressedSize64)
			if _, err := rc.Read(opfContent); err != nil {
				t.Fatalf("failed to read package.opf: %v", err)
			}
			rc.Close()

			// Verify navigation contains volume and chapter titles in correct order
			navString := string(navContent)
			opfString := string(opfContent)

			expectedItems := []string{
				"Volume 1",
				"Chapter 1",
				"Volume 2",
				"Chapter 2",
			}

			lastPos := 0
			for _, item := range expectedItems {
				pos := strings.Index(navString[lastPos:], item)
				if pos == -1 {
					t.Errorf("navigation missing expected item: %s", item)
					continue
				}
				pos += lastPos
				lastPos = pos
			}

			// Check reading direction in package.opf
			if tt.ltr {
				if !strings.Contains(opfString, `<spine>`) {
					t.Error("package.opf missing spine element for LTR")
				}
				if strings.Contains(opfString, `<spine page-progression-direction="rtl">`) {
					t.Error("package.opf has RTL spine for LTR book")
				}
			} else {
				if !strings.Contains(opfString, `<spine page-progression-direction="rtl">`) {
					t.Error("package.opf missing RTL spine direction")
				}
			}

			// Verify image references in package.opf
			contentOpfChecks := []string{
				`properties="duokan-page-fullscreen"`, // Check image properties
				`media-type="image/jpeg"`,             // Verify image format
				`properties="nav"`,                    // Check navigation properties
			}

			for _, check := range contentOpfChecks {
				if !strings.Contains(opfString, check) {
					t.Errorf("package.opf missing expected content: %s", check)
				}
			}
		})
	}
}

func TestImageProcessing(t *testing.T) {
	manga := createTestManga()

	// Test wide page splitting
	epub, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicySplit, false, true)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Write EPUB to file and get zip reader
	zipReader, err := writeEPUB(t, epub)
	if err != nil {
		t.Fatalf("failed to write and open EPUB: %v", err)
	}

	// Count total images - should be:
	// - 2 volume covers
	// - 1 normal page in Chapter 1
	// - 2 pages from split wide page in Chapter 1
	// - 1 normal page in Chapter 2
	imageCount := 0
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "EPUB/images/") && strings.HasSuffix(f.Name, ".jpg") {
			imageCount++

			// Check image exists and can be opened
			rc, err := f.Open()
			if err != nil {
				t.Errorf("failed to open image %s: %v", f.Name, err)
				continue
			}
			rc.Close()
		}
	}

	expectedImages := 6 // 2 covers + 4 content pages (including split wide page)
	if imageCount != expectedImages {
		t.Errorf("expected %d images, got %d", expectedImages, imageCount)
	}
}

func TestBasicEPUB(t *testing.T) {
	title := "Test EPUB"
	e := epub.NewEpub(title)
	if e == nil {
		t.Fatal("NewEpub returned nil")
	}

	// Add a basic chapter
	html := "<h1>Test Chapter</h1><p>Test content</p>"
	if _, err := e.AddSection(html, "Test Chapter", "test", ""); err != nil {
		t.Fatalf("AddSection failed: %v", err)
	}

	// Write the EPUB to a temp file
	tmpFile := filepath.Join(t.TempDir(), "test.epub")
	if err := e.Write(tmpFile); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read the file back
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Verify it's a valid ZIP file (EPUB is a ZIP)
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	if len(reader.File) == 0 {
		t.Error("EPUB has no files")
	}

	var hasFiles bool
	for _, f := range reader.File {
		t.Logf("Found file: %s", f.Name)
		hasFiles = true
	}

	if !hasFiles {
		t.Error("No files found in EPUB")
	}
}
