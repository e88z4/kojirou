package epub

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	"github.com/leotaku/kojirou/cmd/formats/testhelpers"
)

// TestKEPUBSimple tests the basic functionality of the KEPUB converter
func TestKEPUBSimple(t *testing.T) {
	// Create a test manga
	manga := testhelpers.CreateTestManga()

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

	// Verify KEPUB data is not empty
	if len(kepubData) == 0 {
		t.Error("KEPUB data is empty")
	}

	// Verify KEPUB can be opened as a ZIP file
	reader, err := zip.NewReader(bytes.NewReader(kepubData), int64(len(kepubData)))
	if err != nil {
		t.Fatalf("Failed to open KEPUB as ZIP: %v", err)
	}

	// Verify KEPUB contains basic EPUB files
	requiredFiles := []string{
		"mimetype",
		"META-INF/container.xml",
		"EPUB/package.opf",
	}

	for _, filename := range requiredFiles {
		found := false
		for _, f := range reader.File {
			if f.Name == filename {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required file missing in KEPUB: %s", filename)
		}
	}

	// Verify at least one file contains Kobo namespace
	hasKoboNamespace := false
	for _, f := range reader.File {
		if !strings.HasSuffix(f.Name, ".xhtml") && !strings.HasSuffix(f.Name, ".opf") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Errorf("Failed to open file %s: %v", f.Name, err)
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Errorf("Failed to read file %s: %v", f.Name, err)
			continue
		}

		if bytes.Contains(content, []byte("xmlns:kobo")) {
			hasKoboNamespace = true
			break
		}
	}

	if !hasKoboNamespace {
		t.Error("No file contains Kobo namespace")
	}
}
