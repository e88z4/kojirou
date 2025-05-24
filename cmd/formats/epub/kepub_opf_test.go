package epub

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestUpdateOPFMetadata(t *testing.T) {
	// Create a temporary OPF file for testing
	tempDir, err := createTempDir(t, "kepub-opf-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer cleanupTempDir(t, tempDir)

	// Create a sample OPF file
	opfPath := filepath.Join(tempDir, "package.opf")
	testOPF := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Manga</dc:title>
    <dc:language>en</dc:language>
    <dc:identifier id="uid">test-manga-id</dc:identifier>
    <meta property="dcterms:modified">2022-01-01T12:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="page1" href="page1.xhtml" media-type="application/xhtml+xml"/>
    <item id="image1" href="images/img1.jpg" media-type="image/jpeg"/>
  </manifest>
  <spine>
    <itemref idref="page1"/>
  </spine>
</package>`

	if err := os.WriteFile(opfPath, []byte(testOPF), 0644); err != nil {
		t.Fatalf("Failed to write test OPF file: %v", err)
	}

	// Update the OPF metadata
	if err := updateOPFMetadata(opfPath, map[string]string{}); err != nil {
		t.Fatalf("Failed to update OPF metadata: %v", err)
	}

	// Read the updated OPF file
	updatedContent, err := os.ReadFile(opfPath)
	if err != nil {
		t.Fatalf("Failed to read updated OPF file: %v", err)
	}

	// Parse the updated OPF
	doc, err := parseOPF(updatedContent)
	if err != nil {
		t.Fatalf("Failed to parse updated OPF: %v", err)
	}

	// Validate the Kobo metadata
	t.Run("Check added Kobo metadata", func(t *testing.T) {
		metaProperties := extractMetaProperties(doc)

		// Check required Kobo metadata
		requiredProps := []string{
			"kobo:content-type",
			"kobo:epub-version",
			"dcterms:modified",
			"rendition:flow",
			"rendition:layout",
			"rendition:orientation",
			"rendition:spread",
		}

		for _, prop := range requiredProps {
			if _, exists := metaProperties[prop]; !exists {
				t.Errorf("Required metadata property not found: %s", prop)
			}
		}

		// Check specific values
		if metaProperties["kobo:content-type"] != "comic" {
			t.Errorf("Expected kobo:content-type to be 'comic', got '%s'", metaProperties["kobo:content-type"])
		}

		if metaProperties["kobo:epub-version"] != "3.0" {
			t.Errorf("Expected kobo:epub-version to be '3.0', got '%s'", metaProperties["kobo:epub-version"])
		}

		if metaProperties["rendition:layout"] != "pre-paginated" {
			t.Errorf("Expected rendition:layout to be 'pre-paginated', got '%s'", metaProperties["rendition:layout"])
		}
	})

	// Validate the reading direction
	t.Run("Check reading direction metadata", func(t *testing.T) {
		spineNode := findNode(doc, "spine")
		if spineNode == nil {
			t.Fatalf("Spine element not found in updated OPF")
		}

		// Check page-progression-direction attribute
		var direction string
		for _, attr := range spineNode.Attr {
			if attr.Key == "page-progression-direction" {
				direction = attr.Val
				break
			}
		}

		if direction != "rtl" {
			t.Errorf("Expected page-progression-direction to be 'rtl', got '%s'", direction)
		}
	})
}

// extractMetaProperties extracts all meta property values from OPF
func extractMetaProperties(doc *html.Node) map[string]string {
	props := make(map[string]string)
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var property, content string
			for _, attr := range n.Attr {
				if attr.Key == "property" {
					property = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}
			if property != "" && content != "" {
				props[property] = content
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return props
}

func TestOPFParsing(t *testing.T) {
	testXML := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata>
    <dc:title>Test Title</dc:title>
  </metadata>
</package>`

	doc, err := parseOPF([]byte(testXML))
	if err != nil {
		t.Fatalf("Failed to parse OPF: %v", err)
	}

	packageNode := findNode(doc, "package")
	if packageNode == nil {
		t.Error("Failed to find package node")
	}

	metadataNode := findNode(doc, "metadata")
	if metadataNode == nil {
		t.Error("Failed to find metadata node")
	}
}

func TestSaveOPF(t *testing.T) {
	// Parse a simple XML
	doc, err := parseOPF([]byte("<package><metadata></metadata></package>"))
	if err != nil {
		t.Fatalf("Failed to parse test OPF: %v", err)
	}

	// Add an element
	metadataNode := findNode(doc, "metadata")
	addMetaElement(metadataNode, "meta", map[string]string{
		"property": "test:property",
		"content":  "test-value",
	})

	// Create a temp file
	tempDir, err := createTempDir(t, "kepub-opf-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer cleanupTempDir(t, tempDir)

	opfPath := filepath.Join(tempDir, "test.opf")

	// Save the OPF
	if err := saveOPF(doc, opfPath); err != nil {
		t.Fatalf("Failed to save OPF: %v", err)
	}

	// Read it back
	content, err := os.ReadFile(opfPath)
	if err != nil {
		t.Fatalf("Failed to read saved OPF: %v", err)
	}

	// Verify content
	if !bytes.Contains(content, []byte("test:property")) || !bytes.Contains(content, []byte("test-value")) {
		t.Error("Saved OPF does not contain expected content")
	}
}

func TestAddMetaElement(t *testing.T) {
	// Create a simple document with metadata
	docStr := "<package><metadata></metadata></package>"
	doc, _ := html.Parse(strings.NewReader(docStr))

	// Find metadata node
	metadataNode := findNode(doc, "metadata")
	if metadataNode == nil {
		t.Fatal("Failed to find metadata node")
	}

	// Add a meta element
	addMetaElement(metadataNode, "meta", map[string]string{
		"property": "test:property",
		"content":  "test-value",
	})

	// Verify the element was added
	var buffer bytes.Buffer
	if err := html.Render(&buffer, doc); err != nil {
		t.Fatalf("Failed to render HTML: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "test:property") || !strings.Contains(output, "test-value") {
		t.Errorf("Meta element was not added correctly: %s", output)
	}
}

func TestSetAttribute(t *testing.T) {
	// Create a test node
	node := &html.Node{
		Type: html.ElementNode,
		Data: "test",
		Attr: []html.Attribute{
			{Key: "existing", Val: "value"},
		},
	}

	// Update existing attribute
	setAttribute(node, "existing", "new-value")

	// Add new attribute
	setAttribute(node, "new", "value")

	// Verify changes
	foundExisting := false
	foundNew := false

	for _, attr := range node.Attr {
		if attr.Key == "existing" {
			foundExisting = true
			if attr.Val != "new-value" {
				t.Errorf("Expected value 'new-value' for 'existing', got '%s'", attr.Val)
			}
		}
		if attr.Key == "new" {
			foundNew = true
			if attr.Val != "value" {
				t.Errorf("Expected value 'value' for 'new', got '%s'", attr.Val)
			}
		}
	}

	if !foundExisting {
		t.Error("Existing attribute not found")
	}
	if !foundNew {
		t.Error("New attribute not found")
	}
}
