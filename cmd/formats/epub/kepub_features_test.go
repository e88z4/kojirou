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
	md "github.com/leotaku/kojirou/mangadex"
)

// TestKEPUBKoboSpecificFeatures tests Kobo-specific features in KEPUB format
func TestKEPUBKoboSpecificFeatures(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Generate a standard EPUB first
	manga := createTestManga()
	epubObj, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Verify Kobo-specific features
	testFeatures := []struct {
		name     string
		validate func(*testing.T, []byte)
	}{
		{
			name: "kobo spans in text",
			validate: func(t *testing.T, data []byte) {
				verifyKoboSpans(t, data)
			},
		},
		{
			name: "kobo fixed layout",
			validate: func(t *testing.T, data []byte) {
				verifyKoboFixedLayout(t, data)
			},
		},
		{
			name: "kobo namespaces",
			validate: func(t *testing.T, data []byte) {
				verifyKoboNamespaces(t, data)
			},
		},
		{
			name: "kobo content type",
			validate: func(t *testing.T, data []byte) {
				verifyKoboContentType(t, data)
			},
		},
	}

	for _, feature := range testFeatures {
		t.Run(feature.name, func(t *testing.T) {
			feature.validate(t, kepubData)
		})
	}
}

// TestKEPUBFormatStructure tests the structure of KEPUB format files
func TestKEPUBFormatStructure(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create test manga with different configurations
	testCases := []struct {
		name     string
		manga    md.Manga
		widepage kindle.WidepagePolicy
		autocrop bool
		ltr      bool
	}{
		{
			name:     "standard manga RTL",
			manga:    createTestManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      false,
		},
		{
			name:     "western comic LTR",
			manga:    createTestManga(), // Using same manga but with LTR flag
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
		},
		{
			name:     "split wide pages",
			manga:    createWidePageTestManga(),
			widepage: kindle.WidepagePolicySplit,
			autocrop: false,
			ltr:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate EPUB
			epubObj, cleanup, err := GenerateEPUB(tc.manga, tc.widepage, tc.autocrop, tc.ltr)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Convert to KEPUB
			kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Verify KEPUB structure
			verifyKEPUBStructure(t, kepubData)

			// Check reading direction
			verifyKEPUBReadingDirection(t, kepubData, tc.ltr)

			// Check wide page handling if applicable
			if tc.widepage == kindle.WidepagePolicySplit {
				verifyKEPUBWidePageHandling(t, kepubData)
			}
		})
	}
}

// TestKEPUBExtensionHandling tests proper extension handling
func TestKEPUBExtensionHandling(t *testing.T) {
	// Skip test until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Check that the extension constant is defined correctly
	if KEPUBExtension != ".kepub.epub" {
		t.Errorf("KEPUBExtension incorrect, got %s, expected .kepub.epub", KEPUBExtension)
	}

	// Create a test manga and convert to KEPUB
	manga := createTestManga()
	epubObj, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Convert to KEPUB and verify content
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Create a simple EPUB manually
	e := epub.NewEpub("Test EPUB")
	e.SetAuthor("Test Author")
	_, err = e.AddSection("<h1>Test</h1><p>Content</p>", "Chapter", "ch1", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Convert this simple EPUB to KEPUB
	simpleKepubData, err := kepubconv.ConvertToKEPUB(e)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed for simple EPUB: %v", err)
	}

	// Both outputs should have kepub-specific features
	verifyKoboEnhancements(t, kepubData)
	verifyKoboEnhancements(t, simpleKepubData)
}

// Helper functions for KEPUB validation

// verifyKoboSpans checks for Kobo span elements in HTML
func verifyKEPUBStructure(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Required files
	requiredFiles := map[string]bool{
		"mimetype":               false,
		"META-INF/container.xml": false,
	}

	hasOPF := false
	hasNCX := false
	hasHTML := false

	for _, f := range r.File {
		if _, ok := requiredFiles[f.Name]; ok {
			requiredFiles[f.Name] = true
		}

		if strings.HasSuffix(f.Name, ".opf") {
			hasOPF = true
		}

		if strings.HasSuffix(f.Name, ".ncx") {
			hasNCX = true
		}

		if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			hasHTML = true
		}
	}

	// Check required files
	for file, found := range requiredFiles {
		if !found {
			t.Errorf("Required file missing: %s", file)
		}
	}

	// Check other essential components
	if !hasOPF {
		t.Error("No OPF file found in KEPUB")
	}

	if !hasNCX {
		t.Error("No NCX file found in KEPUB")
	}

	if !hasHTML {
		t.Error("No HTML/XHTML files found in KEPUB")
	}

	// Check mimetype is uncompressed and first in the archive
	if len(r.File) > 0 {
		if r.File[0].Name != "mimetype" {
			t.Error("mimetype file must be first in the archive")
		} else if r.File[0].Method != zip.Store {
			t.Error("mimetype file must be stored without compression")
		}
	}
}

// verifyKEPUBReadingDirection checks reading direction in KEPUB
func verifyKEPUBReadingDirection(t *testing.T, data []byte, ltr bool) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	expectedDirection := "rtl"
	if ltr {
		expectedDirection = "ltr"
	}

	directionFound := false
	for _, f := range r.File {
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

			if bytes.Contains(content, []byte("page-progression-direction=\""+expectedDirection+"\"")) {
				directionFound = true
				break
			}
		}
	}

	if !directionFound {
		t.Errorf("Expected reading direction (%s) not found in KEPUB", expectedDirection)
	}
}

// verifyKEPUBWidePageHandling checks for proper handling of wide pages
func verifyKEPUBWidePageHandling(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Look for split page references in HTML
	splitPagesFound := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Look for indicators of split pages
			if bytes.Contains(content, []byte("part-left")) ||
				bytes.Contains(content, []byte("part-right")) ||
				bytes.Contains(content, []byte("wide-page-")) {
				splitPagesFound = true
				break
			}
		}
	}

	if !splitPagesFound {
		t.Error("No evidence of wide page handling found in KEPUB")
	}
}

// verifyKoboEnhancements checks KEPUB has Kobo-specific enhancements
