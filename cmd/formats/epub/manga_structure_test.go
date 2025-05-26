package epub

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

func TestMangaStructureVariations(t *testing.T) {
	tests := []struct {
		name       string
		setupManga func() md.Manga
		wantErr    bool
		verify     func(t *testing.T, epub *epub.Epub)
	}{
		{
			name: "single volume manga",
			setupManga: func() md.Manga {
				return md.Manga{
					Info: md.MangaInfo{
						Title:   "Single Volume Test",
						Authors: []string{"Test Author"},
						ID:      "test-single",
					},
					Volumes: map[md.Identifier]md.Volume{
						md.NewIdentifier("1"): {
							Info: md.VolumeInfo{Identifier: md.NewIdentifier("1")},
							Chapters: map[md.Identifier]md.Chapter{
								md.NewIdentifier("1"): {
									Info: md.ChapterInfo{
										Title:            "Chapter 1",
										Language:         language.English,
										Identifier:       md.NewIdentifier("1"),
										VolumeIdentifier: md.NewIdentifier("1"),
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
			verify: func(t *testing.T, epub *epub.Epub) {
				verifyMangaEPUBStructure(t, epub, 1, 1, true)
			},
		},
		{
			name: "multi-volume no chapters",
			setupManga: func() md.Manga {
				manga := md.Manga{
					Info: md.MangaInfo{
						Title:   "Multi-Volume No Chapters",
						Authors: []string{"Test Author"},
						ID:      "test-multi-empty",
					},
					Volumes: make(map[md.Identifier]md.Volume),
				}

				// Add empty volumes
				for i := 1; i <= 3; i++ {
					volID := md.NewIdentifier(fmt.Sprintf("%d", i))
					manga.Volumes[volID] = md.Volume{
						Info:     md.VolumeInfo{Identifier: volID},
						Chapters: make(map[md.Identifier]md.Chapter),
						Cover:    createTestImage(1000, 1500, color.White),
					}
				}
				return manga
			},
			wantErr: true, // Should error with no chapters
		},
		{
			name: "multi-language manga",
			setupManga: func() md.Manga {
				return md.Manga{
					Info: md.MangaInfo{
						Title:   "Multi-Language Test",
						Authors: []string{"Test Author"},
						ID:      "test-multi-lang",
					},
					Volumes: map[md.Identifier]md.Volume{
						md.NewIdentifier("1"): {
							Info: md.VolumeInfo{Identifier: md.NewIdentifier("1")},
							Chapters: map[md.Identifier]md.Chapter{
								md.NewIdentifier("1"): {
									Info: md.ChapterInfo{
										Title:            "Chapter 1",
										Language:         language.English,
										Identifier:       md.NewIdentifier("1"),
										VolumeIdentifier: md.NewIdentifier("1"),
									},
									Pages: map[int]image.Image{
										0: createTestImage(800, 1200, color.White),
									},
								},
								md.NewIdentifier("2"): {
									Info: md.ChapterInfo{
										Title:            "Chapter 2",
										Language:         language.Japanese,
										Identifier:       md.NewIdentifier("2"),
										VolumeIdentifier: md.NewIdentifier("1"),
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
			verify: func(t *testing.T, epub *epub.Epub) {
				verifyMangaEPUBStructure(t, epub, 1, 2, true)
			},
		},
		{
			name: "nested chapter structure",
			setupManga: func() md.Manga {
				manga := md.Manga{
					Info: md.MangaInfo{
						Title:   "Nested Chapters Test",
						Authors: []string{"Test Author"},
						ID:      "test-nested",
					},
					Volumes: map[md.Identifier]md.Volume{
						md.NewIdentifier("1"): {
							Info:     md.VolumeInfo{Identifier: md.NewIdentifier("1")},
							Chapters: make(map[md.Identifier]md.Chapter),
							Cover:    createTestImage(1000, 1500, color.White),
						},
					},
				}

				// Add chapters with nested structure (1, 1.1, 1.2, 2, 2.1, etc.)
				vol := manga.Volumes[md.NewIdentifier("1")]
				for i := 1; i <= 2; i++ {
					// Main chapter
					mainID := md.NewIdentifier(fmt.Sprintf("%d", i))
					vol.Chapters[mainID] = md.Chapter{
						Info: md.ChapterInfo{
							Title:            fmt.Sprintf("Chapter %d", i),
							Language:         language.English,
							Identifier:       mainID,
							VolumeIdentifier: md.NewIdentifier("1"),
						},
						Pages: map[int]image.Image{
							0: createTestImage(800, 1200, color.White),
						},
					}

					// Sub-chapters
					for j := 1; j <= 2; j++ {
						subID := md.NewIdentifier(fmt.Sprintf("%d.%d", i, j))
						vol.Chapters[subID] = md.Chapter{
							Info: md.ChapterInfo{
								Title:            fmt.Sprintf("Chapter %d.%d", i, j),
								Language:         language.English,
								Identifier:       subID,
								VolumeIdentifier: md.NewIdentifier("1"),
							},
							Pages: map[int]image.Image{
								0: createTestImage(800, 1200, color.White),
							},
						}
					}
				}
				manga.Volumes[md.NewIdentifier("1")] = vol
				return manga
			},
			wantErr: false,
			verify: func(t *testing.T, epub *epub.Epub) {
				verifyMangaEPUBStructure(t, epub, 1, 6, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := tt.setupManga()
			epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cleanup != nil {
				// defer cleanup()
			}
			if epub == nil && !tt.wantErr {
				t.Fatal("GenerateEPUB() returned nil without error")
			}

			if tt.verify != nil {
				tt.verify(t, epub)
			}
		})
	}
}

// verifyMangaEPUBStructure checks the basic structure and content of an EPUB
func verifyMangaEPUBStructure(t *testing.T, epub *epub.Epub, expectedVolumes, expectedChapters int, shouldHaveCovers bool) {
	// Write EPUB and get zip reader for verification
	zipReader, err := writeEPUB(t, epub)
	if err != nil {
		t.Fatalf("failed to verify EPUB structure: %v", err)
	}

	// Verify required files exist
	requiredFiles := []string{
		"mimetype",
		"META-INF/container.xml",
		"EPUB/package.opf",
		"EPUB/nav.xhtml",
		"EPUB/css/style.css",
	}

	for _, required := range requiredFiles {
		found := false
		for _, f := range zipReader.File {
			if f.Name == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing required file: %s", required)
		}
	}

	fmt.Println("Files in EPUB archive:")
	for _, f := range zipReader.File {
		fmt.Println(f.Name)
	}
	chapterCount := 0
	imageCount := 0
	coverCount := 0
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "EPUB/xhtml/chapter-") {
			chapterCount++
		}
		if strings.HasPrefix(f.Name, "EPUB/images/") && strings.HasSuffix(f.Name, ".jpg") {
			imageCount++
			if strings.Contains(f.Name, "cover") {
				coverCount++
			}
		}
	}

	if chapterCount != expectedChapters {
		t.Errorf("expected %d chapters, got %d", expectedChapters, chapterCount)
	}

	if shouldHaveCovers {
		expectedCovers := expectedVolumes
		if coverCount != expectedCovers {
			t.Errorf("expected %d covers, got %d", expectedCovers, coverCount)
		}
	}

	// Verify navigation file structure
	var navFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "EPUB/nav.xhtml" {
			navFile = f
			break
		}
	}

	if navFile != nil {
		fmt.Printf("nav.xhtml size: %d bytes\n", navFile.UncompressedSize64)
		rc, err := navFile.Open()
		if err != nil {
			t.Fatalf("failed to open nav.xhtml: %v", err)
		}
		defer rc.Close()

		navContent, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("failed to read nav.xhtml: %v", err)
		}
		fmt.Printf("nav.xhtml raw content:\n%s\n", string(navContent))

		navStr := string(navContent)
		if expectedVolumes > 1 && !strings.Contains(navStr, "Volume") {
			t.Error("nav.xhtml missing volume entries")
		}
		if !strings.Contains(navStr, "Chapter") {
			t.Error("nav.xhtml missing chapter entries")
		}
	} else {
		t.Error("nav.xhtml not found")
	}
}
