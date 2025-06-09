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
		name        string
		setupEpub   func() *epub.Epub
		seriesTitle string
		seriesIndex float64
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
			seriesTitle: "",
			seriesIndex: 0,
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
				return e
			},
			seriesTitle: "",
			seriesIndex: 0,
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
			seriesTitle: "",
			seriesIndex: 0,
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
				defer func() {
					if cleanup != nil {
						cleanup()
					}
				}()
				return epubObj
			},
			seriesTitle: "Manga Metadata Test",
			seriesIndex: 1,
		},
		{
			name: "series metadata",
			setupEpub: func() *epub.Epub {
				// Create a manga object with series info
				manga := md.Manga{
					Info: md.MangaInfo{
						Title:   "Test Series Volume 1",
						Authors: []string{"Test Author"},
						ID:      "test-series-v1",
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
				defer func() {
					if cleanup != nil {
						cleanup()
					}
				}()
				return epubObj
			},
			seriesTitle: "Test Series",
			seriesIndex: 1.0,
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
			kepubData, err := kepubconv.ConvertToKEPUB(epubObj, tc.seriesTitle, tc.seriesIndex)
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
	kepubData, err := kepubconv.ConvertToKEPUB(e, "", 0)
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
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj, manga.Info.Title, 1.0)
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
	requiredExtensions := []string{
		"meta property=\"kobo:content-type\"",
		"meta property=\"kobo:epub-version\"",
		"meta property=\"rendition:layout\"",
		"meta property=\"rendition:orientation\"",
		"meta property=\"rendition:spread\"",
		"meta property=\"rendition:flow\"",
		"meta property=\"dcterms:modified\"",
		"meta property=\"page-progression-direction\"",
	}

	optionalExtensions := []string{
		"meta name=\"calibre:series\"",
		"meta name=\"calibre:series_index\"",
	}

	// Check for required extensions
	for _, ext := range requiredExtensions {
		if !strings.Contains(opfContent, ext) {
			t.Errorf("Missing required Kobo extension: %s", ext)
		}
	}

	// Log if optional extensions are found
	for _, ext := range optionalExtensions {
		if strings.Contains(opfContent, ext) {
			t.Logf("Found optional extension: %s", ext)
		}
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
		{
			name:  "series",
			value: originalManga.Info.Title,
		},
	}

	for _, meta := range mangaMetadata {
		if !strings.Contains(opfContent, meta.value) {
			t.Errorf("Manga %s metadata not properly preserved", meta.name)
		}
	}

	// Check for series index metadata (should be "1.0" for most manga tests)
	if !strings.Contains(opfContent, "meta name=\"calibre:series_index\" content=\"1.0\"") {
		t.Error("Series index metadata not found or incorrect")
	}
}
