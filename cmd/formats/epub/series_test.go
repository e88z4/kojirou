package epub

import (
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

// TestSeriesMetadataInKEPUB tests series metadata handling in KEPUB conversion
func TestSeriesMetadataInKEPUB(t *testing.T) {
	testCases := []struct {
		name        string
		title       string
		seriesTitle string
		seriesIndex float64
	}{
		{
			name:        "basic series metadata",
			title:       "Book Title",
			seriesTitle: "Test Series",
			seriesIndex: 1.0,
		},
		{
			name:        "series with decimal index",
			title:       "Book Title 2",
			seriesTitle: "Test Series",
			seriesIndex: 2.5,
		},
		{
			name:        "series with special characters",
			title:       "Book Title 3",
			seriesTitle: "Special: Characters & Test",
			seriesIndex: 3.0,
		},
		{
			name:        "very long series name",
			title:       "Book Title 4",
			seriesTitle: "This is an extremely long series name that tests the handling of long strings in metadata",
			seriesIndex: 4.0,
		},
		{
			name:        "no series metadata",
			title:       "Standalone Book",
			seriesTitle: "",
			seriesIndex: 0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a basic EPUB
			e := epub.NewEpub(tc.title)
			e.SetAuthor("Test Author")

			// Add basic content
			_, err := e.AddSection("<h1>Chapter 1</h1><p>Test content.</p>", "Chapter 1", "ch1", "")
			if err != nil {
				t.Fatalf("Failed to add section: %v", err)
			}

			// Convert to KEPUB with series metadata
			kepubData, err := kepubconv.ConvertToKEPUB(e, tc.seriesTitle, tc.seriesIndex)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Verify output
			if len(kepubData) == 0 {
				t.Error("KEPUB data is empty")
			}

			// Check series metadata handling
			if tc.seriesTitle != "" && tc.seriesIndex > 0 {
				validateSeriesMetadata(t, kepubData)
			} else {
				validateNoSeriesMetadata(t, kepubData)
			}
		})
	}
}

// TestSeriesMetadataWithManga tests series metadata with manga content
func TestSeriesMetadataWithManga(t *testing.T) {
	// Create a manga for testing
	manga := createTestManga()
	manga.Info.Title = "Test Manga Title"

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	// Convert to KEPUB with series metadata
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj, "Test Manga Series", 1.0)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Verify series metadata in output
	validateSeriesMetadata(t, kepubData)
}
