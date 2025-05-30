package epub

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

// TestEPUBToKEPUBDependencies tests the dependencies between EPUB and KEPUB formats
func TestEPUBToKEPUBDependencies(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a test EPUB with standard content
	epubObj := epub.NewEpub("Test Cross-Format")
	epubObj.SetAuthor("Test Author")

	// Add a standard section
	_, err := epubObj.AddSection("<h1>Test Chapter</h1><p>This is a standard test.</p>", "Chapter 1", "ch1", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Test KEPUB conversion
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("Basic ConvertToKEPUB() failed: %v", err)
	}

	// Verify KEPUB data is valid
	if len(kepubData) == 0 {
		t.Error("KEPUB data is empty")
	}

	// Ensure KEPUB has proper Kobo-specific changes
	verifyKoboEnhancements(t, kepubData)
}

// TestEPUBToKEPUBPreservesMetadata tests that metadata is preserved when converting from EPUB to KEPUB
func TestEPUBToKEPUBPreservesMetadata(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create EPUB with rich metadata
	e := epub.NewEpub("Metadata Test")
	e.SetAuthor("Test Author")
	e.SetDescription("This is a test description that should be preserved")
	e.SetLang("en")
	e.SetTitle("Metadata Test - Custom Title")

	// Add a section
	_, err := e.AddSection("<h1>Test Chapter</h1><p>This is a test.</p>", "Chapter 1", "ch1", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(e)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Verify metadata is preserved in OPF
	verifyMetadataPreservation(t, kepubData, "Metadata Test - Custom Title", "Test Author", "en")
}

// TestEPUBToKEPUBWithManga tests the conversion with actual manga data
func TestEPUBToKEPUBWithManga(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a test manga
	manga := createTestManga()

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Verify KEPUB structure and manga-specific elements
	verifyMangaKEPUB(t, kepubData)
}

// verifyKoboEnhancements checks that the KEPUB has Kobo-specific enhancements
func verifyKoboEnhancements(t *testing.T, data []byte) {
	// Read the KEPUB as ZIP
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check for Kobo-specific features
	hasKoboHTML := false
	hasKoboOPF := false

	for _, f := range r.File {
		// Check for Kobo spans in HTML files
		if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			rc, err := f.Open()
			if err != nil {
				t.Errorf("Failed to open HTML file: %v", err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read HTML: %v", err)
				continue
			}

			if bytes.Contains(content, []byte("kobo")) {
				hasKoboHTML = true
			}
		}

		// Check for Kobo metadata in OPF
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Errorf("Failed to open OPF file: %v", err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read OPF: %v", err)
				continue
			}

			if bytes.Contains(content, []byte("kobo:")) {
				hasKoboOPF = true
			}
		}
	}

	if !hasKoboHTML {
		t.Error("KEPUB doesn't contain Kobo-specific HTML enhancements")
	}

	if !hasKoboOPF {
		t.Error("KEPUB doesn't contain Kobo-specific OPF metadata")
	}
}

// verifyMetadataPreservation checks that metadata is preserved in the KEPUB
func verifyMetadataPreservation(t *testing.T, data []byte, title, author, language string) {
	// Read the KEPUB as ZIP
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and check the OPF file
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Errorf("Failed to open OPF file: %v", err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read OPF: %v", err)
				continue
			}

			// Check that metadata is preserved
			contentStr := string(content)
			if !strings.Contains(contentStr, title) {
				t.Errorf("KEPUB OPF doesn't contain title: %s", title)
			}

			if !strings.Contains(contentStr, author) {
				t.Errorf("KEPUB OPF doesn't contain author: %s", author)
			}

			if !strings.Contains(contentStr, language) {
				t.Errorf("KEPUB OPF doesn't contain language: %s", language)
			}

			return
		}
	}

	t.Error("No OPF file found in KEPUB")
}

// verifyMangaKEPUB checks that the KEPUB contains manga-specific elements
func verifyMangaKEPUB(t *testing.T, data []byte) {
	// Read the KEPUB as ZIP
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check for manga-specific features
	hasImages := false
	hasChapters := false
	hasRTL := false

	for _, f := range r.File {
		// Check for images
		if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
			hasImages = true
		}

		// Check for chapters
		if strings.Contains(f.Name, "chapter") && (strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml")) {
			hasChapters = true
		}

		// Check for RTL direction in OPF
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			if bytes.Contains(content, []byte("page-progression-direction=\"rtl\"")) {
				hasRTL = true
			}
		}
	}

	if !hasImages {
		t.Error("Manga KEPUB doesn't contain image files")
	}

	if !hasChapters {
		t.Error("Manga KEPUB doesn't contain chapter files")
	}

	if !hasRTL {
		t.Error("Manga KEPUB doesn't specify right-to-left reading direction")
	}
}
