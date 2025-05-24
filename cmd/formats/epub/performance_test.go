package epub

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

// createPerformanceTestManga is a wrapper around createLargeTestManga
// Keeping this function for backward compatibility
func createPerformanceTestManga() md.Manga {
	return createLargeTestManga(10, 20)
}

func BenchmarkEPUBPerformance(b *testing.B) {
	tests := []struct {
		name     string
		manga    md.Manga
		widepage kindle.WidepagePolicy
		autocrop bool
		ltr      bool
	}{
		{
			name:     "small manga LTR no split",
			manga:    createTestManga(),
			widepage: kindle.WidepagePolicyPreserve,
			autocrop: false,
			ltr:      true,
		},
		{
			name:     "large manga RTL with split",
			manga:    createLargeTestManga(10, 20),
			widepage: kindle.WidepagePolicySplit,
			autocrop: false,
			ltr:      false,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			tmpDir := b.TempDir()

			start := time.Now()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			startAlloc := m.Alloc

			for i := 0; i < b.N; i++ {
				epub, cleanup, err := GenerateEPUB(tt.manga, tt.widepage, tt.autocrop, tt.ltr)
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
			b.ReportMetric(float64(duration.Milliseconds())/float64(b.N), "ms/op")

			runtime.ReadMemStats(&m)
			allocBytes := m.Alloc - startAlloc
			b.ReportMetric(float64(allocBytes)/float64(b.N), "B/op")

			// Calculate and report images per second
			totalImages := countTotalImages(tt.manga)
			imagesPerSecond := float64(totalImages*b.N) / duration.Seconds()
			b.ReportMetric(imagesPerSecond, "images/sec")
		})
	}
}

func countTotalImages(manga md.Manga) int {
	total := len(manga.Volumes) // Volume covers
	for _, vol := range manga.Volumes {
		for _, chap := range vol.Chapters {
			total += len(chap.Pages)
		}
	}
	return total
}

func TestPerformanceConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test")
	}

	tests := []struct {
		name          string
		manga         md.Manga
		widepage      kindle.WidepagePolicy
		maxDurationMs int64
		maxMemoryMB   int64
		maxFileSizeMB int64
	}{
		{
			name:          "small manga performance",
			manga:         createTestManga(),
			widepage:      kindle.WidepagePolicyPreserve,
			maxDurationMs: 1000, // 1 second
			maxMemoryMB:   100,  // 100 MB
			maxFileSizeMB: 20,   // 20 MB
		},
		{
			name:          "large manga performance",
			manga:         createLargeTestManga(10, 20),
			widepage:      kindle.WidepagePolicyPreserve,
			maxDurationMs: 5000, // 5 seconds
			maxMemoryMB:   500,  // 500 MB
			maxFileSizeMB: 100,  // 100 MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			var startStats runtime.MemStats
			runtime.ReadMemStats(&startStats)
			startAlloc := startStats.Alloc

			epub, cleanup, err := GenerateEPUB(tt.manga, tt.widepage, false, true)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Check execution time
			duration := time.Since(start)
			if duration.Milliseconds() > tt.maxDurationMs {
				t.Errorf("execution took %v ms, want < %v ms", duration.Milliseconds(), tt.maxDurationMs)
			}

			// Check memory usage
			var endStats runtime.MemStats
			runtime.ReadMemStats(&endStats)
			allocBytes := endStats.Alloc - startAlloc
			allocMB := int64(allocBytes / (1024 * 1024))
			if allocMB > tt.maxMemoryMB {
				t.Errorf("memory usage %v MB, want < %v MB", allocMB, tt.maxMemoryMB)
			}

			// Check output file size by writing to temp file
			tmpFile := filepath.Join(t.TempDir(), "test.epub")
			if err := epub.Write(tmpFile); err != nil {
				t.Fatal("failed to write EPUB:", err)
			}

			info, err := os.Stat(tmpFile)
			if err != nil {
				t.Fatal("failed to get file info:", err)
			}

			fileSizeMB := float64(info.Size()) / (1024 * 1024)
			if fileSizeMB > float64(tt.maxFileSizeMB) {
				t.Errorf("file size %.2f MB, want < %v MB", fileSizeMB, tt.maxFileSizeMB)
			}
		})
	}
}
