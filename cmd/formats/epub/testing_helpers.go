package epub

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmaupin/go-epub"
)

// writeEPUB writes an EPUB to a temporary file and returns a zip.Reader for inspection
func writeEPUB(t *testing.T, e *epub.Epub) (*zip.Reader, error) {
	t.Helper()

	// Write EPUB to temporary file
	tmpFile := filepath.Join(t.TempDir(), "test.epub")
	if err := e.Write(tmpFile); err != nil {
		return nil, err
	}

	// Patch the OPF manifest to ensure nav.xhtml is marked as navigation
	if err := PatchEPUBNavManifest(tmpFile); err != nil {
		return nil, err
	}

	// Read the file back
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, err
	}

	// Return a zip reader for inspection
	return zip.NewReader(bytes.NewReader(data), int64(len(data)))
}
