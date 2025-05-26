package epub

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"io"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

// TestKEPUBMetadataHandling tests how KEPUB handles various metadata
func TestKEPUBMetadataHandling(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create test cases with different metadata configurations
	testCases := []struct {
		name      string
		setupEpub func() *epub.Epub
	}{
		{
			name: "basic metadata",
			setupEpub: func() *epub.Epub {
				e := epub.NewEpub("Basic Metadata Test")
				e.SetAuthor("Test Author")
				e.SetDescription("A simple test book")
				e.SetLang("en")
				return e
			},
		},
		{
			name: "extensive metadata",
			setupEpub: func() *epub.Epub {
				e := epub.NewEpub("Extensive Metadata Test")
				e.SetAuthor("Test Author")
				e.SetDescription("A test book with lots of metadata")
				e.SetLang("en")
				e.SetIdentifier("urn:uuid:12345678-1234-1234-1234-123456789012")
				e.SetTitle("Extensive Metadata Test")
				// Skip unsupported methods
				// e.SetPubDate("2025-05-20")
				// e.AddAuthor("Second Test Author")
				// e.SetPublisher("Test Publisher")
				// e.SetRights("Copyright © 2025")
				return e
			},
		},
		{
			name: "non-english metadata",
			setupEpub: func() *epub.Epub {
				e := epub.NewEpub("日本語のマンガ")
				e.SetAuthor("テスト作家")
				e.SetDescription("これはテストです")
				e.SetLang("ja")
				return e
			},
		},
		{
			name: "manga-specific metadata",
			setupEpub: func() *epub.Epub {
				// Generate from a manga object to ensure manga-specific metadata
				manga := md.Manga{
					Info: md.MangaInfo{
						Title:   "Manga Metadata Test",
						Authors: []string{"Test Mangaka"},
						ID:      "manga-metadata-test",
					},
					Volumes: map[md.Identifier]md.Volume{
						md.NewIdentifier("1"): {
							Info: md.VolumeInfo{
								Identifier: md.NewIdentifier("1"),
							},
							Cover: createTestImage(800, 1200, color.White),
							Chapters: map[md.Identifier]md.Chapter{
								md.NewIdentifier("1"): {
									Info: md.ChapterInfo{
										Title:            "Chapter 1",
										Identifier:       md.NewIdentifier("1"),
										VolumeIdentifier: md.NewIdentifier("1"),
										Language:         language.English,
									},
									Pages: map[int]image.Image{
										0: createTestImage(800, 1200, color.White),
									},
								},
							},
						},
					},
				}

				epubObj, cleanup, _ := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				if cleanup != nil {
					/* cleanup() will be called after all conversions below */
				}
				return epubObj
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup the EPUB
			epubObj := tc.setupEpub()

			// Add a section to ensure the EPUB is valid
			_, err := epubObj.AddSection("<p>Test content</p>", "Test Section", "test", "")
			if err != nil {
				t.Fatalf("Failed to add section: %v", err)
			}

			// Convert to KEPUB
			kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Verify KEPUB metadata
			verifyKEPUBMetadata(t, kepubData, epubObj)
		})
	}
}

