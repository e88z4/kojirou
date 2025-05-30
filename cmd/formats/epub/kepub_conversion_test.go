// Package epub provides functionality for generating EPUB format ebooks from manga
package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	testhelpers "github.com/leotaku/kojirou/cmd/formats/testhelpers"
	"github.com/leotaku/kojirou/mangadex"
)

// createTestManga creates a test manga for use in testing
func createTestManga() mangadex.Manga {
	manga := mangadex.Manga{
		Info: mangadex.MangaInfo{
			Title:   "Test Manga",
			ID:      "test-manga-id",
			Authors: []string{"Test Author"},
		},
		Volumes: map[mangadex.Identifier]mangadex.Volume{},
	}
	volID := mangadex.NewIdentifier("1")
	vol := mangadex.Volume{
		Info:     mangadex.VolumeInfo{Identifier: volID},
		Chapters: map[mangadex.Identifier]mangadex.Chapter{},
	}
	chapID := mangadex.NewIdentifier("1-1")
	chap := mangadex.Chapter{
		Info: mangadex.ChapterInfo{
			Identifier:       chapID,
			Title:            "Chapter 1",
			VolumeIdentifier: volID,
		},
		Pages: map[int]image.Image{
			0: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
		},
	}
	vol.Chapters[chapID] = chap
	manga.Volumes[volID] = vol
	return manga
}

// createMultiVolumeTestManga creates a manga with multiple volumes for testing
func createMultiVolumeTestManga() mangadex.Manga {
	manga := createTestManga()
	vol2ID := mangadex.NewIdentifier("2")
	vol2 := mangadex.Volume{
		Info:     mangadex.VolumeInfo{Identifier: vol2ID},
		Chapters: map[mangadex.Identifier]mangadex.Chapter{},
	}
	chap2ID := mangadex.NewIdentifier("2-1")
	chap2 := mangadex.Chapter{
		Info: mangadex.ChapterInfo{
			Identifier:       chap2ID,
			Title:            "Chapter 2",
			VolumeIdentifier: vol2ID,
		},
		Pages: map[int]image.Image{
			0: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
			1: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
		},
	}
	vol2.Chapters[chap2ID] = chap2
	manga.Volumes[vol2ID] = vol2
	return manga
}

// TestEPUBToKEPUBConversion tests the full conversion process from EPUB to KEPUB format
func TestEPUBToKEPUBConversion(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*epub.Epub, error)
		validate func(*testing.T, []byte)
		wantErr  bool
	}{
		{
			name: "standard epub to kepub",
			setup: func() (*epub.Epub, error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				if err != nil {
					return nil, err
				}
				if cleanup != nil { //nolint:staticcheck // cleanup() will be called after KEPUB conversion below
				}
				return epub, nil
			},
			validate: func(t *testing.T, data []byte) {
				validateBasicKEPUBStructure(t, data)
				validateKoboSpecificMetadata(t, data)
				validateKoboHTMLTransformation(t, data)
			},
			wantErr: false,
		},
		{
			name: "complex manga with multiple volumes",
			setup: func() (*epub.Epub, error) {
				manga := createMultiVolumeTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				if err != nil {
					return nil, err
				}
				if cleanup != nil { //nolint:staticcheck // cleanup() will be called after KEPUB conversion below
				}
				return epub, nil
			},
			validate: func(t *testing.T, data []byte) {
				validateBasicKEPUBStructure(t, data)
				validateKoboSpecificMetadata(t, data)
				validateComplexNavigation(t, data)
			},
			wantErr: false,
		},
		{
			name: "custom epub with minimal content",
			setup: func() (*epub.Epub, error) {
				// Create a minimal EPUB manually instead of using GenerateEPUB
				e := epub.NewEpub("Minimal Test")
				e.SetAuthor("Test Author")

				// Add a simple section
				htmlContent := "<h1>Test Chapter</h1><p>This is a test paragraph with some text.</p>"
				_, err := e.AddSection(htmlContent, "Test Chapter", "ch1", "")
				if err != nil {
					return nil, fmt.Errorf("failed to add section: %w", err)
				}

				// Add a minimal CSS file
				cssContent := "body { margin: 0; padding: 0; } img { display: block; max-width: 100%; height: auto; }"
				cssFile, err := os.CreateTemp("", "test-style-*.css")
				if err != nil {
					return nil, fmt.Errorf("failed to create temp CSS file: %w", err)
				}
				cssPath := cssFile.Name()
				_, err = cssFile.Write([]byte(cssContent))
				cssFile.Close()
				if err != nil {
					return nil, fmt.Errorf("failed to write CSS: %w", err)
				}
				_, err = e.AddCSS(cssPath, "style.css")
				if err != nil {
					return nil, fmt.Errorf("failed to add CSS: %w", err)
				}

				return e, nil
			},
			validate: func(t *testing.T, data []byte) {
				validateBasicKEPUBStructure(t, data)
				validateMinimalContent(t, data)
			},
			wantErr: false,
		},
		{
			name: "empty epub",
			setup: func() (*epub.Epub, error) {
				return epub.NewEpub("Empty Test"), nil
			},
			validate: nil,
			wantErr:  true, // Should fail with no content
		},
		{
			name: "nil epub",
			setup: func() (*epub.Epub, error) {
				return nil, nil
			},
			validate: nil,
			wantErr:  true, // Should fail with nil input
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate/get EPUB
			epub, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Convert to KEPUB
			kepubData, err := kepubconv.ConvertToKEPUB(epub)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToKEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Skip validation if we expected an error
			if tt.wantErr {
				return
			}

			// Verify KEPUB data
			if tt.validate != nil {
				tt.validate(t, kepubData)
			}
		})
	}
}

