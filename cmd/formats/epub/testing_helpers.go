package epub

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/draw"
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

	// Read the file back
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, err
	}

	// Return a zip reader for inspection
	return zip.NewReader(bytes.NewReader(data), int64(len(data)))
}

// createTestImage creates a simple test image of the specified size and color
func createTestImageColored(width, height int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
	return img
}
