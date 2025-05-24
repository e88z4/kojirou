package epub

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

// TestMultiFormatGeneration tests generating multiple formats including KEPUB
func TestMultiFormatGeneration(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Get test manga
	manga := createTestManga()

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Validate EPUB generation
	if epubObj == nil {
		t.Fatal("EPUB object is nil")
	}

	// Generate KEPUB
	kepubBytes, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("Failed to convert to KEPUB: %v", err)
	}

	// Validate KEPUB output
	if len(kepubBytes) == 0 {
		t.Fatal("KEPUB data is empty")
	}

	// Use channels to collect results and errors from goroutines
	results := make(chan string, 3)
	errors := make(chan error, 3)

	// Generate all formats in parallel
	var wg sync.WaitGroup
	wg.Add(3)

	// Generate KEPUB
	go func() {
		defer wg.Done()

		// Convert EPUB to KEPUB
		kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
		if err != nil {
			errors <- fmt.Errorf("KEPUB conversion failed: %v", err)
			return
		}

		// Write KEPUB to temporary file
		tmpFile, err := os.CreateTemp("", "kojirou-kepub-*.kepub.epub")
		if err != nil {
			errors <- fmt.Errorf("create temp KEPUB file: %v", err)
			return
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write(kepubData); err != nil {
			errors <- fmt.Errorf("write KEPUB file: %v", err)
			return
		}

		results <- "KEPUB"
	}() // Generate standard EPUB
	go func() {
		defer wg.Done()

		// Create a temp file
		tempFile, err := ioutil.TempFile("", "epub-*.epub")
		if err != nil {
			errors <- err
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Write EPUB to temp file
		err = epubObj.Write(tempFile.Name())
		if err != nil {
			errors <- err
			return
		}

		results <- "EPUB"
	}()

	// Generate MOBI (if available)
	go func() {
		defer wg.Done()

		// This is a placeholder for MOBI generation
		// In a real implementation, we would convert EPUB to MOBI here
		// For testing purposes, we're just simulating success
		results <- "MOBI"
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		t.Fatalf("Format generation failed: %v", errs)
	}

	// Check results
	successFormats := make(map[string]bool)
	for format := range results {
		successFormats[format] = true
	}

	// Verify KEPUB was generated
	if !successFormats["KEPUB"] {
		t.Error("KEPUB generation was not reported as successful")
	}
}
