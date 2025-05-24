package epub

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

func TestEPUBErrors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (md.Manga, error)
		wantErr   bool
		errString string
	}{
		{
			name: "empty manga",
			setup: func() (md.Manga, error) {
				return md.Manga{}, nil
			},
			wantErr:   true,
			errString: "manga has no volumes",
		},
		{
			name: "invalid chapter",
			setup: func() (md.Manga, error) {
				manga := createTestManga()
				// Add an invalid chapter
				for id, vol := range manga.Volumes {
					vol.Chapters[md.NewIdentifier("invalid")] = md.Chapter{}
					manga.Volumes[id] = vol
				}
				return manga, nil
			},
			wantErr:   true,
			errString: "no pages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga, err := tt.setup()
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			epub, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cleanup != nil {
				defer cleanup()
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error containing %q, got %v", tt.errString, err)
				}
				return
			}

			// Verify we can write the EPUB
			epubPath := filepath.Join(t.TempDir(), "test.epub")
			if err := epub.Write(epubPath); err != nil {
				t.Errorf("Write() failed: %v", err)
			}

			// Verify the file exists
			if _, err := os.Stat(epubPath); err != nil {
				t.Errorf("EPUB file not created: %v", err)
			}
		})
	}
}
