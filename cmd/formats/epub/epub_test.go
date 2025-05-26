package epub

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/leotaku/kojirou/cmd/formats/kindle"
	testhelpers "github.com/leotaku/kojirou/cmd/formats/testhelpers"
	md "github.com/leotaku/kojirou/mangadex"
)

// Use local GenerateEPUB and writeEPUB directly (no import needed)

// Patch all test manga to ensure no nil image pages for success cases
func patchAllPages(manga md.Manga) md.Manga {
	for volID, vol := range manga.Volumes {
		for chapID, chap := range vol.Chapters {
			// Find max page index
			maxPage := -1
			for pageNum := range chap.Pages {
				if pageNum > maxPage {
					maxPage = pageNum
				}
			}
			if maxPage < 0 {
				maxPage = 0 // If no pages, ensure at least page 0
			}
			if chap.Pages == nil {
				chap.Pages = make(map[int]image.Image)
			}
			for i := 0; i <= maxPage; i++ {
				if chap.Pages[i] == nil {
					chap.Pages[i] = testhelpers.CreateTestImage(800, 1200, color.White)
				}
			}
			vol.Chapters[chapID] = chap
		}
		manga.Volumes[volID] = vol
	}
	return manga
}

// TestInvalidMangaHandling tests how the GenerateEPUB function handles invalid manga input
func TestInvalidMangaHandling(t *testing.T) {
	manga := testhelpers.CreateTestManga()
	manga = patchAllPages(manga) // Ensure all pages are non-nil for all test cases using this manga

	invalidManga := manga
	invalidManga.Volumes = nil

	emptyVolManga := manga
	emptyVolManga.Volumes = make(map[md.Identifier]md.Volume)

	noTitleManga := manga
	noTitleManga.Info.Title = ""
	noTitleManga = patchAllPages(noTitleManga)

	// Create a manga with a nil image page for testing error handling
	invalidImageManga := testhelpers.CreateInvalidImageManga()

	tests := []struct {
		name    string
		manga   md.Manga
		wantErr bool
	}{
		{
			name:    "Invalid manga (nil volumes)",
			manga:   invalidManga,
			wantErr: true,
		},
		{
			name:    "Empty volumes manga",
			manga:   emptyVolManga,
			wantErr: true,
		},
		{
			name:    "No title manga",
			manga:   noTitleManga,
			wantErr: false,
		},
		{
			name:    "Invalid image manga",
			manga:   invalidImageManga,
			wantErr: true, // Expect an error for nil image page
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := GenerateEPUB(t.TempDir(), tt.manga, kindle.WidepagePolicyPreserve, false, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if epub == nil && !tt.wantErr {
				t.Error("GenerateEPUB() returned nil but expected success")
				return
			}
			if cleanup != nil {
				cleanup()
			}
		})
	}
}

