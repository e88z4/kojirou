package epub

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestAddKoboNamespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "add namespace to html element",
			input:    `<!DOCTYPE html><html><head></head><body></body></html>`,
			expected: `xmlns:epub="http://www.kobo.com/ns/1.0"`,
		},
		{
			name:     "replace existing namespace",
			input:    `<!DOCTYPE html><html xmlns:epub="http://www.idpf.org/2007/ops"><head></head><body></body></html>`,
			expected: `xmlns:epub="http://www.kobo.com/ns/1.0"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply transformation
			modified := addKoboNamespaceToDoc(doc)
			if !modified {
				t.Errorf("Expected document to be modified")
			}

			// Render back to string
			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				t.Fatalf("Failed to render HTML: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Output doesn't contain expected namespace.\nExpected: %s\nGot: %s", tt.expected, output)
			}
		})
	}
}

func TestProcessTextNodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "add spans to paragraph text",
			input:    `<p>This is a paragraph.</p>`,
			expected: `<p><span class="koboSpan" id="kobo.`,
		},
		{
			name:     "handle multiple paragraphs",
			input:    `<div><p>First paragraph.</p><p>Second paragraph.</p></div>`,
			expected: `<span class="koboSpan" id="kobo.`,
		},
		{
			name:     "ignore script tags",
			input:    `<script>var x = 1;</script>`,
			expected: `<script>var x = 1;</script>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply transformation
			processTextNodes(doc)

			// Render back to string
			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				t.Fatalf("Failed to render HTML: %v", err)
			}

			output := buf.String()
			if tt.expected == `<script>var x = 1;</script>` {
				// Special case for script tags which should be unchanged
				if !strings.Contains(output, tt.expected) {
					t.Errorf("Output doesn't match expected.\nExpected to contain: %s\nGot: %s", tt.expected, output)
				}
			} else if !strings.Contains(output, tt.expected) {
				t.Errorf("Output doesn't contain expected spans.\nExpected to contain: %s\nGot: %s", tt.expected, output)
			}
		})
	}
}

func TestProcessImageElements(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectId       bool
		expectEpubType bool
	}{
		{
			name:           "add attributes to image",
			input:          `<img src="image.jpg" alt="An image">`,
			expectId:       true,
			expectEpubType: true,
		},
		{
			name:           "preserve existing id",
			input:          `<img id="existing" src="image.jpg" alt="An image">`,
			expectId:       true,
			expectEpubType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply transformation
			modified := processImageElements(doc)
			if !modified {
				t.Errorf("Expected document to be modified")
			}

			// Render back to string
			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				t.Fatalf("Failed to render HTML: %v", err)
			}

			output := buf.String()

			// Check for id attribute
			if tt.expectId && !strings.Contains(output, `id="`) {
				t.Errorf("Output doesn't contain id attribute.\nGot: %s", output)
			}

			// Check for epub:type attribute
			if tt.expectEpubType && !strings.Contains(output, `epub:type="kobo"`) {
				t.Errorf("Output doesn't contain epub:type attribute.\nGot: %s", output)
			}
		})
	}
}

func TestTransformHTMLFile(t *testing.T) {
	// Create a temporary HTML file
	htmlContent := `<!DOCTYPE html>
<html>
<head>
  <title>Test Document</title>
</head>
<body>
  <h1>Test Heading</h1>
  <p>This is a paragraph with some text.</p>
  <div>
    <p>Another paragraph in a div.</p>
    <img src="test.jpg" alt="Test Image">
  </div>
</body>
</html>`

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "kepub-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test HTML file
	htmlPath := filepath.Join(tempDir, "test.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to write test HTML: %v", err)
	}

	// Transform the HTML file
	if err := transformHTMLFile(htmlPath); err != nil {
		t.Fatalf("transformHTMLFile() error: %v", err)
	}

	// Read the transformed file
	transformedContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read transformed file: %v", err)
	}

	transformed := string(transformedContent)

	// Check for Kobo namespace
	if !strings.Contains(transformed, `xmlns:epub="http://www.kobo.com/ns/1.0"`) {
		t.Error("Transformed HTML is missing Kobo namespace")
	}

	// Check for Kobo spans
	if !strings.Contains(transformed, `<span class="koboSpan" id="kobo`) {
		t.Error("Transformed HTML is missing Kobo spans")
	}

	// Check for image attributes
	if !strings.Contains(transformed, `epub:type="kobo"`) {
		t.Error("Transformed HTML is missing epub:type attribute on images")
	}
}

