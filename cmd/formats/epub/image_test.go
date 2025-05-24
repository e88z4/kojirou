package epub

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
)

func TestEPUBImageProcessingDetailed(t *testing.T) {
	tests := []struct {
		name      string
		img       image.Image
		wantErr   bool
		errString string
	}{
		{
			name:    "valid image",
			img:     createTestImage(800, 1200, color.White),
			wantErr: false,
		},
		{
			name:      "nil image",
			img:       nil,
			wantErr:   true,
			errString: "nil image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new EPUB for each test
			e := epub.NewEpub("Test")

			// Set up a temp file for the image
			if tt.img != nil {
				tmpFile := filepath.Join(t.TempDir(), "test.jpg")
				f, err := os.Create(tmpFile)
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				defer f.Close()

				if err := jpeg.Encode(f, tt.img, &jpeg.Options{Quality: 85}); err != nil {
					t.Fatalf("failed to encode image: %v", err)
				}

				_, err = e.AddImage(tmpFile, "")
				if (err != nil) != tt.wantErr {
					t.Errorf("AddImage() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error containing %q, got %v", tt.errString, err)
				}
			}

			// Try to write the EPUB
			epubPath := filepath.Join(t.TempDir(), "test.epub")
			if err := e.Write(epubPath); err != nil {
				if !tt.wantErr {
					t.Errorf("Write() failed: %v", err)
				}
			}
		})
	}
}
