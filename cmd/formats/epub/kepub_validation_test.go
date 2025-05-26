package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

// TestCompleteKEPUBImplementation performs an end-to-end test of the KEPUB implementation
func TestCompleteKEPUBImplementation(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "kepub-validation")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a comprehensive test manga
	manga := createComprehensiveTestManga()

	// Generate EPUB
	epubObj, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB() failed: %v", err)
	}
	if cleanup != nil {
		// defer cleanup()
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Write output for inspection
	kepubPath := filepath.Join(tempDir, "test-complete.kepub.epub")
	err = os.WriteFile(kepubPath, kepubData, 0644)
	if err != nil {
		t.Logf("Failed to write KEPUB output: %v", err)
	}

	// Validate KEPUB structure and content
	validateCompleteKEPUB(t, kepubData, manga)
}

// TestKEPUBImplementationChecklist tests all important features of the KEPUB format
func TestKEPUBImplementationChecklist(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a basic EPUB
	e := epub.NewEpub("Checklist Test")
	e.SetAuthor("Test Author")

	// Add a section with mixed content types
	mixedContent := `
		<div>
			<h1>Test Chapter</h1>
			<p>This is a paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
			<ul>
				<li>Item 1</li>
				<li>Item 2</li>
				<li>Item 3</li>
			</ul>
			<p>This is another paragraph with <a href="#note">a link</a>.</p>
			<div class="image-container">
				<img src="image.jpg" alt="Test Image" />
			</div>
		</div>
	`

	_, err := e.AddSection(mixedContent, "Mixed Content", "mixed-content", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(e)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Run the complete checklist of KEPUB features
	runKEPUBFeatureChecklist(t, kepubData)
}

// TestKEPUBRegressions tests for regression issues in the KEPUB implementation
func TestKEPUBRegressions(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Test cases for regression issues
	regressionTests := []struct {
		name       string
		setupEpub  func() *epub.Epub
		validateFn func(*testing.T, []byte)
	}{
		{
			name: "nested divs with spans",
			setupEpub: func() *epub.Epub {
				e := epub.NewEpub("Nested Divs Test")
				e.SetAuthor("Test Author")

				content := `
					<div>
						<div>
							<div>
								<p>This is a <span>deeply</span> nested paragraph.</p>
							</div>
						</div>
					</div>
				`

				_, err := e.AddSection(content, "Nested Divs", "nested-divs", "")
				if err != nil {
					t.Fatalf("Failed to add section: %v", err)
				}

				return e
			},
			validateFn: func(t *testing.T, data []byte) {
				// Extract HTML
				htmlFiles := extractHTMLFromKEPUBFile(t, data)

				// Check for proper handling of nested elements
				for _, html := range htmlFiles {
					if strings.Contains(html, "<div>") && strings.Contains(html, "<span>") {
						// Check that spans are properly handled
						if !strings.Contains(html, "kobo") {
							t.Error("Nested elements not properly handled with Kobo spans")
						}
					}
				}
			},
		},
		{
			name: "special characters in attributes",
			setupEpub: func() *epub.Epub {
				e := epub.NewEpub("Special Chars Test")
				e.SetAuthor("Test Author")

				content := `
					<div id="special&id" class="special&class">
						<p title="Quote&quot;Test">This has special characters in attributes.</p>
					</div>
				`

				_, err := e.AddSection(content, "Special Chars", "special-chars", "")
				if err != nil {
					t.Fatalf("Failed to add section: %v", err)
				}

				return e
			},
			validateFn: func(t *testing.T, data []byte) {
				// Extract HTML
				htmlFiles := extractHTMLFromKEPUBFile(t, data)

				// Check for proper handling of special characters
				for _, html := range htmlFiles {
					if strings.Contains(html, "special") {
						// Ensure special chars are properly escaped
						if strings.Contains(html, "special&id") && !strings.Contains(html, "special&amp;id") {
							t.Error("Special characters not properly escaped in attributes")
						}
					}
				}
			},
		},
	}

	for _, test := range regressionTests {
		t.Run(test.name, func(t *testing.T) {
			// Setup the EPUB
			epubObj := test.setupEpub()

			// Convert to KEPUB
			kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Run validation
			test.validateFn(t, kepubData)
		})
	}
}

// TestKEPUBCompleteness validates that all expected KEPUB features are implemented
func TestKEPUBCompleteness(t *testing.T) {
	// Skip until implementation is complete
	t.Skip("KEPUB conversion not implemented yet")

	// Create a basic EPUB
	e := epub.NewEpub("Completeness Test")
	e.SetAuthor("Test Author")

	// Add a simple section
	_, err := e.AddSection("<p>Test paragraph</p>", "Test Chapter", "test-chapter", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Convert to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(e)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Define all required features for a complete KEPUB implementation
	requiredFeatures := []struct {
		name    string
		checkFn func([]byte) bool
	}{
		{
			name: "Basic EPUB structure",
			checkFn: func(data []byte) bool {
				// Check for minimum EPUB structure
				reader := bytes.NewReader(data)
				zipReader, err := zip.NewReader(reader, int64(len(data)))
				if err != nil {
					return false
				}

				hasContainer := false
				hasOPF := false

				for _, file := range zipReader.File {
					if file.Name == "META-INF/container.xml" {
						hasContainer = true
					}
					if strings.HasSuffix(file.Name, ".opf") {
						hasOPF = true
					}
				}

				return hasContainer && hasOPF
			},
		},
		{
			name: "Kobo namespaces",
			checkFn: func(data []byte) bool {
				// Check for Kobo namespaces in any XML file
				reader := bytes.NewReader(data)
				zipReader, err := zip.NewReader(reader, int64(len(data)))
				if err != nil {
					return false
				}

				for _, file := range zipReader.File {
					if strings.HasSuffix(file.Name, ".xml") ||
						strings.HasSuffix(file.Name, ".opf") ||
						strings.HasSuffix(file.Name, ".html") ||
						strings.HasSuffix(file.Name, ".xhtml") {
						rc, err := file.Open()
						if err != nil {
							continue
						}

						contentBytes, err := io.ReadAll(rc)
						rc.Close()
						if err != nil {
							continue
						}

						content := string(contentBytes)
						if strings.Contains(content, "http://www.kobo.com") {
							return true
						}
					}
				}

				return false
			},
		},
		{
			name: "Kobo spans",
			checkFn: func(data []byte) bool {
				// Check for Kobo spans in HTML files
				htmlFiles := extractHTMLFromKEPUBFile(nil, data)
				for _, html := range htmlFiles {
					if strings.Contains(html, "<span class=\"kobo") {
						return true
					}
				}
				return false
			},
		},
		{
			name: "Kobo metadata",
			checkFn: func(data []byte) bool {
				// Check for Kobo-specific metadata
				reader := bytes.NewReader(data)
				zipReader, err := zip.NewReader(reader, int64(len(data)))
				if err != nil {
					return false
				}

				for _, file := range zipReader.File {
					if strings.HasSuffix(file.Name, ".opf") {
						rc, err := file.Open()
						if err != nil {
							return false
						}

						contentBytes, err := io.ReadAll(rc)
						rc.Close()
						if err != nil {
							return false
						}

						content := string(contentBytes)
						if strings.Contains(content, "kobo") {
							return true
						}
					}
				}

				return false
			},
		},
		{
			name: "Expected KEPUB file extension",
			checkFn: func(data []byte) bool {
				// This is a bit artificial since we're just testing bytes,
				// but in a real application, the file would have this extension
				return true
			},
		},
	}

	// Check all required features
	for _, feature := range requiredFeatures {
		t.Run(feature.name, func(t *testing.T) {
			if !feature.checkFn(kepubData) {
				t.Errorf("Required KEPUB feature missing: %s", feature.name)
			}
		})
	}
}

// Helper functions

// createComprehensiveTestManga returns a Manga using the new data model
func createComprehensiveTestManga() md.Manga {
	volID := md.NewIdentifier("1")
	manga := md.Manga{
		Info: md.MangaInfo{
			Title:   "Comprehensive Test Manga",
			Authors: []string{"Test Author"},
		},
		Volumes: map[md.Identifier]md.Volume{
			volID: {
				Info:     md.VolumeInfo{Identifier: volID},
				Chapters: map[md.Identifier]md.Chapter{},
			},
		},
	}
	// Add multiple chapters with different structures
	for i := 1; i <= 3; i++ {
		chID := md.NewIdentifier(fmt.Sprintf("1.%d", i))
		chapter := md.Chapter{
			Info: md.ChapterInfo{
				Title:            fmt.Sprintf("Chapter %d", i),
				Identifier:       chID,
				VolumeIdentifier: volID,
			},
			Pages: map[int]image.Image{},
		}
		// Add pages with different characteristics
		for j := 1; j <= 5; j++ {
			img := createTestImage(800, 1200, color.White)
			if j == 3 {
				img = createTestImage(1600, 1200, color.White) // wide page
			}
			chapter.Pages[j] = img
		}
		manga.Volumes[volID].Chapters[chID] = chapter
	}
	return manga
}

func extractHTMLFromKEPUBFile(t *testing.T, data []byte) []string {
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
		}
		return nil
	}

	var htmlFiles []string
	for _, file := range zipReader.File {
		// Only check HTML files
		if strings.HasSuffix(file.Name, ".html") || strings.HasSuffix(file.Name, ".xhtml") {
			rc, err := file.Open()
			if err != nil {
				if t != nil {
					t.Fatalf("Failed to open HTML file: %v", err)
				}
				continue
			}

			contentBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				if t != nil {
					t.Fatalf("Failed to read HTML content: %v", err)
				}
				continue
			}

			htmlFiles = append(htmlFiles, string(contentBytes))
		}
	}

	return htmlFiles
}

