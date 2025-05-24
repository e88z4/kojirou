package progress_test

import (
"testing"

"github.com/leotaku/kojirou/cmd/formats/progress"
)

// Test format functionality
func TestFormat(t *testing.T) {
p := progress.TitledProgress("Test")
p.SetFormat("epub")
p.Cancel("Done")
}
