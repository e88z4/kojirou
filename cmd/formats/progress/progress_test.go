package progress_test

import (
"testing"

"github.com/leotaku/kojirou/cmd/formats/progress"
)

// Basic tests for progress reporting functionality
func TestProgressBasics(t *testing.T) {
// Test TitledProgress
p1 := progress.TitledProgress("Test Title")
p1.Add(10)
p1.Done()

// Test FormatTitledProgress
p2 := progress.FormatTitledProgress("Test Format", "epub")
p2.Add(10)
p2.Done()

// Test VanishingProgress
p3 := progress.VanishingProgress("Test Vanishing")
p3.Add(10)
p3.Done()

// Test FormatVanishingProgress
p4 := progress.FormatVanishingProgress("Test Format Vanishing", "mobi")
p4.Add(10)
p4.Done()
}
