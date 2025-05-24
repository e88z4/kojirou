package epub

import (
	"path/filepath"
	"testing"

	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

func TestEPUBIntegration(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() md.Manga
		widepage kindle.WidepagePolicy
		autocrop bool
		ltr      bool
	}{
		{
			name: "basic integration",
			setup: func() md.Manga {
				return createTestManga()
			},
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test manga
			manga := tt.setup()

			// Generate EPUB
			epub, cleanup, err := GenerateEPUB(manga, tt.widepage, tt.autocrop, tt.ltr)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Write EPUB to temp file
			tmpFile := filepath.Join(t.TempDir(), "test.epub")
			if err := epub.Write(tmpFile); err != nil {
				t.Fatalf("failed to write EPUB: %v", err)
			}
		})
	}
}