func TestFindHTMLFiles(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "kepub-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some HTML files
	htmlFiles := []string{
		filepath.Join(tempDir, "chapter1.html"),
		filepath.Join(tempDir, "chapter2.xhtml"),
		filepath.Join(tempDir, "subfolder", "chapter3.html"),
	}

	// Create non-HTML files
	nonHtmlFiles := []string{
		filepath.Join(tempDir, "image.jpg"),
		filepath.Join(tempDir, "styles.css"),
	}

	// Create the directory structure
	if err := os.MkdirAll(filepath.Join(tempDir, "subfolder"), 0755); err != nil {
		t.Fatalf("Failed to create subfolder: %v", err)
	}

	// Create the files
	for _, file := range append(htmlFiles, nonHtmlFiles...) {
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Find HTML files
	foundFiles, err := findHTMLFiles(tempDir)
	if err != nil {
		t.Fatalf("findHTMLFiles() error: %v", err)
	}

	// Check if all HTML files were found
	if len(foundFiles) != len(htmlFiles) {
		t.Errorf("Expected %d HTML files, found %d", len(htmlFiles), len(foundFiles))
	}

	// Check that each HTML file was found
	foundMap := make(map[string]bool)
	for _, file := range foundFiles {
		foundMap[file] = true
	}

	for _, expectedFile := range htmlFiles {
		if !foundMap[expectedFile] {
			t.Errorf("Expected to find %s but it wasn't returned", expectedFile)
		}
	}
}

func TestFindOPFFiles(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "kepub-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create OPF file
	opfPath := filepath.Join(tempDir, "package.opf")
	if err := os.WriteFile(opfPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test OPF file: %v", err)
	}

	// Find OPF files
	foundFiles, err := findOPFFiles(tempDir)
	if err != nil {
		t.Fatalf("findOPFFiles() error: %v", err)
	}

	// Check if OPF file was found
	if len(foundFiles) != 1 {
		t.Errorf("Expected 1 OPF file, found %d", len(foundFiles))
	}

	if len(foundFiles) > 0 && foundFiles[0] != opfPath {
		t.Errorf("Expected to find %s but found %s", opfPath, foundFiles[0])
	}
}

func TestHTMLProcessor(t *testing.T) {
	tests := []struct {
		name          string
		htmlContent   string
		expectHeadTag bool
		expectBodyTag bool
		wantErr       bool
	}{
		{
			name:          "valid html document",
			htmlContent:   `<!DOCTYPE html><html><head><title>Test</title></head><body><p>Hello</p></body></html>`,
			expectHeadTag: true,
			expectBodyTag: true,
			wantErr:       false,
		},
		{
			name:          "no head tag",
			htmlContent:   `<!DOCTYPE html><html><body><p>Hello</p></body></html>`,
			expectHeadTag: false,
			expectBodyTag: true,
			wantErr:       false,
		},
		{
			name:          "no body tag",
			htmlContent:   `<!DOCTYPE html><html><head><title>Test</title></head></html>`,
			expectHeadTag: true,
			expectBodyTag: false,
			wantErr:       false,
		},
		{
			name:          "with existing kobo namespace",
			htmlContent:   `<!DOCTYPE html><html xmlns:kobo="http://www.kobo.com/ns/1.0"><head></head><body></body></html>`,
			expectHeadTag: true,
			expectBodyTag: true,
			wantErr:       false,
		},
		{
			name:          "malformed html",
			htmlContent:   `<not valid html`,
			expectHeadTag: false,
			expectBodyTag: false,
			// Note: html.Parse doesn't typically return errors, it tries to fix malformed HTML
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create processor
			processor, err := NewHTMLProcessor([]byte(tt.htmlContent))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHTMLProcessor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Verify key nodes were found as expected
			if (processor.headNode != nil) != tt.expectHeadTag {
				t.Errorf("Expected headNode to be %v, got %v", tt.expectHeadTag, processor.headNode != nil)
			}

			if (processor.bodyNode != nil) != tt.expectBodyTag {
				t.Errorf("Expected bodyNode to be %v, got %v", tt.expectBodyTag, processor.bodyNode != nil)
			}

			// Test namespace handling
			initialHasNamespace := processor.HasKoboNamespace()

			// Add namespace
			processor.AddKoboNamespace()

			// Verify namespace was added if it wasn't there already
			if !processor.HasKoboNamespace() {
				t.Errorf("Expected Kobo namespace to be added")
			}

			// Serialize document and check namespace in output
			output, err := processor.Serialize()
			if err != nil {
				t.Errorf("Failed to serialize HTML: %v", err)
				return
			}

			outputString := string(output)
			if !initialHasNamespace && !strings.Contains(outputString, "xmlns:kobo=\"http://www.kobo.com/ns/1.0\"") {
				t.Errorf("Kobo namespace not found in serialized output")
			}
		})
	}
}
