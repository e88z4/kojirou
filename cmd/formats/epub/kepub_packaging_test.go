// Package epub provides functionality for generating EPUB format ebooks from manga
package epub

import (
"testing"
)

func TestKEPUBExtension(t *testing.T) {
// Verify KEPUB extension is defined correctly
if KEPUBExtension != ".kepub.epub" {
t.Errorf("Expected KEPUB extension to be \\\".kepub.epub\\\", got %s", KEPUBExtension)
}
}