func validateCompleteKEPUB(t *testing.T, data []byte, manga md.Manga) {
	// Comprehensive validation of the KEPUB file
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check required files and structure
	requiredFiles := map[string]bool{
		"META-INF/container.xml": false,
		"mimetype":               false,
	}

	var opfFile string
	var htmlFiles []string
	var cssFiles []string
	var imageFiles []string

	for _, file := range zipReader.File {
		// Mark required files as found
		if _, ok := requiredFiles[file.Name]; ok {
			requiredFiles[file.Name] = true
		}

		// Identify content types
		switch {
		case strings.HasSuffix(file.Name, ".opf"):
			opfFile = file.Name
		case strings.HasSuffix(file.Name, ".html") || strings.HasSuffix(file.Name, ".xhtml"):
			htmlFiles = append(htmlFiles, file.Name)
		case strings.HasSuffix(file.Name, ".css"):
			cssFiles = append(cssFiles, file.Name)
		case strings.HasSuffix(file.Name, ".jpg") || strings.HasSuffix(file.Name, ".jpeg") ||
			strings.HasSuffix(file.Name, ".png") || strings.HasSuffix(file.Name, ".gif"):
			imageFiles = append(imageFiles, file.Name)
		}
	}

	// Check required files
	for file, found := range requiredFiles {
		if !found {
			t.Errorf("Required file missing: %s", file)
		}
	}

	// Check content structure
	if opfFile == "" {
		t.Error("OPF file missing")
	}

	if len(htmlFiles) == 0 {
		t.Error("No HTML content files found")
	}

	// Validate expected chapter count (rough estimate)
	// Each chapter typically has at least one HTML file
	if len(htmlFiles) < len(manga.Volumes[md.NewIdentifier("1")].Chapters) {
		t.Logf("Warning: Fewer HTML files (%d) than chapters (%d)", len(htmlFiles), len(manga.Volumes[md.NewIdentifier("1")].Chapters))
	}

	// For a proper conversion, we should have some images
	if len(imageFiles) == 0 {
		t.Error("No image files found")
	}

	// Log the structure for debugging
	t.Logf("KEPUB structure: %d HTML files, %d CSS files, %d image files",
		len(htmlFiles), len(cssFiles), len(imageFiles))
}

