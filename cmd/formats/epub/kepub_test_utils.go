package epub

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// HTMLProcessor handles parsing and modifying HTML content
type HTMLProcessor struct {
	doc      *html.Node
	headNode *html.Node
	bodyNode *html.Node
}

// NewHTMLProcessor creates a new HTMLProcessor instance from HTML content
func NewHTMLProcessor(content []byte) (*HTMLProcessor, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find head and body nodes
	processor := &HTMLProcessor{doc: doc}

	// Find the HTML element and its head/body children
	var findHeadAndBody func(*html.Node)
	findHeadAndBody = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "head" {
				processor.headNode = n
			} else if n.Data == "body" {
				processor.bodyNode = n
			}
		}

		// Continue search in children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findHeadAndBody(c)
		}
	}

	findHeadAndBody(doc)
	return processor, nil
}

// GetDocument returns the HTML document
func (p *HTMLProcessor) GetDocument() *html.Node {
	return p.doc
}

// Serialize converts the HTML document back to bytes
func (p *HTMLProcessor) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	err := html.Render(&buf, p.doc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize HTML: %w", err)
	}
	return buf.Bytes(), nil
}

// HasKoboNamespace checks if the HTML element has the Kobo namespace attribute
func (p *HTMLProcessor) HasKoboNamespace() bool {
	// Find the HTML element
	var htmlNode *html.Node
	var findHTMLElement func(*html.Node)
	findHTMLElement = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "html" {
			htmlNode = n
			return
		}
		for c := n.FirstChild; c != nil && htmlNode == nil; c = c.NextSibling {
			findHTMLElement(c)
		}
	}
	findHTMLElement(p.doc)

	if htmlNode == nil {
		return false
	}

	// Check for Kobo namespace
	for _, attr := range htmlNode.Attr {
		if attr.Key == "xmlns:epub" && attr.Val == "http://www.kobo.com/ns/1.0" {
			return true
		}
	}

	return false
}

// AddKoboNamespace adds the Kobo namespace to the HTML element
func (p *HTMLProcessor) AddKoboNamespace() {
	// Find the HTML element
	var htmlNode *html.Node
	var findHTMLElement func(*html.Node)
	findHTMLElement = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "html" {
			htmlNode = n
			return
		}
		for c := n.FirstChild; c != nil && htmlNode == nil; c = c.NextSibling {
			findHTMLElement(c)
		}
	}
	findHTMLElement(p.doc)

	if htmlNode == nil {
		return
	}

	// Add or replace Kobo namespace
	koboNS := "http://www.kobo.com/ns/1.0"
	nsExists := false

	for i, attr := range htmlNode.Attr {
		if attr.Key == "xmlns:epub" {
			htmlNode.Attr[i].Val = koboNS
			nsExists = true
			break
		}
	}

	if !nsExists {
		htmlNode.Attr = append(htmlNode.Attr, html.Attribute{
			Key: "xmlns:epub",
			Val: koboNS,
		})
	}
}

// addKoboNamespaceToDoc adds the Kobo namespace to the HTML document root
// Returns true if a modification was made, false otherwise
func addKoboNamespaceToDoc(doc *html.Node) bool {
	// Find the HTML element
	var htmlNode *html.Node
	var findHTMLElement func(*html.Node)
	findHTMLElement = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "html" {
			htmlNode = n
			return
		}
		for c := n.FirstChild; c != nil && htmlNode == nil; c = c.NextSibling {
			findHTMLElement(c)
		}
	}
	findHTMLElement(doc)

	if htmlNode == nil {
		return false
	}

	// Check if Kobo namespace already exists
	modified := false
	namespaceExists := false
	for i, attr := range htmlNode.Attr {
		if attr.Key == "xmlns:epub" {
			if attr.Val != "http://www.kobo.com/ns/1.0" {
				htmlNode.Attr[i].Val = "http://www.kobo.com/ns/1.0"
				modified = true
			}
			namespaceExists = true
			break
		}
	}

	// Add namespace if it doesn't exist
	if !namespaceExists {
		htmlNode.Attr = append(htmlNode.Attr, html.Attribute{
			Key: "xmlns:epub",
			Val: "http://www.kobo.com/ns/1.0",
		})
		modified = true
	}

	return modified
}

// processTextNodes processes text nodes in the HTML document, adding Kobo-specific spans
func processTextNodes(doc *html.Node) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "p" || n.Data == "div") {
			processTextNodesForKobo(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
}

// processTextNodesForKobo is a test-local copy for test helpers
func processTextNodesForKobo(n *html.Node) {
	// Collect text nodes
	var textNodes []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
			textNodes = append(textNodes, c)
		}
	}

	// Replace each text node with a span-wrapped version
	for _, textNode := range textNodes {
		text := textNode.Data

		// Create span element
		span := &html.Node{
			Type: html.ElementNode,
			Data: "span",
			Attr: []html.Attribute{
				{Key: "class", Val: "koboSpan"},
				{Key: "id", Val: "kobo-span-test"}, // test id
			},
		}

		// Create new text node with the same content
		newText := &html.Node{
			Type: html.TextNode,
			Data: text,
		}

		// Add text node to span
		span.AppendChild(newText)

		// Replace original text node with span
		n.InsertBefore(span, textNode)
		n.RemoveChild(textNode)
	}
}

// processImageElements adds Kobo-specific attributes to image elements
func processImageElements(doc *html.Node) bool {
	modified := false
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			// Check if we need to add the kobo class
			hasClass := false
			for i, attr := range n.Attr {
				if attr.Key == "class" {
					if !strings.Contains(attr.Val, "kobo") {
						n.Attr[i].Val = attr.Val + " kobo-image"
						modified = true
					}
					hasClass = true
					break
				}
			}
			if !hasClass {
				n.Attr = append(n.Attr, html.Attribute{Key: "class", Val: "kobo-image"})
				modified = true
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return modified
}

// transformHTMLFile applies Kobo-specific transformations to an HTML file
func transformHTMLFile(htmlPath string) error {
	content, err := ioutil.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}

	processor, err := NewHTMLProcessor(content)
	if err != nil {
		return fmt.Errorf("failed to create HTML processor: %w", err)
	}

	doc := processor.GetDocument()
	processTextNodes(doc)
	processImageElements(doc)

	modified, err := processor.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize HTML: %w", err)
	}

	err = ioutil.WriteFile(htmlPath, modified, 0644)
	if err != nil {
		return fmt.Errorf("failed to write modified HTML: %w", err)
	}

	return nil
}

// findHTMLFiles recursively finds all HTML files in a directory
func findHTMLFiles(rootDir string) ([]string, error) {
	var htmlFiles []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".xhtml")) {
			htmlFiles = append(htmlFiles, path)
		}
		return nil
	})
	return htmlFiles, err
}
