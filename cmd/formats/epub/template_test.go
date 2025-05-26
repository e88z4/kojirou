package epub

import (
	"archive/zip"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

func TestTemplateRendering(t *testing.T) {
	tests := []struct {
		name       string
		setupManga func() md.Manga
		validate   func(t *testing.T, epub *epub.Epub)
		wantErr    bool
	}{
		{
			name: "basic template",
			setupManga: func() md.Manga {
				return createTestManga()
			},
			validate: func(t *testing.T, e *epub.Epub) {
				// Write to temp file and verify contents
				tmpFile := filepath.Join(t.TempDir(), "test.epub")
				if err := e.Write(tmpFile); err != nil {
					t.Fatalf("failed to write EPUB: %v", err)
					return
				}

				// Open as ZIP to check rendered content
				r, err := zip.OpenReader(tmpFile)
				if err != nil {
					t.Fatalf("failed to open EPUB: %v", err)
					return
				}
				defer r.Close()

				// Look for key files
				foundNav := false
				foundCSS := false
				for _, f := range r.File {
					switch {
					case strings.HasSuffix(f.Name, "nav.xhtml"):
						foundNav = true
					case strings.HasSuffix(f.Name, "style.css"):
						foundCSS = true
					}
				}

				if !foundNav {
					t.Error("navigation file not found")
				}
				if !foundCSS {
					t.Error("stylesheet not found")
				}
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
				// cleanup() will be called after all conversions below
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, epub)
			}
		})
	}
}