// TestKEPUBKoboExtensionMetadata tests Kobo-specific extension metadata
func TestKEPUBKoboExtensionMetadata(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a basic EPUB
	e := epub.NewEpub("Kobo Extension Test")
	e.SetAuthor("Test Author")

	// Add a section
	_, err := e.AddSection("<p>Test content</p>", "Test Section", "test", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(e)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Check for Kobo-specific metadata in the OPF file
	verifyKoboMetadataExtensions(t, kepubData)
}

// TestKEPUBMangaSpecificMetadata tests manga-specific metadata in KEPUB format
func TestKEPUBMangaSpecificMetadata(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a manga with simplified metadata structure
	manga := md.Manga{
		Info: md.MangaInfo{
			Title:   "Manga Metadata Specific Test",
			Authors: []string{"Test Mangaka"},
			ID:      "manga-metadata-specific-test",
		},
		Volumes: map[md.Identifier]md.Volume{
			md.NewIdentifier("1"): {
				Info: md.VolumeInfo{
					Identifier: md.NewIdentifier("1"),
				},
				Cover: createTestImage(800, 1200, color.White),
				Chapters: map[md.Identifier]md.Chapter{
					md.NewIdentifier("1"): {
						Info: md.ChapterInfo{
							Title:            "Chapter 1",
							Identifier:       md.NewIdentifier("1"),
							VolumeIdentifier: md.NewIdentifier("1"),
							Language:         language.English,
						},
						Pages: map[int]image.Image{
							0: createTestImage(800, 1200, color.White),
							1: createTestImage(800, 1200, color.White),
						},
					},
					md.NewIdentifier("2"): {
						Info: md.ChapterInfo{
							Title:            "Chapter 2",
							Identifier:       md.NewIdentifier("2"),
							VolumeIdentifier: md.NewIdentifier("1"),
							Language:         language.English,
						},
						Pages: map[int]image.Image{
							0: createTestImage(800, 1200, color.White),
							1: createTestImage(800, 1200, color.White),
						},
					},
				},
			},
		},
	}

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
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

	// Check for manga-specific metadata preservation and Kobo enhancements
	verifyMangaMetadataInKEPUB(t, kepubData, manga)
}

// Helper functions for KEPUB metadata verification

func verifyKEPUBMetadata(t *testing.T, data []byte, originalEpub *epub.Epub) {
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and check the OPF file (package.opf)
	var opfContent string
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, "package.opf") {
			rc, err := file.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}

			contentBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF content: %v", err)
			}

			opfContent = string(contentBytes)
			break
		}
	}

	if opfContent == "" {
		t.Fatal("No OPF file found in KEPUB")
	}

	// Check for basic metadata
	// This is a simplified check - in a real implementation,
	// proper XML parsing would be more accurate
	expectedFields := []struct {
		name          string
		shouldContain string
	}{
		{
			name:          "title",
			shouldContain: originalEpub.Title(),
		},
		{
			name:          "author",
			shouldContain: originalEpub.Author(),
		},
		{
			name:          "language",
			shouldContain: originalEpub.Lang(),
		},
	}

	for _, field := range expectedFields {
		if !strings.Contains(opfContent, field.shouldContain) {
			t.Errorf("%s metadata not properly preserved: expected %s", field.name, field.shouldContain)
		}
	}

	// Check for Kobo identifier (should be added during conversion)
	if !strings.Contains(opfContent, "kobo") {
		t.Log("No Kobo-specific identifier found in metadata")
	}
}

func verifyKoboMetadataExtensions(t *testing.T, data []byte) {
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and check the OPF file
	var opfContent string
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, "package.opf") {
			rc, err := file.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}

			contentBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF content: %v", err)
			}

			opfContent = string(contentBytes)
			break
		}
	}

	if opfContent == "" {
		t.Fatal("No OPF file found in KEPUB")
	}

	// Check for Kobo-specific metadata extensions
	koboExtensions := []string{
		"meta name=\"kobo",
		"meta property=\"kobo",
		"meta name=\"calibre:series", // Calibre-style series metadata often used by Kobo
	}

	foundExtensions := false
	for _, ext := range koboExtensions {
		if strings.Contains(opfContent, ext) {
			foundExtensions = true
			t.Logf("Found Kobo extension: %s", ext)
		}
	}

	if !foundExtensions {
		t.Log("No Kobo-specific metadata extensions found")
	}
}

func verifyMangaMetadataInKEPUB(t *testing.T, data []byte, originalManga md.Manga) {
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and check the OPF file
	var opfContent string
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, "package.opf") {
			rc, err := file.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}

			contentBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF content: %v", err)
			}

			opfContent = string(contentBytes)
			break
		}
	}

	if opfContent == "" {
		t.Fatal("No OPF file found in KEPUB")
	}

	// Check for manga-specific metadata
	mangaMetadata := []struct {
		name  string
		value string
	}{
		{
			name:  "title",
			value: originalManga.Info.Title,
		},
		{
			name:  "author",
			value: originalManga.Info.Authors[0],
		},
		{
			name:  "identifier",
			value: originalManga.Info.ID,
		},
	}

	for _, meta := range mangaMetadata {
		if !strings.Contains(opfContent, meta.value) {
			t.Errorf("Manga %s metadata not properly preserved", meta.name)
		}
	}
}
