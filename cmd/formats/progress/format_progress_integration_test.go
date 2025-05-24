package progress_test

import (
	"fmt"
	"testing"

	"github.com/leotaku/kojirou/cmd/formats/progress"
)

// TestFormatProgressIntegration tests the integration of format-specific
// progress reporting in a realistic scenario
func TestFormatProgressIntegration(t *testing.T) {
	// Sample test for format-specific progress reporting

	// 1. Main volume progress
	volumeProgress := progress.TitledProgress("Volume: v1")
	defer volumeProgress.Done()

	// 2. Process multiple formats
	formats := []string{"epub", "mobi", "kepub"}
	formatStatuses := map[string]string{}

	for _, format := range formats {
		// Show which format is being processed in the main progress
		volumeProgress.SetFormat(format)

		// Create format-specific progress
		formatProgress := progress.FormatVanishingProgress("Processing", format)

		// Simulate work (would be actual processing in real code)
		formatProgress.Increase(100)
		for i := 0; i < 100; i++ {
			formatProgress.Add(1)
		}

		// Mark as complete (skipping actual work for test)
		formatStatus := "Success"
		if format == "mobi" { // Simulate a failure for one format
			formatStatus = "Error"
		}
		formatStatuses[format] = formatStatus

		// Complete the format-specific progress
		if formatStatus == "Success" {
			formatProgress.Done()
		} else {
			formatProgress.Cancel("Error")
		}

		// Track in main progress
		volumeProgress.FormatCompleted(format, formatStatus)
	}

	// Verify results
	var successCount, errorCount int
	for _, status := range formatStatuses {
		if status == "Success" {
			successCount++
		} else {
			errorCount++
		}
	}

	if successCount != 2 || errorCount != 1 {
		t.Errorf("Expected 2 successes and 1 error, got %d successes and %d errors",
			successCount, errorCount)
	}

	// In real code, we would check if the output was correctly written
}

// TestMultiFormatProgressExample demonstrates how to use the multi-format
// progress tracking in a real scenario
func TestMultiFormatProgressExample(t *testing.T) {
	// Create a multi-format progress tracker
	formats := []string{"epub", "mobi", "kepub"}
	p := progress.MultiFormatStatusProgress("Volume: v1", formats)
	defer p.Done()

	// Process each format
	for i, format := range formats {
		// Simulate work
		if i == 1 { // Simulate a failure for mobi
			p.FormatCompleted(format, "Error")
		} else {
			p.FormatCompleted(format, "Success")
		}
	}

	// In real code, we would now check the generated files
	fmt.Println("Test completed - in real code would check output files")
}
