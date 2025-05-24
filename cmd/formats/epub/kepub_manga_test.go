package epub

import (
	"testing"

	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

// TestKEPUBFromManga tests KEPUB conversion from a manga
func TestKEPUBFromManga(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Test with manga data
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
