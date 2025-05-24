package epub

import (
	"image"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

func TestImageResolutionScaling(t *testing.T) {
	tests := []struct {
		name          string
		imageWidth    int
		imageHeight   int
		expectedError bool
		verifyImage   bool // Whether to verify the image dimensions in the EPUB
	}{
		{
			name:          "standard manga page",
			imageWidth:    1200,
			imageHeight:   1800,
			expectedError: false,
			verifyImage:   true,
		},
		{
			name:          "high resolution page",
			imageWidth:    2400,
			imageHeight:   3600,
			expectedError: false,
			verifyImage:   true,
		},
		{
			name:          "very small page",
			imageWidth:    300,
			imageHeight:   450,
			expectedError: false,
			verifyImage:   true,
		},
		{
			name:          "extra wide page",
			imageWidth:    3000,
			imageHeight:   1500,
			expectedError: false,
			verifyImage:   true,
		},
		{
			name:          "excessive resolution",
			imageWidth:    10000,
			imageHeight:   15000,
			expectedError: false, // Should scale down automatically
			verifyImage:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test manga with specified image dimensions
			manga := md.Manga{
				Info: md.MangaInfo{
					Title:   "Resolution Test Manga",
					ID:      "test-manga",
					Authors: []string{"Test Author"},
				},
				Volumes: map[md.Identifier]md.Volume{
					md.NewIdentifier("1"): {
						Info: md.VolumeInfo{
							Identifier: md.NewIdentifier("1"),
						},
						Chapters: map[md.Identifier]md.Chapter{
							md.NewIdentifier("1"): {
								Info: md.ChapterInfo{
									Title:            "Chapter 1",
									Identifier:       md.NewIdentifier("1"),
									VolumeIdentifier: md.NewIdentifier("1"),
								},
								Pages: map[int]image.Image{
									0: createTestImage(tt.imageWidth, tt.imageHeight, color.White),
								},
							},
						},
					},
				},
			}

			// Generate EPUB
			epub, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.expectedError {
				t.Errorf("GenerateEPUB() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if cleanup != nil {
				defer cleanup()
			}
			if tt.expectedError || epub == nil {
				return
			}

			if tt.verifyImage {
				// Write EPUB and get zip reader for verification
				zipReader, err := writeEPUB(t, epub)
				if err != nil {
					t.Fatalf("failed to write and open EPUB: %v", err)
				}

				// Find and validate image dimensions
				var foundImage bool
				for _, file := range zipReader.File {
					if strings.HasSuffix(file.Name, ".jpg") {
						foundImage = true
						rc, err := file.Open()
						if err != nil {
							t.Fatalf("failed to open image: %v", err)
						}
						defer rc.Close()

						img, _, err := image.Decode(rc)
						if err != nil {
							t.Fatalf("failed to decode image: %v", err)
						}

						bounds := img.Bounds()
						width := bounds.Dx()
						height := bounds.Dy()

						if width > 1600 {
							t.Errorf("image width %d exceeds maximum allowed 1600", width)
						}

						aspectRatio := float64(width) / float64(height)
						expectedRatio := float64(tt.imageWidth) / float64(tt.imageHeight)
						if math.Abs(aspectRatio-expectedRatio) > 0.01 {
							t.Errorf("aspect ratio changed: expected %.2f, got %.2f", expectedRatio, aspectRatio)
						}
					}
				}
				if !foundImage {
					t.Error("no image found in EPUB")
				}
			}
		})
	}
}

func TestInvalidImageDimensions(t *testing.T) {
	tests := []struct {
		name          string
		imageWidth    int
		imageHeight   int
		expectedError string
	}{
		{
			name:          "zero dimensions",
			imageWidth:    0,
			imageHeight:   0,
			expectedError: "invalid image dimensions",
		},
		{
			name:          "negative dimensions",
			imageWidth:    -100,
			imageHeight:   -150,
			expectedError: "invalid image dimensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test manga with invalid dimensions
			manga := md.Manga{
				Info: md.MangaInfo{
					Title:   "Invalid Dimensions Test",
					ID:      "test-id",
					Authors: []string{"Test Author"},
				},
				Volumes: map[md.Identifier]md.Volume{
					md.NewIdentifier("1"): {
						Info: md.VolumeInfo{
							Identifier: md.NewIdentifier("1"),
						},
						Chapters: map[md.Identifier]md.Chapter{
							md.NewIdentifier("1"): {
								Info: md.ChapterInfo{
									Title:            "Chapter 1",
									Identifier:       md.NewIdentifier("1"),
									VolumeIdentifier: md.NewIdentifier("1"),
								},
								Pages: map[int]image.Image{
									0: createTestImage(tt.imageWidth, tt.imageHeight, color.White),
								},
							},
						},
					},
				},
			}

			// Generate EPUB and verify it fails
			epub, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
			if err == nil {
				t.Error("expected error for invalid dimensions, got nil")
			}
			if cleanup != nil {
				defer cleanup()
			}
			if epub != nil {
				t.Error("expected nil EPUB for invalid dimensions")
			}
			if err != nil && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error containing %q, got %v", tt.expectedError, err)
			}
		})
	}
}
