// Package epub provides functions for converting EPUB and KEPUB formats
package epub

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// KoboHTMLProcessor processes HTML files for the Kobo KEPUB format
type KoboHTMLProcessor struct {
	Doc           *html.Node
	SpanIDCounter int
}

// NewKoboHTMLProcessor creates a new HTML processor from HTML content
func NewKoboHTMLProcessor(content []byte) (*KoboHTMLProcessor, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return &KoboHTMLProcessor{
		Doc:           doc,
		SpanIDCounter: 1,
	}, nil
}

// GenerateSpanID generates a unique ID for kobo spans
func (p *KoboHTMLProcessor) GenerateSpanID() string {
	id := fmt.Sprintf("%d", p.SpanIDCounter)
	p.SpanIDCounter++
	return id
}

// ProcessTextNodes processes all text nodes in <p> and <div> elements only
func (p *KoboHTMLProcessor) ProcessTextNodes() {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "p" || n.Data == "div") {
			processTextNodesInElement(n, p)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(p.Doc)
}

// processTextNodesInElement wraps text nodes in Kobo spans for a given element
func processTextNodesInElement(node *html.Node, p *KoboHTMLProcessor) {
	var textNodes []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
			textNodes = append(textNodes, c)
		}
	}
	for _, textNode := range textNodes {
		text := textNode.Data
		span := &html.Node{
			Type: html.ElementNode,
			Data: "span",
			Attr: []html.Attribute{
				{Key: "class", Val: "koboSpan"},
				{Key: "id", Val: "kobo-span-" + p.GenerateSpanID()},
			},
		}
		newText := &html.Node{
			Type: html.TextNode,
			Data: text,
		}
		span.AppendChild(newText)
		node.InsertBefore(span, textNode)
		node.RemoveChild(textNode)
	}
}

// ProcessImageElements adds Kobo-specific attributes to image elements
func (p *KoboHTMLProcessor) ProcessImageElements() {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			// Add class="kobo-image" if not present
			hasClass := false
			for i, attr := range n.Attr {
				if attr.Key == "class" {
					if !strings.Contains(attr.Val, "kobo-image") {
						n.Attr[i].Val = attr.Val + " kobo-image"
					}
					hasClass = true
					break
				}
			}
			if !hasClass {
				n.Attr = append(n.Attr, html.Attribute{Key: "class", Val: "kobo-image"})
			}
			// Add epub:type="kobo" if not present
			hasEpubType := false
			for _, attr := range n.Attr {
				if attr.Key == "epub:type" && attr.Val == "kobo" {
					hasEpubType = true
					break
				}
			}
			if !hasEpubType {
				n.Attr = append(n.Attr, html.Attribute{Key: "epub:type", Val: "kobo"})
			}
			// Add id if not present
			hasID := false
			for _, attr := range n.Attr {
				if attr.Key == "id" {
					hasID = true
					break
				}
			}
			if !hasID {
				id := fmt.Sprintf("kobo_img_%d", rand.Intn(10000))
				n.Attr = append(n.Attr, html.Attribute{Key: "id", Val: id})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(p.Doc)
}

// AddKoboNamespace adds the Kobo namespace to the HTML element
func (p *KoboHTMLProcessor) AddKoboNamespace() bool {
	// Find the HTML node
	var htmlNode *html.Node
	var findHTML func(*html.Node) *html.Node
	findHTML = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode && n.Data == "html" {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if found := findHTML(c); found != nil {
				return found
			}
		}
		return nil
	}

	htmlNode = findHTML(p.Doc)
	if htmlNode == nil {
		return false
	}

	// Add or replace the epub namespace
	modified := false
	namespaceAttr := "xmlns:epub"
	namespaceVal := "http://www.kobo.com/ns/1.0"

	for i, attr := range htmlNode.Attr {
		if attr.Key == namespaceAttr {
			if attr.Val != namespaceVal {
				htmlNode.Attr[i].Val = namespaceVal
				modified = true
			}
			return modified
		}
	}

	// Add namespace if not found
	htmlNode.Attr = append(htmlNode.Attr, html.Attribute{
		Key: namespaceAttr,
		Val: namespaceVal,
	})
	return true
}

// RenderToString renders the HTML document to a string
func (p *KoboHTMLProcessor) RenderToString() (string, error) {
	var buf bytes.Buffer
	if err := html.Render(&buf, p.Doc); err != nil {
		return "", fmt.Errorf("failed to render HTML: %w", err)
	}
	return buf.String(), nil
}

// TransformHTMLFile processes an HTML file for Kobo compatibility
func TransformHTMLFile(htmlPath string) error {
	// Read file
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}

	// Parse HTML
	processor, err := NewKoboHTMLProcessor(content)
	if err != nil {
		return err
	}

	// Apply transformations
	processor.AddKoboNamespace()
	processor.ProcessTextNodes()
	processor.ProcessImageElements()

	// Write back
	rendered, err := processor.RenderToString()
	if err != nil {
		return err
	}

	if err := os.WriteFile(htmlPath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// FindHTMLFiles finds all HTML files in a directory
func FindHTMLFiles(rootDir string) ([]string, error) {
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
	if err != nil {
		return nil, fmt.Errorf("failed to find HTML files: %w", err)
	}
	return htmlFiles, nil
}
