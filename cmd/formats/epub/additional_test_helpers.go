package epub

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"

	md "github.com/leotaku/kojirou/mangadex"
)

// createTestImage creates a test image with the specified dimensions and background color
func createTestImage(width, height int, bgColor color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}
	return img
}

// createEmptyVolumeManga creates a manga with volumes but no chapters
func createEmptyVolumeManga() md.Manga {
	manga := md.Manga{
		Info: md.MangaInfo{
			Title:   "Test Manga",
			ID:      "test-manga-id",
			Authors: []string{"Test Author"},
		},
		Volumes: map[md.Identifier]md.Volume{},
	}

	// Add an empty volume with no chapters
	volID := md.NewIdentifier("vol1")
	manga.Volumes[volID] = md.Volume{
		Info: md.VolumeInfo{
			Identifier: volID,
		},
		Chapters: map[md.Identifier]md.Chapter{},
	}

	return manga
}

// createNoTitleManga creates a manga with no title
func createNoTitleManga() md.Manga {
	manga := md.Manga{
		Info: md.MangaInfo{
			Title:   "",
			ID:      "test-manga-id",
			Authors: []string{"Test Author"},
		},
		Volumes: map[md.Identifier]md.Volume{},
	}
	return manga
}

// createDetailedWidePageManga creates a manga with wide pages for testing page splitting
// Different name to avoid conflicts with createWidePageTestManga in other files
func createDetailedWidePageManga() md.Manga {
	manga := md.Manga{
		Info: md.MangaInfo{
			Title:   "Test Manga",
			ID:      "test-manga-id",
			Authors: []string{"Test Author"},
		},
		Volumes: map[md.Identifier]md.Volume{},
	}

	// Create volume 1
	vol1ID := md.NewIdentifier("vol1")
	vol1 := md.Volume{
		Info: md.VolumeInfo{
			Identifier: vol1ID,
		},
		Chapters: map[md.Identifier]md.Chapter{},
	}

	// Add chapter to volume 1
	chap1ID := md.NewIdentifier("ch1")
	chap1 := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chap1ID,
			Title:            "Chapter 1",
			VolumeIdentifier: vol1ID,
		},
		Pages: map[int]image.Image{
			0: createTestImage(1000, 1500, color.White), // Normal portrait
			1: createTestImage(2000, 1000, color.White), // Wide landscape (1:2 ratio)
			2: createTestImage(3000, 1000, color.White), // Very wide landscape (1:3 ratio)
			3: createTestImage(1000, 1500, color.White), // Normal portrait
		},
	}
	vol1.Chapters[chap1ID] = chap1

	// Create volume 2
	vol2ID := md.NewIdentifier("vol2")
	vol2 := md.Volume{
		Info: md.VolumeInfo{
			Identifier: vol2ID,
		},
		Chapters: map[md.Identifier]md.Chapter{},
	}

	// Add chapter to volume 2
	chap2ID := md.NewIdentifier("ch2")
	chap2 := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chap2ID,
			Title:            "Chapter 2",
			VolumeIdentifier: vol2ID,
		},
		Pages: map[int]image.Image{
			0: createTestImage(1000, 1500, color.White), // Normal portrait
			1: createTestImage(2000, 1200, color.White), // Wide landscape
		},
	}
	vol2.Chapters[chap2ID] = chap2

	// Add volumes to manga
	manga.Volumes[vol1ID] = vol1
	manga.Volumes[vol2ID] = vol2

	return manga
}

// createTempDir returns (string, error) for compatibility with kepub_opf_test.go
func createTempDir(t *testing.T, prefix string) (string, error) {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir, nil
}

// cleanupTempDir returns error for compatibility
func cleanupTempDir(t *testing.T, dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to cleanup temp dir: %v", err)
	}
	return nil
}

