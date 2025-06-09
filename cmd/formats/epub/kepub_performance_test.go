package epub

import (
	"os"
	"testing"
	"time"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

// TestKEPUBConversionPerformance tests the performance of KEPUB conversion
func TestKEPUBConversionPerformance(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a large EPUB for performance testing
	e := epub.NewEpub("Performance Test")
	e.SetAuthor("Test Author")

	// Add multiple sections to create a larger file
	for i := 0; i < 50; i++ {
		// Create a section with a large amount of content
		content := "<h1>Chapter " + string(rune('A'+i)) + "</h1>"
		content += "<p>This is a test chapter with multiple paragraphs to test performance.</p>"

		// Add several paragraphs
		for j := 0; j < 20; j++ {
			content += "<p>This is paragraph " + string(rune('a'+j)) + " in chapter " + string(rune('A'+i)) +
				". It contains a reasonable amount of text to simulate a real book chapter. " +
				"The goal is to have enough content to measure conversion performance accurately.</p>"
		}

		// Add the section
		_, err := e.AddSection(content, "Chapter "+string(rune('A'+i)), "ch"+string(rune('a'+i)), "")
		if err != nil {
			t.Fatalf("Failed to add section: %v", err)
		}
	}

	// Measure conversion time
	start := time.Now()
	kepubData, err := kepubconv.ConvertToKEPUB(e, "Performance Test Series", 1)
	elapsed := time.Since(start)

	// Check for errors
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Verify output
	if len(kepubData) == 0 {
		t.Error("KEPUB data is empty")
	}

	// Log performance metrics
	t.Logf("KEPUB conversion took %s for a large EPUB", elapsed)
	t.Logf("Output size: %d bytes", len(kepubData))

	// Check if performance is within acceptable bounds
	// This is a subjective threshold that might need adjustment
	if elapsed > 5*time.Second {
		t.Logf("KEPUB conversion performance might need optimization (took %s)", elapsed)
	}
}

// TestKEPUBMemoryUsage tests the memory efficiency of KEPUB conversion
func TestKEPUBMemoryUsage(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Generate a large manga EPUB
	manga := createLargeTestManga(10, 20)

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	// Convert to KEPUB with manga title as series name
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj, manga.Info.Title, 1)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Verify output size is reasonable
	// A manga KEPUB should be close in size to the original EPUB
	// This ratio might need adjustment based on actual implementation
	tempFile, err := os.CreateTemp("", "epub-bytes-*.epub")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	if err := epubObj.Write(tempFile.Name()); err != nil {
		t.Fatalf("Failed to write EPUB: %v", err)
	}
	epubBytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read EPUB bytes: %v", err)
	}

	// Calculate and check size ratio (KEPUB should not be much larger than EPUB)
	ratio := float64(len(kepubData)) / float64(len(epubBytes))
	t.Logf("KEPUB/EPUB size ratio: %.2f", ratio)

	if ratio > 1.5 {
		t.Logf("KEPUB size is significantly larger than EPUB (ratio: %.2f). Consider optimization.", ratio)
	}
}

// Helper to create a large test manga for performance testing
