package epub

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

// createLargeTestManga creates a test manga with many pages and chapters
func createLargeTestManga(numChapters, pagesPerChapter int) md.Manga {
	manga := md.Manga{
		Info: md.MangaInfo{
			Title: "Large Test Manga",
		},
		Volumes: make(map[md.Identifier]md.Volume),
	}

	// Create volumes with multiple chapters and pages
	for v := 1; v <= 2; v++ {
		volID := md.NewIdentifier(fmt.Sprintf("%d", v))
		volume := md.Volume{
			Info: md.VolumeInfo{
				Identifier: volID,
			},
			Chapters: make(map[md.Identifier]md.Chapter),
		}

		for c := 1; c <= numChapters; c++ {
			chapID := md.NewIdentifier(fmt.Sprintf("%d-%d", v, c))
			chapter := md.Chapter{
				Info: md.ChapterInfo{
					Title:            fmt.Sprintf("Chapter %d", c),
					Identifier:       chapID,
					VolumeIdentifier: volID,
				},
				Pages: make(map[int]image.Image),
			}

			// Add pages with varying sizes
			for p := 0; p < pagesPerChapter; p++ {
				width := 1200
				height := 1800
				if p%5 == 0 { // Every 5th page is wide
					width = 2400
					height = 1800
				}
				chapter.Pages[p] = createTestImage(width, height, color.White)
			}

			volume.Chapters[chapID] = chapter
		}
		manga.Volumes[volID] = volume
	}

	return manga
}

func BenchmarkEPUBGeneration(b *testing.B) {
	tests := []struct {
		name         string
		numChapters  int
		pagesPerChap int
		widepage     kindle.WidepagePolicy
		autocrop     bool
		ltr          bool
	}{
		{
			name:         "small manga - preserve wide",
			numChapters:  2,
			pagesPerChap: 10,
			widepage:     kindle.WidepagePolicyPreserve,
			autocrop:     false,
			ltr:          true,
		},
		{
			name:         "small manga - split wide",
			numChapters:  2,
			pagesPerChap: 10,
			widepage:     kindle.WidepagePolicySplit,
			autocrop:     false,
			ltr:          true,
		},
		{
			name:         "medium manga - preserve",
			numChapters:  5,
			pagesPerChap: 20,
			widepage:     kindle.WidepagePolicyPreserve,
			autocrop:     false,
			ltr:          true,
		},
		{
			name:         "large manga - preserve",
			numChapters:  10,
			pagesPerChap: 30,
			widepage:     kindle.WidepagePolicyPreserve,
			autocrop:     false,
			ltr:          true,
		},
		{
			name:         "large manga with autocrop",
			numChapters:  10,
			pagesPerChap: 30,
			widepage:     kindle.WidepagePolicyPreserve,
			autocrop:     true,
			ltr:          true,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			manga := createLargeTestManga(tt.numChapters, tt.pagesPerChap)

			// Create temp directory for benchmark output
			tmpDir := b.TempDir()

			// Get initial memory stats
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			startAlloc := m.Alloc

			// Start timing
			start := time.Now()

			// Run the benchmark
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				epub, cleanup, err := GenerateEPUB(manga, tt.widepage, tt.autocrop, tt.ltr)
				if err != nil {
					b.Fatalf("GenerateEPUB() failed: %v", err)
				}
				if cleanup != nil {
					defer cleanup()
				}

				// Write EPUB to temp file to get size
				epubPath := filepath.Join(tmpDir, fmt.Sprintf("test-%d.epub", i))
				if err := epub.Write(epubPath); err != nil {
					b.Fatalf("Write() failed: %v", err)
				}

				// Get file size
				info, err := os.Stat(epubPath)
				if err != nil {
					b.Fatalf("Stat() failed: %v", err)
				}
				fileSizeMB := float64(info.Size()) / (1024 * 1024)

				// Check file size constraints
				if fileSizeMB > 100 { // 100MB limit
					b.Errorf("EPUB file too large: %.2f MB", fileSizeMB)
				}
			}

			duration := time.Since(start)

			// Report memory usage
			runtime.ReadMemStats(&m)
			allocBytes := m.Alloc - startAlloc
			b.ReportMetric(float64(allocBytes)/float64(b.N), "B/op")

			// Calculate and report processing speed
			totalPages := tt.numChapters * tt.pagesPerChap * 2 // 2 volumes
			pagesPerSecond := float64(totalPages*b.N) / duration.Seconds()
			b.ReportMetric(pagesPerSecond, "pages/sec")

			// Report average time per page
			msPerPage := float64(duration.Milliseconds()) / float64(totalPages*b.N)
			b.ReportMetric(msPerPage, "ms/page")
		})
	}
}