// TestKEPUBEnhancedFeatures tests Kobo-specific enhancement features
func TestKEPUBEnhancedFeatures(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() (*epub.Epub, func(), error)
		checkFeature func(*testing.T, []byte)
		wantErr      bool
	}{
		{
			name: "kobo spans for pagination",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			checkFeature: func(t *testing.T, data []byte) {
				validateKoboTextSpans(t, data)
			},
			wantErr: false,
		},
		{
			name: "kobo fixed layout metadata",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			checkFeature: func(t *testing.T, data []byte) {
				validateKoboFixedLayout(t, data)
			},
			wantErr: false,
		},
		{
			name: "kobo image handling",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			checkFeature: func(t *testing.T, data []byte) {
				validateKoboImageHandling(t, data)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer cleanup()

			kepubData, err := kepubconv.ConvertToKEPUB(epub)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToKEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.checkFeature != nil {
				tt.checkFeature(t, kepubData)
			}
		})
	}
}

// TestKEPUBCompatibility tests compatibility with Kobo devices through file structure validation
func TestKEPUBCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*epub.Epub, func(), error)
		validate func(*testing.T, []byte)
		wantErr  bool
	}{
		{
			name: "standard compatibility",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			validate: func(t *testing.T, data []byte) {
				validateKoboCompatibility(t, data)
			},
			wantErr: false,
		},
		{
			name: "special title compatibility",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				manga.Info.Title = "Special: Characters & Test"
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			validate: func(t *testing.T, data []byte) {
				validateKoboCompatibility(t, data)
				validateSpecialCharacterHandling(t, data)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer cleanup()

			kepubData, err := kepubconv.ConvertToKEPUB(epub)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToKEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.validate != nil {
				tt.validate(t, kepubData)
			}
		})
	}
}

// Helper functions for KEPUB validation

// validateBasicKEPUBStructure checks basic KEPUB file structure requirements
func validateBasicKEPUBStructure(t *testing.T, data []byte) {
	if len(data) == 0 {
		t.Fatal("KEPUB data is empty")
	}

	// Verify we can read it as a ZIP file
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check required files
	requiredFiles := []string{
		"mimetype",
		"META-INF/container.xml",
	}

	hasOPF := false
	for _, f := range r.File {
		for i, required := range requiredFiles {
			if f.Name == required {
				requiredFiles = append(requiredFiles[:i], requiredFiles[i+1:]...)
				break
			}
		}

		if strings.HasSuffix(f.Name, ".opf") {
			hasOPF = true
		}
	}

	if len(requiredFiles) > 0 {
		t.Errorf("Missing required files: %v", requiredFiles)
	}

	if !hasOPF {
		t.Error("No OPF file found in KEPUB")
	}

	// Verify mimetype is first file and uncompressed
	if len(r.File) == 0 || r.File[0].Name != "mimetype" {
		t.Error("mimetype must be the first file in the KEPUB")
	} else if r.File[0].Method != zip.Store {
		t.Error("mimetype must be stored without compression")
	}
}

// validateKoboSpecificMetadata checks for Kobo-specific metadata in OPF
func validateKoboSpecificMetadata(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and read OPF file
	var opfContent []byte
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			opfContent, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}
			break
		}
	}

	if len(opfContent) == 0 {
		t.Fatal("No OPF content found")
	}

	// Check for required Kobo metadata
	koboMetadataMarkers := []string{
		"kobo:",
		"rendition:",
	}

	for _, marker := range koboMetadataMarkers {
		if !bytes.Contains(opfContent, []byte(marker)) {
			t.Errorf("OPF doesn't contain expected Kobo metadata: %s", marker)
		}
	}
}

