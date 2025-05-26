package epub

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

// ... existing test code ...

// Helper functions for KEPUB validation

func verifyKEPUBStructure(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

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

	for file, found := range requiredFiles {
		if !found {
			t.Errorf("Required file missing: %s", file)
		}
	}
	if !hasOPF {
		t.Error("No OPF file found in KEPUB")
	}
	if !hasNCX {
		t.Error("No NCX file found in KEPUB")
	}
	if !hasHTML {
		t.Error("No HTML/XHTML files found in KEPUB")
	}
	if len(r.File) > 0 {
		if r.File[0].Name != "mimetype" {
			t.Error("mimetype file must be first in the archive")
		} else if r.File[0].Method != zip.Store {
			t.Error("mimetype file must be stored without compression")
		}
	}
}

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

func verifyKEPUBWidePageHandling(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}
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
