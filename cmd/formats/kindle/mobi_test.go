package kindle

import (
	"image"
	"image/color"
	"strings"
	"testing"

	md "github.com/leotaku/kojirou/mangadex"
	"github.com/leotaku/mobi"
	"golang.org/x/text/language"
)

// TestGenerateMOBI tests the MOBI generation functionality with various input
func TestGenerateMOBI(t *testing.T) {
	// Create test cases with different manga configurations
	tests := []struct {
		name     string
		setup    func() md.Manga
		widepage WidepagePolicy
		autocrop bool
		ltr      bool
		validate func(*testing.T, mobi.Book)
	}{
		{
			name: "basic manga with standard pages",
			setup: func() md.Manga {
				return createTestManga()
			},
			widepage: WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			validate: func(t *testing.T, book mobi.Book) {
				// Validate core properties
				if book.Title == "" {
					t.Error("MOBI book title is empty")
				}
				if len(book.Images) == 0 {
					t.Error("MOBI book has no images")
				}
				if len(book.Chapters) == 0 {
					t.Error("MOBI book has no chapters")
				}
				if book.FixedLayout != true {
					t.Error("MOBI book should have fixed layout")
				}
			},
		},
		{
			name: "right-to-left manga",
			setup: func() md.Manga {
				return createTestManga()
			},
			widepage: WidepagePolicyPreserve,
			autocrop: false,
			ltr:      false, // Right-to-left
			validate: func(t *testing.T, book mobi.Book) {
				if !book.RightToLeft {
					t.Error("MOBI book should be right-to-left")
				}
			},
		},
		{
			name: "manga with wide page splitting",
			setup: func() md.Manga {
				manga := createTestManga()
				// Add a wide page to the first chapter
				for volID, vol := range manga.Volumes {
					for chapID, chap := range vol.Chapters {
						// Add a wide page (2:1 aspect ratio)
						chap.Pages[99] = createTestImage(2000, 1000, color.White)
						vol.Chapters[chapID] = chap
					}
					manga.Volumes[volID] = vol
					break // Just modify the first volume
				}
				return manga
			},
			widepage: WidepagePolicySplit, // Split wide pages
			autocrop: false,
			ltr:      true,
			validate: func(t *testing.T, book mobi.Book) {
				// The wide page should be split, resulting in more images than pages
				originalPageCount := countPages(createTestManga())
				if len(book.Images) <= originalPageCount {
					t.Errorf("Expected more than %d images due to wide page splitting, got %d",
						originalPageCount, len(book.Images))
				}
			},
		},
		{
			name: "manga with autocrop enabled",
			setup: func() md.Manga {
				return createTestManga()
			},
			widepage: WidepagePolicyPreserve,
			autocrop: true, // Enable autocrop
			ltr:      true,
			validate: func(t *testing.T, book mobi.Book) {
				// Basic validation as the actual cropping is tested elsewhere
				if len(book.Images) == 0 {
					t.Error("MOBI book has no images after autocrop")
				}
			},
		},
		{
			name: "manga with special characters in title",
			setup: func() md.Manga {
				manga := createTestManga()
				manga.Info.Title = "Test & Manga: Special < Characters > "
				return manga
			},
			widepage: WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			validate: func(t *testing.T, book mobi.Book) {
				if book.Title == "" {
					t.Error("MOBI book title is empty")
				}
				// The title should have the special characters
				if !containsSubstring(book.Title, "Special") {
					t.Errorf("Expected title to contain 'Special', got %s", book.Title)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := tt.setup()
			book := GenerateMOBI(manga, tt.widepage, tt.autocrop, tt.ltr)

			// Title should never be empty
			if book.Title == "" {
				t.Error("MOBI book title is empty")
			}

			// Book should have images
			if len(book.Images) == 0 {
				t.Error("MOBI book has no images")
			}

			// Book should have chapters
			if len(book.Chapters) == 0 {
				t.Error("MOBI book has no chapters")
			}

			// Run test-specific validations
			if tt.validate != nil {
				tt.validate(t, book)
			}
		})
	}
}

// TestMOBIMetadataExtraction tests the metadata extraction functions
func TestMOBIMetadataExtraction(t *testing.T) {
	manga := createTestManga()

	// Test title extraction
	title := mangaToTitle(manga)
	if title == "" {
		t.Error("mangaToTitle returned empty title")
	}

	// Test language extraction
	lang := mangaToLanguage(manga)
	if lang == language.Und {
		t.Error("mangaToLanguage returned undefined language")
	}

	// Test cover extraction
	cover := mangaToCover(manga)
	if cover == nil {
		t.Error("mangaToCover returned nil cover image")
	}

	// Test uniqueID extraction
	id := mangaToUniqueID(manga)
	if id == 0 {
		t.Error("mangaToUniqueID returned zero ID")
	}

	// Test with empty manga
	emptyManga := md.Manga{}
	emptyTitle := mangaToTitle(emptyManga)
	if emptyTitle == "" {
		t.Log("mangaToTitle handled empty manga correctly")
	} else {
		t.Errorf("Expected empty title for empty manga, got %s", emptyTitle)
	}

	emptyLang := mangaToLanguage(emptyManga)
	if emptyLang == language.Und {
		t.Log("mangaToLanguage handled empty manga correctly")
	} else {
		t.Errorf("Expected undefined language for empty manga, got %s", emptyLang)
	}
}

// TestTemplateRendering tests the template rendering function
func TestTemplateRendering(t *testing.T) {
	// Test normal template rendering
	result := templateToString(pageTemplate, "test")
	if result == "" {
		t.Error("templateToString returned empty string")
	}

	// Result should have the expected format
	expected := `<div>.</div><img src="kindle:embed:test?mime=image/jpeg">`
	if result != expected {
		t.Errorf("Expected template result %q, got %q", expected, result)
	}
}

// createTestManga creates a test manga with basic structure for testing
func createTestManga() md.Manga {
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
				Cover: createTestImage(1000, 1500, color.White),
				Chapters: map[md.Identifier]md.Chapter{
					md.NewIdentifier("1.1"): {
						Info: md.ChapterInfo{
							Title:            "Chapter 1",
							Identifier:       md.NewIdentifier("1.1"),
							VolumeIdentifier: md.NewIdentifier("1"),
							Language:         language.English,
							GroupNames:       []string{"Test Group"},
						},
						Pages: map[int]image.Image{
							0: createTestImage(800, 1200, color.White),
							1: createTestImage(800, 1200, color.White),
						},
					},
					md.NewIdentifier("1.2"): {
						Info: md.ChapterInfo{
							Title:            "Chapter 2",
							Identifier:       md.NewIdentifier("1.2"),
							VolumeIdentifier: md.NewIdentifier("1"),
							Language:         language.English,
							GroupNames:       []string{"Test Group"},
						},
						Pages: map[int]image.Image{
							0: createTestImage(800, 1200, color.White),
						},
					},
				},
			},
			md.NewIdentifier("2"): {
				Info: md.VolumeInfo{
					Identifier: md.NewIdentifier("2"),
				},
				Cover: createTestImage(1000, 1500, color.White),
				Chapters: map[md.Identifier]md.Chapter{
					md.NewIdentifier("2.1"): {
						Info: md.ChapterInfo{
							Title:            "Chapter 3",
							Identifier:       md.NewIdentifier("2.1"),
							VolumeIdentifier: md.NewIdentifier("2"),
							Language:         language.English,
							GroupNames:       []string{"Test Group"},
						},
						Pages: map[int]image.Image{
							0: createTestImage(800, 1200, color.White),
							1: createTestImage(800, 1200, color.White),
						},
					},
				},
			},
		},
	}
}

// createTestImage creates a simple test image of the specified size and color
func createTestImage(width, height int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// countPages counts the total number of pages in a manga
func countPages(manga md.Manga) int {
	count := 0
	for _, vol := range manga.Volumes {
		for _, chap := range vol.Chapters {
			count += len(chap.Pages)
		}
	}
	return count
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
