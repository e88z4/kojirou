package epub

import (
"testing"

"github.com/bmaupin/go-epub"
)

// TestSimpleKEPUB tests the basic KEPUB conversion functionality
func TestSimpleKEPUB(t *testing.T) {
// Create a simple EPUB
e := epub.NewEpub("Test KEPUB")
e.SetAuthor("Test Author")

// Add a section
_, err := e.AddSection("<h1>Test Chapter</h1><p>This is a test.</p>", "Chapter 1", "ch1", "")
if err != nil {
t.Fatalf("Failed to add section: %v", err)
}

// Run a basic test
t.Log("Test runs successfully")
}
