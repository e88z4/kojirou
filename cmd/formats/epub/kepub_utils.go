package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Manga-specific processing for KEPUB

// ProcessMangaForKEPUB applies manga-specific enhancements to KEPUB files
func ProcessMangaForKEPUB(extractDir string) error {
	// Find and process all content files
	htmlFiles, err := findContentFiles(extractDir)
	if err != nil {
		return fmt.Errorf("failed to find content files: %w", err)
	}

	for _, htmlFile := range htmlFiles {
		if err := processMangaHTML(htmlFile); err != nil {
			return fmt.Errorf("failed to process manga HTML file %s: %w", htmlFile, err)
		}
	}

	// Find and process OPF files to add manga-specific metadata
	opfFiles, err := findOPFFiles(extractDir)
	if err != nil {
		return fmt.Errorf("failed to find OPF files: %w", err)
	}

	for _, opfFile := range opfFiles {
		if err := addMangaMetadata(opfFile); err != nil {
			return fmt.Errorf("failed to add manga metadata to %s: %w", opfFile, err)
		}
	}

	return nil
}

// processMangaHTML processes HTML content specifically for manga
func processMangaHTML(htmlFile string) error {
	// Read the HTML content
	content, err := ioutil.ReadFile(htmlFile)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}

	// Parse the HTML
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Apply manga-specific enhancements
	optimizeMangaImages(doc)
	addMangaFixedLayoutAttributes(doc)

	// Write the modified HTML back to the file
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return fmt.Errorf("failed to render modified HTML: %w", err)
	}

	if err := ioutil.WriteFile(htmlFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write modified HTML: %w", err)
	}

	return nil
}

// optimizeMangaImages optimizes image elements for manga viewing
func optimizeMangaImages(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "img" {
		// Add Kobo-specific attributes for better image rendering
		hasClass := false
		for i, attr := range n.Attr {
			if attr.Key == "class" {
				n.Attr[i].Val = attr.Val + " kobo-manga-image"
				hasClass = true
				break
			}
		}

		if !hasClass {
			n.Attr = append(n.Attr, html.Attribute{
				Key: "class",
				Val: "kobo-manga-image",
			})
		}

		// Ensure image has proper width/height (unless already specified)
		hasWidth := false
		hasHeight := false
		for _, attr := range n.Attr {
			if attr.Key == "width" {
				hasWidth = true
			}
			if attr.Key == "height" {
				hasHeight = true
			}
		}

		if !hasWidth {
			n.Attr = append(n.Attr, html.Attribute{
				Key: "width",
				Val: "100%",
			})
		}

		if !hasHeight {
			n.Attr = append(n.Attr, html.Attribute{
				Key: "height",
				Val: "auto",
			})
		}
	}

	// Process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		optimizeMangaImages(c)
	}
}

// addMangaFixedLayoutAttributes adds fixed layout attributes for manga pages
func addMangaFixedLayoutAttributes(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "body" {
		// Add fixed layout attributes
		hasClass := false
		for i, attr := range n.Attr {
			if attr.Key == "class" {
				n.Attr[i].Val = attr.Val + " kobo-fixed-layout"
				hasClass = true
				break
			}
		}

		if !hasClass {
			n.Attr = append(n.Attr, html.Attribute{
				Key: "class",
				Val: "kobo-fixed-layout",
			})
		}

		// Add manga orientation
		n.Attr = append(n.Attr, html.Attribute{
			Key: "epub:type",
			Val: "kobo:manga",
		})
	}

	// Process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		addMangaFixedLayoutAttributes(c)
	}
}

// addMangaMetadata adds manga-specific metadata to OPF file
func addMangaMetadata(opfFile string) error {
	// Read the OPF content
	content, err := ioutil.ReadFile(opfFile)
	if err != nil {
		return fmt.Errorf("failed to read OPF file: %w", err)
	}

	opfContent := string(content)

	// Find the metadata section
	metadataRegex := regexp.MustCompile(`<metadata[^>]*>.*?</metadata>`)
	metadataMatch := metadataRegex.FindString(opfContent)

	if metadataMatch == "" {
		// If no metadata section found, don't modify the content
		return nil
	}

	// Add manga-specific metadata
	mangaMetadata := `
    <meta property="rendition:layout">pre-paginated</meta>
    <meta property="rendition:orientation">portrait</meta>
    <meta property="rendition:spread">none</meta>
    <meta property="kobo:manga">true</meta>
  `

	// Insert before the closing metadata tag
	modifiedMetadata := strings.Replace(
		metadataMatch,
		"</metadata>",
		mangaMetadata+"</metadata>",
		1,
	)

	// Replace the original metadata section with the modified one
	modifiedContent := metadataRegex.ReplaceAllString(opfContent, modifiedMetadata)

	// Write the modified OPF back to the file
	if err := ioutil.WriteFile(opfFile, []byte(modifiedContent), 0644); err != nil {
		return fmt.Errorf("failed to write modified OPF: %w", err)
	}

	return nil
}

// CheckForKoboSpanID checks if a span has a valid Kobo ID
func CheckForKoboSpanID(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "span" {
		for _, attr := range n.Attr {
			if attr.Key == "id" && strings.HasPrefix(attr.Val, "kobo") {
				return true
			}
		}
	}

	// Check children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if CheckForKoboSpanID(c) {
			return true
		}
	}

	return false
}

// IsKEPUB checks if an EPUB file has been converted to KEPUB format
func IsKEPUB(filePath string) (bool, error) {
	// Open the EPUB file
	reader, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()

	// Get file info to determine size
	info, err := reader.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to get file info: %w", err)
	}

	// Open as ZIP
	r, err := zip.NewReader(reader, info.Size())
	if err != nil {
		return false, fmt.Errorf("failed to open as ZIP: %w", err)
	}

	// Check for Kobo-specific features
	for _, f := range r.File {
		// Check OPF files for Kobo metadata
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				continue
			}

			content, err := ioutil.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			if strings.Contains(string(content), "kobo") {
				return true, nil
			}
		}

		// Check HTML files for Kobo spans
		if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}

			content, err := ioutil.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			if strings.Contains(string(content), "kobo") {
				return true, nil
			}
		}
	}

	return false, nil
}

// findContentFiles is a test-local copy for test helpers
func findContentFiles(extractDir string) ([]string, error) {
	var contentFiles []string

	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".xhtml")) {
			contentFiles = append(contentFiles, path)
		}

		return nil
	})

	return contentFiles, err
}

// findOPFFiles is a test-local copy for test helpers
func findOPFFiles(extractDir string) ([]string, error) {
	var opfFiles []string

	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".opf") {
			opfFiles = append(opfFiles, path)
		}

		return nil
	})

	return opfFiles, err
}
