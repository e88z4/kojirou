package progress_test

import (
	"testing"

	"github.com/leotaku/kojirou/cmd/formats/progress"
)

// Test format tracking functionality
func TestFormatTracking(t *testing.T) {
	formats := []string{"epub", "mobi", "kepub"}
	p := progress.MultiFormatStatusProgress("Test Multi", formats)

	for _, format := range formats {
		p.FormatCompleted(format, "Success")
	}

	p.Done()
}

// Test format-specific messages
func TestFormatMessages(t *testing.T) {
	p := progress.TitledProgress("Test")

	// Test setting format
	p.SetFormat("epub")

	// Test cancellation with format
	p.CancelWithFormat("mobi", "Error")
}