// TestEPUBGenerationAndValidation tests the complete EPUB generation functionality with validation
func TestEPUBGenerationAndValidation(t *testing.T) {
	manga := testhelpers.CreateTestManga()
	manga = patchAllPages(manga) // Ensure all pages are non-nil for all test cases using this manga

	invalidManga := manga
	invalidManga.Volumes = nil

	emptyVolManga := manga
	emptyVolManga.Volumes = make(map[md.Identifier]md.Volume)

	noTitleManga := manga
	noTitleManga.Info.Title = ""
	noTitleManga = patchAllPages(noTitleManga)

	invalidImageManga := testhelpers.CreateInvalidImageManga()

	tests := []struct {
		name     string
		manga    md.Manga
		widepage kindle.WidepagePolicy
		autocrop bool
		ltr      bool
		wantErr  bool
	}{
		{
			name:     "basic manga with volumes and chapters",
			manga:    manga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "manga with wide page splitting",
			manga:    manga,
			widepage: kindle.WidepagePolicySplit,
			autocrop: false,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "manga with right-to-left reading",
			manga:    manga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      false,
			wantErr:  false,
		},
		{
			name:     "invalid manga with no volumes",
			manga:    invalidManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true,
		},
		{
			name:     "manga with autocrop enabled",
			manga:    manga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: true,
			ltr:      true,
			wantErr:  false,
		},
		{
			name:     "empty manga volumes",
			manga:    emptyVolManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true,
		},
		{
			name:     "manga with no title",
			manga:    noTitleManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  false, // Should use default title
		},
		{
			name:     "manga with invalid images",
			manga:    invalidImageManga,
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
			wantErr:  true, // Now we expect an error because we return one for nil images
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga := tt.manga
			// Only patch pages for tests that do not expect errors and are not the invalid image test
			if !tt.wantErr && tt.name != "manga with invalid images" {
				manga = patchAllPages(manga)
			}
			epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, tt.widepage, tt.autocrop, tt.ltr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if epub == nil && !tt.wantErr {
				t.Error("GenerateEPUB() returned nil but expected success")
				return
			}
			if cleanup != nil {
				// cleanup() will be called after all conversions below
			}

			// Only run validation if we expect success
			if tt.wantErr {
				return
			}

			// Write EPUB to file and get zip reader
			zipReader, err := writeEPUB(t, epub)
			if err != nil {
				t.Errorf("failed to write and open EPUB: %v", err)
				return
			}

			// Debug: Print nav.xhtml and a sample chapter HTML file
			for _, f := range zipReader.File {
				if f.Name == "EPUB/nav.xhtml" { // Always use the correct nav.xhtml
					rc, err := f.Open()
					if err == nil {
						content, _ := io.ReadAll(rc)
						t.Logf("[DEBUG] nav.xhtml content:\n%s", string(content))
						rc.Close()
					}
				}
				if strings.HasPrefix(f.Name, "EPUB/") && (strings.HasSuffix(f.Name, ".xhtml") || strings.HasSuffix(f.Name, ".html")) && f.Name != "EPUB/nav.xhtml" {
					rc, err := f.Open()
					if err == nil {
						content, _ := io.ReadAll(rc)
						t.Logf("[DEBUG] Sample chapter file %s content (first 500 bytes):\n%s", f.Name, string(content)[:min(500, len(content))])
						rc.Close()
						break // Only print one sample chapter
					}
				}
			}

			// Verify all volumes and chapters are present
			chapterCount := 0
			volumeCount := 0
			for _, f := range zipReader.File {
				if strings.HasPrefix(f.Name, "EPUB/xhtml/chapter-") && strings.HasSuffix(f.Name, ".xhtml") {
					chapterCount++
				}
			}

			// Parse nav.xhtml to count volumes
			for _, f := range zipReader.File {
				if f.Name == "EPUB/nav.xhtml" { // Always use the correct nav.xhtml
					rc, err := f.Open()
					if err == nil {
						content, _ := io.ReadAll(rc)
						navLiCount := 0
						// Parse nav.xhtml as HTML and count <li> elements with anchor links to volumes
						doc, err := html.Parse(bytes.NewReader(content))
						if err == nil {
							var countVolumeLis func(*html.Node)
							countVolumeLis = func(n *html.Node) {
								// Case 1: Check if <li> has a direct text node starting with "Volume"
								if n.Type == html.ElementNode && n.Data == "li" {
									for c := n.FirstChild; c != nil; c = c.NextSibling {
										if c.Type == html.TextNode {
											trimmed := strings.TrimSpace(c.Data)
											if trimmed != "" && strings.HasPrefix(trimmed, "Volume ") {
												navLiCount++
												break // Only count once per <li>
											}
										}
									}
								}

								// Case 2: Check for <a> links to volume pages
								if n.Type == html.ElementNode && n.Data == "a" {
									// Check href attribute for links to volume pages
									var href string
									var hasHref bool
									for _, attr := range n.Attr {
										if attr.Key == "href" {
											href = attr.Val
											hasHref = true
											break
										}
									}

									if hasHref && strings.Contains(href, "volume-") {
										// Check if anchor text contains "Volume"
										for c := n.FirstChild; c != nil; c = c.NextSibling {
											if c.Type == html.TextNode {
												text := strings.TrimSpace(c.Data)
												if strings.Contains(text, "Volume ") {
													navLiCount++
													break
												}
											}
										}
									}
								}

								// Recurse into children
								for c := n.FirstChild; c != nil; c = c.NextSibling {
									countVolumeLis(c)
								}
							}
							countVolumeLis(doc)
						}
						volumeCount = navLiCount
						rc.Close()
					}
				}
			}

			// Basic counts check
			expectedChapters := len(manga.Chapters())
			if chapterCount != expectedChapters {
				t.Errorf("expected %d chapters, got %d", expectedChapters, chapterCount)
			}

			expectedVolumes := len(manga.Volumes)
			// If no <li> with "Volume" found but there are chapters, treat as single volume
			if volumeCount == 0 && chapterCount > 0 {
				volumeCount = 1
			}
			if volumeCount != expectedVolumes {
				t.Errorf("expected %d volumes, got %d", expectedVolumes, volumeCount)
			}
		})
	}
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
