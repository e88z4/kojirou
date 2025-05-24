// Package kepubconv provides KEPUB conversion logic without import cycles.
package kepubconv

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bmaupin/go-epub"
	"golang.org/x/net/html"
)

// KEPUBExtension is the standard extension for Kobo KEPUB files
const KEPUBExtension = ".kepub.epub"

// ConvertToKEPUB transforms a standard EPUB object into a Kobo-compatible KEPUB.
func ConvertToKEPUB(epubBook *epub.Epub) ([]byte, error) {
	// Input validation
	if epubBook == nil {
		return nil, errors.New("nil EPUB object provided")
	}

	// Create a temporary directory for processing
	tempDir, err := ioutil.TempDir("", "kepub-conversion")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Step 1: Write the EPUB to a temporary file
	epubPath := filepath.Join(tempDir, "original.epub")
	err = epubBook.Write(epubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to write EPUB to temp file: %w", err)
	}

	// Step 2: Extract EPUB contents to a directory
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create extraction directory: %w", err)
	}

	if err := extractEPUB(epubPath, extractDir); err != nil {
		return nil, fmt.Errorf("failed to extract EPUB: %w", err)
	}

	// Step 3: Process EPUB contents for Kobo compatibility
	if err := processEPUBForKobo(extractDir); err != nil {
		return nil, fmt.Errorf("failed to process EPUB for Kobo: %w", err)
	}

	// Step 3b: Apply manga-specific enhancements
	// TODO: Implement ProcessMangaForKEPUB function
	/*
		if err := ProcessMangaForKEPUB(extractDir); err != nil {
			return nil, fmt.Errorf("failed to apply manga enhancements: %w", err)
		}
	*/

	// Step 4: Repackage as KEPUB
	kepubPath := filepath.Join(tempDir, "converted.kepub.epub")
	if err := packageKEPUB(extractDir, kepubPath); err != nil {
		return nil, fmt.Errorf("failed to package KEPUB: %w", err)
	}

	// Step 5: Read the final KEPUB data
	kepubData, err := ioutil.ReadFile(kepubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read KEPUB data: %w", err)
	}

	return kepubData, nil
}

// extractEPUB extracts the contents of an EPUB file to a specified directory.
func extractEPUB(epubPath, extractDir string) error {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer r.Close()

	for _, file := range r.File {
		fPath := filepath.Join(extractDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(fPath, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
			return fmt.Errorf("failed to create file directory: %w", err)
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to open file for writing: %w", err)
		}
		defer outFile.Close()

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}
		defer rc.Close()

		if _, err := io.Copy(outFile, rc); err != nil {
			return fmt.Errorf("failed to copy file contents: %w", err)
		}
	}

	return nil
}

// processEPUBForKobo modifies the contents of an extracted EPUB directory for Kobo compatibility.
func processEPUBForKobo(extractDir string) error {
	// Example: Add Kobo-specific attributes to HTML files
	htmlFiles, err := filepath.Glob(filepath.Join(extractDir, "*.html"))
	if err != nil {
		return fmt.Errorf("failed to find HTML files: %w", err)
	}

	for _, htmlFile := range htmlFiles {
		data, err := ioutil.ReadFile(htmlFile)
		if err != nil {
			return fmt.Errorf("failed to read HTML file: %w", err)
		}

		modifiedData := addKoboAttributes(data)

		if err := ioutil.WriteFile(htmlFile, modifiedData, 0644); err != nil {
			return fmt.Errorf("failed to write modified HTML file: %w", err)
		}
	}

	return nil
}

// addKoboAttributes adds Kobo-specific attributes to HTML content.
func addKoboAttributes(data []byte) []byte {
	// Example: Add Kobo spans around paragraphs
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return data // Return original data if parsing fails
	}

	var modifyNode func(*html.Node)
	modifyNode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			span := &html.Node{
				Type: html.ElementNode,
				Data: "span",
				Attr: []html.Attribute{
					{Key: "class", Val: "koboSpan"},
				},
			}
			span.AppendChild(n.FirstChild)
			n.AppendChild(span)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			modifyNode(c)
		}
	}

	modifyNode(doc)

	var buf bytes.Buffer
	html.Render(&buf, doc)
	return buf.Bytes()
}

// packageKEPUB repackages the contents of a directory into a KEPUB file.
func packageKEPUB(extractDir, kepubPath string) error {
	outFile, err := os.Create(kepubPath)
	if err != nil {
		return fmt.Errorf("failed to create KEPUB file: %w", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(extractDir, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, file)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to package KEPUB: %w", err)
	}

	return nil
}
