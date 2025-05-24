package main

import (
	"fmt"
	"log"

	"github.com/bmaupin/go-epub"
)

func main() {
	// Create a simple EPUB
	e := epub.NewEpub("Test KEPUB")
	e.SetAuthor("Test Author")

	// Add a section
	_, err := e.AddSection("<h1>Test Chapter</h1><p>This is a test.</p>", "Chapter 1", "ch1", "")
	if err != nil {
		log.Fatalf("Failed to add section: %v", err)
	}

	// Write the EPUB to a file
	err = e.Write("test.epub")
	if err != nil {
		log.Fatalf("Failed to write EPUB: %v", err)
	}

	fmt.Println("Created test.epub successfully")
}
