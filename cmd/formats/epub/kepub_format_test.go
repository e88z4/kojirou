package epub

import (
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

// TestKEPUBGeneration tests basic KEPUB format generation
func TestKEPUBGeneration(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	manga := createTestManga()

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
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

// TestKEPUBWithWidePages tests KEPUB generation with wide page handling
func TestKEPUBWithWidePages(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	manga := createWidePageTestManga()

	// Generate EPUB with split wide pages
	epubObj, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicySplit, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
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

// TestKEPUBErrorHandling tests error cases for KEPUB conversion
func TestKEPUBErrorHandling(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Test nil input
	_, err := kepubconv.ConvertToKEPUB(nil)
	if err == nil {
		t.Error("ConvertToKEPUB() should fail with nil input")
	}

	// Test empty EPUB
	emptyEpub := epub.NewEpub("Empty Test")
	_, err = kepubconv.ConvertToKEPUB(emptyEpub)
	if err == nil {
		t.Error("ConvertToKEPUB() should fail with empty EPUB")
	}
}
