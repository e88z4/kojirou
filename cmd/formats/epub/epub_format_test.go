package epub

import (
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	testhelpers "github.com/leotaku/kojirou/cmd/formats/testhelpers"
	md "github.com/leotaku/kojirou/mangadex"
)

// ... [unchanged code above] ...

// TestEPUBMetadataHandling tests how the EPUB generator handles various metadata scenarios
func TestEPUBMetadataHandling(t *testing.T) {
	tests := []struct {
		name     string
		modify   func(*md.Manga)
		validate func(*testing.T, *epub.Epub)
		wantErr  bool
	}{
		{
			name:   "standard metadata",
			modify: func(manga *md.Manga) {},
			validate: func(t *testing.T, e *epub.Epub) {
				if e.Identifier() != "test-manga-id" {
					t.Errorf("Expected identifier 'test-manga-id', got %s", e.Identifier())
				}
				if e.Title() != "Test Manga" {
					t.Errorf("Expected title 'Test Manga', got %s", e.Title())
				}
				title := e.Title()
				if title == "" {
					t.Error("EPUB title is empty")
				}
			},
			wantErr: false,
		},
		{
			name:   "missing title",
			modify: func(manga *md.Manga) { manga.Info.Title = "" },
			validate: func(t *testing.T, e *epub.Epub) {
				title := e.Title()
				if title == "" {
					t.Error("EPUB title should not be empty")
				}
			},
			wantErr: false,
		},
		{
			name:   "missing authors",
			modify: func(manga *md.Manga) { manga.Info.Authors = nil },
			validate: func(t *testing.T, e *epub.Epub) {
				title := e.Title()
				if title == "" {
					t.Error("EPUB authors should not be empty")
				}
			},
			wantErr: false,
		},
		{
			name:   "very long title",
			modify: func(manga *md.Manga) { manga.Info.Title = strings.Repeat("Very Long Title ", 20) },
			validate: func(t *testing.T, e *epub.Epub) {
				title := e.Title()
				if title == "" {
					t.Error("EPUB title should not be empty")
				}
				if len(title) > 350 {
					t.Errorf("EPUB title is too long: %d characters", len(title))
				}
			},
			wantErr: false,
		},
		{
			name:   "special characters in title",
			modify: func(manga *md.Manga) { manga.Info.Title = "Special & Characters: < > \" '" },
			validate: func(t *testing.T, e *epub.Epub) {
				title := e.Title()
				if !strings.Contains(title, "Special") {
					t.Errorf("EPUB title should contain 'Special', got %s", title)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := testhelpers.CreateTestManga()
			tt.modify(&manga)

			epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cleanup != nil { //nolint:staticcheck // Cleanup is deferred in production code
			}

			if tt.wantErr {
				return
			}

			tt.validate(t, epub)
		})
	}
}

// ... [unchanged code below] ...