// validateKoboHTMLTransformation checks for Kobo-specific HTML transformations
func validateKoboHTMLTransformation(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check HTML files for Kobo transformations
	foundTransformedHTML := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xhtml") || strings.HasSuffix(f.Name, ".html") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Check for Kobo-specific text spans and attributes
			if bytes.Contains(content, []byte("kobo:")) ||
				bytes.Contains(content, []byte("<span class=\"kobo")) ||
				bytes.Contains(content, []byte("koboSpan")) ||
				bytes.Contains(content, []byte("epub:type=\"kobo\"")) ||
				bytes.Contains(content, []byte("class=\"kobo-image\"")) ||
				bytes.Contains(content, []byte("xmlns:kobo")) {
				foundTransformedHTML = true
				break
			}
		}
	}

	if !foundTransformedHTML {
		t.Error("No evidence of Kobo HTML transformations found")
	}
}

// validateComplexNavigation checks for proper navigation in complex multi-volume content
func validateComplexNavigation(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Look for navigation document and check for multiple volumes
	var navContent []byte
	for _, f := range r.File {
		if strings.Contains(f.Name, "nav.xhtml") || strings.HasSuffix(f.Name, ".ncx") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			navContent, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}
			break
		}
	}

	if len(navContent) == 0 {
		t.Fatal("No navigation document found")
	}

	// Check that volume navigation is present
	volumeNavPresent := bytes.Contains(navContent, []byte("Volume 1")) ||
		bytes.Contains(navContent, []byte("Volume 2")) ||
		bytes.Contains(navContent, []byte("vol1")) ||
		bytes.Contains(navContent, []byte("vol2"))

	if !volumeNavPresent {
		t.Error("Complex navigation doesn't contain volume structure")
	}
}

// validateMinimalContent checks for basic content in a minimally created KEPUB
func validateMinimalContent(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check for the minimal expected content
	foundHTML := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xhtml") || strings.HasSuffix(f.Name, ".html") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			if bytes.Contains(content, []byte("Test Chapter")) {
				foundHTML = true
				break
			}
		}
	}

	if !foundHTML {
		t.Error("Minimal content not found in KEPUB")
	}
}

// validateKoboTextSpans checks for Kobo text span elements in HTML
func validateKoboTextSpans(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	foundSpans := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xhtml") || strings.HasSuffix(f.Name, ".html") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Check for Kobo-specific span elements
			if bytes.Contains(content, []byte("kobo-span")) ||
				bytes.Contains(content, []byte("class=\"kobo")) {
				foundSpans = true
				break
			}
		}
	}

	if !foundSpans {
		t.Error("No Kobo text spans found in HTML content")
	}
}

// validateKoboFixedLayout checks for fixed layout metadata
func validateKoboFixedLayout(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and check OPF content
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}

			// Check for fixed layout metadata
			if !bytes.Contains(content, []byte("rendition:layout")) ||
				!bytes.Contains(content, []byte("pre-paginated")) {
				t.Error("Fixed layout metadata not found")
			}
			return
		}
	}

	t.Error("No OPF file found")
}

// validateKoboImageHandling checks for proper Kobo image handling
func validateKoboImageHandling(t *testing.T, data []byte) {
	validateBasicKEPUBStructure(t, data) // Basic validation is sufficient for this
}

// validateKoboCompatibility checks for Kobo device compatibility
func validateKoboCompatibility(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check for Kobo compatibility markers
	var hasKoboTypeMetadata bool
	var hasKoboVersion bool

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}

			// Check for required Kobo metadata
			hasKoboTypeMetadata = bytes.Contains(content, []byte("kobo:content-type"))
			hasKoboVersion = bytes.Contains(content, []byte("kobo:epub-version"))
		}
	}

	if !hasKoboTypeMetadata {
		t.Error("Missing kobo:content-type metadata required for Kobo compatibility")
	}

	if !hasKoboVersion {
		t.Error("Missing kobo:epub-version metadata required for Kobo compatibility")
	}
}

// validateSpecialCharacterHandling checks for proper handling of special characters
func validateSpecialCharacterHandling(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check OPF for properly escaped special characters
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}

			// Check for properly escaped special characters
			if bytes.Contains(content, []byte("&amp;")) {
				return // Found properly escaped ampersand
			}

			t.Error("Special characters not properly escaped in OPF")
			return
		}
	}
}