// updateOPFMetadata updates the OPF metadata with the provided values
func updateOPFMetadata(opfPath string, metadata map[string]string) error {
	content, err := os.ReadFile(opfPath)
	if err != nil {
		return fmt.Errorf("failed to read OPF file: %w", err)
	}
	doc, err := parseOPF(content)
	if err != nil {
		return err
	}
	metadataNode := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "metadata"
	})
	if metadataNode == nil {
		return fmt.Errorf("metadata section not found in OPF")
	}
	// Insert/overwrite provided metadata keys
	for key, value := range metadata {
		found := false
		for c := metadataNode.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == key {
				if c.FirstChild != nil {
					c.FirstChild.Data = value
				} else {
					textNode := &html.Node{Type: html.TextNode, Data: value}
					c.AppendChild(textNode)
				}
				found = true
				break
			}
		}
		if !found {
			newNode := &html.Node{Type: html.ElementNode, Data: key}
			textNode := &html.Node{Type: html.TextNode, Data: value}
			newNode.AppendChild(textNode)
			metadataNode.AppendChild(newNode)
		}
	}
	// Ensure all <meta property=...> tags have a content attribute (migrate text node to content attr)
	for c := metadataNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "meta" {
			var hasProperty, hasContent bool
			var contentIdx int = -1
			for i, attr := range c.Attr {
				if attr.Key == "property" {
					hasProperty = true
				}
				if attr.Key == "content" {
					hasContent = true
					contentIdx = i
				}
			}
			if hasProperty && c.FirstChild != nil && strings.TrimSpace(c.FirstChild.Data) != "" {
				trimmed := strings.TrimSpace(c.FirstChild.Data)
				if hasContent {
					c.Attr[contentIdx].Val = trimmed
				} else {
					c.Attr = append(c.Attr, html.Attribute{Key: "content", Val: trimmed})
				}
				c.RemoveChild(c.FirstChild)
			}
		}
	}
	// Ensure required Kobo/rendition meta tags are present
	requiredMeta := []struct{ property, content string }{
		{"kobo:content-type", "comic"},
		{"kobo:epub-version", "3.0"},
		{"rendition:layout", "pre-paginated"},
		{"rendition:orientation", "portrait"},
		{"rendition:spread", "none"},
		{"rendition:flow", "paginated"},
	}
	existing := map[string]bool{}
	for c := metadataNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "meta" {
			var prop string
			for _, attr := range c.Attr {
				if attr.Key == "property" {
					prop = attr.Val
				}
			}
			if prop != "" {
				existing[prop] = true
			}
		}
	}
	for _, m := range requiredMeta {
		if !existing[m.property] {
			metaNode := &html.Node{Type: html.ElementNode, Data: "meta"}
			metaNode.Attr = append(metaNode.Attr, html.Attribute{Key: "property", Val: m.property})
			metaNode.Attr = append(metaNode.Attr, html.Attribute{Key: "content", Val: m.content})
			metadataNode.AppendChild(metaNode)
		}
	}
	// Ensure dcterms:modified meta is present (with content attribute)
	hasDctermsModified := false
	for c := metadataNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "meta" {
			var prop, content string
			for _, attr := range c.Attr {
				if attr.Key == "property" {
					prop = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}
			if prop == "dcterms:modified" && content != "" {
				hasDctermsModified = true
				break
			}
		}
	}
	if !hasDctermsModified {
		metaNode := &html.Node{Type: html.ElementNode, Data: "meta"}
		metaNode.Attr = append(metaNode.Attr, html.Attribute{Key: "property", Val: "dcterms:modified"})
		metaNode.Attr = append(metaNode.Attr, html.Attribute{Key: "content", Val: "2022-01-01T12:00:00Z"})
		metadataNode.AppendChild(metaNode)
	}
	// Set page-progression-direction="rtl" on <spine> if not present
	spineNode := findNode(doc, "spine")
	if spineNode != nil {
		hasDir := false
		for _, attr := range spineNode.Attr {
			if attr.Key == "page-progression-direction" {
				hasDir = true
				break
			}
		}
		if !hasDir {
			spineNode.Attr = append(spineNode.Attr, html.Attribute{Key: "page-progression-direction", Val: "rtl"})
		}
	}
	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		return fmt.Errorf("failed to render OPF: %w", err)
	}
	err = os.WriteFile(opfPath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write OPF file: %w", err)
	}
	return nil
}

// parseOPF accepts []byte for compatibility
func parseOPF(content []byte) (*html.Node, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse OPF: %w", err)
	}
	return doc, nil
}

// findNode can accept a string or a matcher func
func findNode(node *html.Node, match interface{}) *html.Node {
	switch m := match.(type) {
	case string:
		return findNode(node, func(n *html.Node) bool { return n.Data == m })
	case func(*html.Node) bool:
		if m(node) {
			return node
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if result := findNode(c, m); result != nil {
				return result
			}
		}
		return nil
	default:
		return nil
	}
}

// addMetaElement adds a <meta> element to the <metadata> section of the OPF document
// Accepts a node, tag name, and attributes as a map[string]string
func addMetaElement(parent *html.Node, tag string, attrs map[string]string) {
	meta := &html.Node{
		Type: html.ElementNode,
		Data: tag,
	}
	for k, v := range attrs {
		meta.Attr = append(meta.Attr, html.Attribute{Key: k, Val: v})
	}
	parent.AppendChild(meta)
}

// saveOPF writes the OPF document to the given file path
func saveOPF(doc *html.Node, path string) error {
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// setAttribute sets or updates an attribute on an HTML node
func setAttribute(node *html.Node, key, value string) {
	for i, attr := range node.Attr {
		if attr.Key == key {
			node.Attr[i].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}