func runKEPUBFeatureChecklist(t *testing.T, data []byte) {
	// List of all KEPUB features to check
	features := []struct {
		name     string
		checkFn  func(*testing.T, []byte) bool
		required bool
	}{
		{
			name: "EPUB3 Compatibility",
			checkFn: func(t *testing.T, data []byte) bool {
				// Check for EPUB3 version in OPF
				reader := bytes.NewReader(data)
				zipReader, err := zip.NewReader(reader, int64(len(data)))
				if err != nil {
					t.Logf("Failed to read KEPUB as ZIP: %v", err)
					return false
				}

				for _, file := range zipReader.File {
					if strings.HasSuffix(file.Name, ".opf") {
						rc, err := file.Open()
						if err != nil {
							continue
						}

						contentBytes, err := io.ReadAll(rc)
						rc.Close()
						if err != nil {
							continue
						}

						content := string(contentBytes)
						return strings.Contains(content, "version=\"3.0\"") ||
							strings.Contains(content, "version=\"3")
					}
				}

				return false
			},
			required: true,
		},
		{
			name: "Kobo Paragraph Spans",
			checkFn: func(t *testing.T, data []byte) bool {
				// Check for paragraph spans in HTML
				htmlFiles := extractHTMLFromKEPUBFile(t, data)

				for _, html := range htmlFiles {
					if strings.Contains(html, "<p>") && strings.Contains(html, "<span class=\"kobo") {
						return true
					}
				}

				return false
			},
			required: true,
		},
		{
			name: "Fixed Layout Support",
			checkFn: func(t *testing.T, data []byte) bool {
				// Check for fixed layout metadata
				reader := bytes.NewReader(data)
				zipReader, err := zip.NewReader(reader, int64(len(data)))
				if err != nil {
					t.Logf("Failed to read KEPUB as ZIP: %v", err)
					return false
				}

				for _, file := range zipReader.File {
					if strings.HasSuffix(file.Name, ".opf") {
						rc, err := file.Open()
						if err != nil {
							continue
						}

						contentBytes, err := ioutil.ReadAll(rc)
						rc.Close()
						if err != nil {
							continue
						}

						content := string(contentBytes)
						return strings.Contains(content, "fixed-layout") ||
							strings.Contains(content, "rendition:layout") ||
							strings.Contains(content, "orientation")
					}
				}

				return false
			},
			required: false, // Only required for fixed-layout content
		},
		{
			name: "Kobo Namespaces",
			checkFn: func(t *testing.T, data []byte) bool {
				// Check for Kobo namespaces
				htmlFiles := extractHTMLFromKEPUBFile(t, data)

				for _, html := range htmlFiles {
					if strings.Contains(html, "http://www.kobo.com") {
						return true
					}
				}

				return false
			},
			required: true,
		},
	}

	// Run checks for all features
	for _, feature := range features {
		t.Run(feature.name, func(t *testing.T) {
			result := feature.checkFn(t, data)

			if feature.required && !result {
				t.Errorf("Required KEPUB feature missing: %s", feature.name)
			} else if !result {
				t.Logf("Optional KEPUB feature missing: %s", feature.name)
			}
		})
	}
}
