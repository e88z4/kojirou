package epub

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

func TestEPUBResolution(t *testing.T) {
	tests := []struct {
		name       string
		setupManga func() md.Manga
		wantErr    bool
		errString  string
		validate   func(*testing.T, *epub.Epub)
	}{
		{
			name: "valid manga",
			setupManga: func() md.Manga {
				return md.Manga{
					Info: md.MangaInfo{
						Title:   "Test Manga",
						Authors: []string{"Test Author"},
						ID:      "test-manga-id",
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
										Language:         language.English,
										Identifier:       md.NewIdentifier("1"),
										VolumeIdentifier: md.NewIdentifier("1"),
										GroupNames:       []string{"Test Group"},
									},
									Pages: map[int]image.Image{
										0: createTestImage(800, 1200, color.White),
									},
								},
							},
							Cover: createTestImage(1000, 1500, color.White),
						},
					},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, e *epub.Epub) {
				t.Helper()

				if e == nil {
					t.Fatal("EPUB is nil")
				}

				// Try to write the EPUB to validate it
				reader, err := writeEPUB(t, e)
				if err != nil {
					t.Fatalf("Failed to write EPUB: %v", err)
				}

				// Basic validation
				files := reader.File
				t.Logf("EPUB contains %d files", len(files))

				// Verify essential files
				hasOPF := false
				hasNCX := false
				hasCSS := false
				hasChapter := false
				hasImages := false

				for _, f := range files {
					name := f.Name
					t.Logf("Found file: %s", name)
					switch {
					case strings.Contains(name, ".opf"):
						hasOPF = true
					case strings.Contains(name, ".ncx"):
						hasNCX = true
					case strings.Contains(name, ".css"):
						hasCSS = true
					case strings.Contains(name, "chapter-") || strings.Contains(name, "ch1-") || strings.Contains(name, "section"):
						hasChapter = true
					case strings.Contains(name, ".jpg") || strings.Contains(name, ".jpeg"):
						hasImages = true
					}
				}

				// Report missing components
				if !hasOPF {
					t.Error("missing OPF file")
				}
				if !hasNCX {
					t.Error("missing NCX file")
				}
				if !hasCSS {
					t.Error("missing CSS file")
				}
				if !hasChapter {
					t.Error("missing chapter file")
				}
				if !hasImages {
					t.Error("missing image files")
				}

				// Title check
				title := e.Title()
				if title != "Test Manga" {
					t.Errorf("wrong title: got %q, want %q", title, "Test Manga")
				}
			},
		},
		{
			name: "invalid manga",
			setupManga: func() md.Manga {
				return md.Manga{
					Info: md.MangaInfo{
						Title: "Empty Manga",
					},
				}
			},
			wantErr:   true,
			errString: "manga has no volumes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := tt.setupManga()
			t.Logf("Testing manga: Title=%q, Volumes=%d", manga.Info.Title, len(manga.Volumes))

			epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, true)

			// Handle error cases first
			if err != nil {
				t.Logf("GenerateEPUB returned error: %v", err)

				if !tt.wantErr {
					t.Errorf("GenerateEPUB() failed with error: %v", err)
					return
				}

				if !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("error %q does not contain expected %q", err.Error(), tt.errString)
				}
				return
			}

			if cleanup != nil {
				defer cleanup()
			}

			// No error occurred
			t.Log("GenerateEPUB completed successfully")

			if tt.wantErr {
				t.Error("GenerateEPUB() succeeded but expected error")
				return
			}

			// For valid cases, run the validation
			if tt.validate != nil {
				tt.validate(t, epub)
			}
		})
	}
}

func TestEPUBResolutions(t *testing.T) {
	tests := []struct {
		name          string
		imageWidth    int
		imageHeight   int
		expectedWidth int
		expectedError bool
	}{
		{
			name:          "standard manga page",
			imageWidth:    1200,
			imageHeight:   1800,
			expectedWidth: 1200,
			expectedError: false,
		},
		{
			name:          "high resolution page",
			imageWidth:    2400,
			imageHeight:   3600,
			expectedWidth: 1600, // Should be scaled down
			expectedError: false,
		},
		{
			name:          "very small page",
			imageWidth:    300,
			imageHeight:   450,
			expectedWidth: 300,
			expectedError: false,
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
			epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.expectedError {
				t.Errorf("GenerateEPUB() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			defer func() {
				if cleanup != nil {
					cleanup()
				}
			}()
			if tt.expectedError || epub == nil {
				return
			}

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

					width := img.Bounds().Dx()
					if width != tt.expectedWidth {
						t.Errorf("expected image width %d, got %d", tt.expectedWidth, width)
					}
				}
			}
			if !foundImage {
				t.Error("no image found in EPUB")
			}
		})
	}
}
