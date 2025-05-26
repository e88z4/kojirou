package epub

import (
	"testing"
	"fmt"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	"github.com/leotaku/kojirou/cmd/formats/testhelpers"
)

// TestKEPUBNilInput tests that ConvertToKEPUB handles nil input
func TestKEPUBNilInput(t *testing.T) {
	// Should fail because input is nil
	_, err := kepubconv.ConvertToKEPUB(nil)
	if err == nil {
		t.Error("ConvertToKEPUB() should fail with nil input")
	}
}

// TestKEPUBEmptyEPUB tests conversion with an empty EPUB
func TestKEPUBEmptyEPUB(t *testing.T) {
	// Create empty EPUB with only metadata
	emptyEpub := epub.NewEpub("Empty Test")
	emptyEpub.SetAuthor("Test Author")

	// Should fail because it has no content
	_, err := kepubconv.ConvertToKEPUB(emptyEpub)
	if err == nil {
		t.Error("ConvertToKEPUB() should fail with empty EPUB")
	}
}

// TestKEPUBSpecialCharacters tests handling of special characters in conversion
func TestKEPUBSpecialCharacters(t *testing.T) {
	// Create a test manga with special characters in the title
	manga := testhelpers.CreateSpecialCharTitleManga()

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		// cleanup() will be called after all conversions below
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	if len(kepubData) == 0 {
		t.Error("KEPUB data is empty")
	}
}

// TestKEPUBMalformedEPUB tests conversion with a malformed EPUB
func TestKEPUBMalformedEPUB(t *testing.T) {
	// Create a minimal but valid EPUB
	malformedEpub := epub.NewEpub("Malformed Test")
	malformedEpub.SetAuthor("Test Author")

	// Add a section with malformed HTML
	_, err := malformedEpub.AddSection("<h1>Malformed HTML</h2><p>This is malformed HTML.</p>", "Malformed", "malformed", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Try to convert to KEPUB - should still succeed since we try to handle malformed HTML
	kepubData, err := kepubconv.ConvertToKEPUB(malformedEpub)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed with malformed HTML: %v", err)
	}

	if len(kepubData) == 0 {
		t.Error("KEPUB data is empty")
	}
}

// TestKEPUBLargeContent tests conversion with large content
func TestKEPUBLargeContent(t *testing.T) {
	// Get test manga with large images
	manga := testhelpers.CreateLargeImageManga()

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Add more content to make it larger
	for i := 0; i < 10; i++ {
		sectionID := fmt.Sprintf("extra-%d", i)
		_, err := epubObj.AddSection(
			"<h1>Additional Chapter</h1><p>This is additional content to make the EPUB larger.</p>",
			"Extra Chapter",
			sectionID,
			"",
		)
		if err != nil {
			t.Fatalf("Failed to add section: %v", err)
		}
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	if len(kepubData) == 0 {
		t.Error("KEPUB data is empty")
	}
}
