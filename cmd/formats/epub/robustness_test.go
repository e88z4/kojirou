package epub

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

// createGrayscaleImage creates a grayscale test image
func createGrayscaleImage(width, height int) image.Image {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{uint8((x + y) % 256)})
		}
	}
	return img
}

// createTransparentImage creates an image with transparency
func createTransparentImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := uint8((x + y) % 256)
			img.Set(x, y, color.RGBA{255, 255, 255, alpha})
		}
	}
	return img
}

// TestImageEdgeCases tests various edge cases for image processing
func TestImageEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		setupManga   func() md.Manga
		widepage     kindle.WidepagePolicy
		autocrop     bool
		wantErr      bool
		verifyOutput func(t *testing.T, epub *epub.Epub)
	}{
		{
			name: "manga with grayscale images",
			setupManga: func() md.Manga {
				m := createTestManga()
				for id, vol := range m.Volumes {
					for chapID, chap := range vol.Chapters {
						for pageNum := range chap.Pages {
							chap.Pages[pageNum] = createGrayscaleImage(800, 1200)
						}
						vol.Chapters[chapID] = chap
					}
					m.Volumes[id] = vol
				}
				return m
			},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			wantErr:  false,
		},
		{
			name: "manga with transparent images",
			setupManga: func() md.Manga {
				m := createTestManga()
				for id, vol := range m.Volumes {
					for chapID, chap := range vol.Chapters {
						for pageNum := range chap.Pages {
							chap.Pages[pageNum] = createTransparentImage(800, 1200)
						}
						vol.Chapters[chapID] = chap
					}
					m.Volumes[id] = vol
				}
				return m
			},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			wantErr:  false,
		},
		{
			name: "manga with tiny images",
			setupManga: func() md.Manga {
				m := createTestManga()
				for id, vol := range m.Volumes {
					for chapID, chap := range vol.Chapters {
						for pageNum := range chap.Pages {
							chap.Pages[pageNum] = createTestImage(50, 75, color.White)
						}
						vol.Chapters[chapID] = chap
					}
					m.Volumes[id] = vol
				}
				return m
			},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			wantErr:  false,
		},
		{
			name: "manga with huge images",
			setupManga: func() md.Manga {
				m := createTestManga()
				for id, vol := range m.Volumes {
					for chapID, chap := range vol.Chapters {
						for pageNum := range chap.Pages {
							chap.Pages[pageNum] = createTestImage(8000, 12000, color.White)
						}
						vol.Chapters[chapID] = chap
					}
					m.Volumes[id] = vol
				}
				return m
			},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			wantErr:  false,
		},
		{
			name: "manga with mixed image types",
			setupManga: func() md.Manga {
				m := createTestManga()
				for id, vol := range m.Volumes {
					for chapID, chap := range vol.Chapters {
						// Mix different image types
						chap.Pages[0] = createGrayscaleImage(800, 1200)
						if len(chap.Pages) > 1 {
							chap.Pages[1] = createTransparentImage(800, 1200)
						}
						vol.Chapters[chapID] = chap
					}
					m.Volumes[id] = vol
				}
				return m
			},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := tt.setupManga()
			epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, tt.widepage, tt.autocrop, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cleanup != nil {
				cleanup()
			}
			if epub == nil && !tt.wantErr {
				t.Fatal("GenerateEPUB() returned nil but expected success")
			}
			if epub != nil && tt.verifyOutput != nil {
				tt.verifyOutput(t, epub)
			}
		})
	}
}

// TestEPUBMetadataEdgeCases tests various edge cases for metadata handling
func TestEPUBMetadataEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		modifyInfo func(info *md.MangaInfo)
		wantErr    bool
		verify     func(t *testing.T, epub *epub.Epub)
	}{
		{
			name: "empty title",
			modifyInfo: func(info *md.MangaInfo) {
				info.Title = ""
			},
			wantErr: false, // Should use default title
		},
		{
			name: "very long title",
			modifyInfo: func(info *md.MangaInfo) {
				info.Title = strings.Repeat("Very long title ", 100)
			},
			wantErr: false,
		},
		{
			name: "special characters in title",
			modifyInfo: func(info *md.MangaInfo) {
				info.Title = "Test 📚 Manga ♥ with 特殊文字"
			},
			wantErr: false,
		},
		{
			name: "no authors",
			modifyInfo: func(info *md.MangaInfo) {
				info.Authors = nil
			},
			wantErr: false,
		},
		{
			name: "many authors",
			modifyInfo: func(info *md.MangaInfo) {
				info.Authors = make([]string, 100)
				for i := range info.Authors {
					info.Authors[i] = "Author " + string(rune('A'+i))
				}
			},
			wantErr: false,
		},
		{
			name: "special characters in authors",
			modifyInfo: func(info *md.MangaInfo) {
				info.Authors = []string{"作者 🎨", "Another ♪ Author"}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := createTestManga()
			tt.modifyInfo(&manga.Info)

			epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer func() {
				if cleanup != nil {
					cleanup()
				}
			}()
			if epub == nil && !tt.wantErr {
				t.Fatal("GenerateEPUB() returned nil but expected success")
			}
			if epub != nil && tt.verify != nil {
				tt.verify(t, epub)
			}
		})
	}
}
