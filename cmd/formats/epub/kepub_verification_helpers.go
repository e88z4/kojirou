// Package epub provides EPUB and KEPUB conversion functionality
// This file contains test utilities for KEPUB verification
package epub

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"io"
	"os"
	"strings"
	"testing"

	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/net/html"
)

// verifyKoboSpans checks if the KEPUB HTML has Kobo spans
func verifyKoboSpans(t *testing.T, data []byte) {
	// Create a temp file with the KEPUB data
	tempFile, err := os.CreateTemp("", "kepub-*.epub")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(data); err != nil {
		t.Fatalf("Failed to write KEPUB data to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Open the EPUB file
	r, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open KEPUB: %v", err)
	}
	defer r.Close()

	// Look for HTML files and check for Kobo spans
	foundSpans := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			rc, err := f.Open()
			if err != nil {
				t.Errorf("Failed to open file %s: %v", f.Name, err)
				continue
			}

			content, err := io.ReadAll(rc)
			errClose := rc.Close()
			if err != nil {
				t.Errorf("Failed to read file %s: %v", f.Name, err)
				continue
			}
			if errClose != nil {
				t.Errorf("Failed to close file %s: %v", f.Name, errClose)
			}

			// Check for koboSpan class
			if bytes.Contains(content, []byte("koboSpan")) {
				foundSpans = true
				break
			}
		}
	}

	if !foundSpans {
		t.Error("No Kobo spans found in KEPUB")
	}
}

// verifyKoboFixedLayout checks if the KEPUB has fixed layout metadata
func verifyKoboFixedLayout(t *testing.T, data []byte) {
	// Create a reader for the zip data
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to open KEPUB as zip: %v", err)
	}

	// Look for OPF file
	foundFixedLayout := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Errorf("Failed to open file %s: %v", f.Name, err)
				continue
			}

			content, err := io.ReadAll(rc)
			errClose := rc.Close()
			if err != nil {
				t.Errorf("Failed to read file %s: %v", f.Name, err)
				continue
			}
			if errClose != nil {
				t.Errorf("Failed to close file %s: %v", f.Name, errClose)
			}

			// Check for fixed-layout property
			if bytes.Contains(content, []byte("rendition:layout")) &&
				bytes.Contains(content, []byte("pre-paginated")) {
				foundFixedLayout = true
				break
			}
		}
	}

	if !foundFixedLayout {
		t.Error("No fixed layout metadata found in KEPUB")
	}
}

// verifyKoboNamespaces checks if the KEPUB has Kobo namespaces
func verifyKoboNamespaces(t *testing.T, data []byte) {
	// Create a reader for the zip data
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to open KEPUB as zip: %v", err)
	}

	// Check for Kobo namespace in HTML files
	foundNamespace := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			rc, err := f.Open()
			if err != nil {
				t.Errorf("Failed to open file %s: %v", f.Name, err)
				continue
			}

			content, err := io.ReadAll(rc)
			errClose := rc.Close()
			if err != nil {
				t.Errorf("Failed to read file %s: %v", f.Name, err)
				continue
			}
			if errClose != nil {
				t.Errorf("Failed to close file %s: %v", f.Name, errClose)
			}

			// Parse the HTML
			doc, err := html.Parse(bytes.NewReader(content))
			if err != nil {
				t.Errorf("Failed to parse HTML %s: %v", f.Name, err)
				continue
			}

			// Check for Kobo namespace in the HTML element
			var checkNamespace func(*html.Node) bool
			checkNamespace = func(n *html.Node) bool {
				if n.Type == html.ElementNode && n.Data == "html" {
					for _, attr := range n.Attr {
						if attr.Key == "xmlns:epub" && attr.Val == "http://www.kobo.com/ns/1.0" {
							return true
						}
					}
				}

				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if checkNamespace(c) {
						return true
					}
				}

				return false
			}

			if checkNamespace(doc) {
				foundNamespace = true
				break
			}
		}
	}

	if !foundNamespace {
		t.Error("No Kobo namespace found in KEPUB")
	}
}

// Add a new verification function for content type
// verifyKoboContentType checks for Kobo-specific content types in the KEPUB
func verifyKoboContentType(t *testing.T, data []byte) {
	// Create a reader for the zip data
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to open KEPUB as zip: %v", err)
	}

	// Look for mimetype file or META-INF/container.xml or package.opf
	found := false
	for _, f := range r.File {
		if f.Name == "mimetype" {
			rc, err := f.Open()
			if err != nil {
				t.Errorf("Failed to open mimetype file: %v", err)
				continue
			}

			content, err := io.ReadAll(rc)
			errClose := rc.Close()
			if err != nil {
				t.Errorf("Failed to read mimetype content: %v", err)
				continue
			}
			if errClose != nil {
				t.Errorf("Failed to close mimetype file: %v", errClose)
			}

			// Check the mimetype (should be standard epub mimetype)
			if string(bytes.TrimSpace(content)) == "application/epub+zip" {
				found = true
			} else {
				t.Errorf("Unexpected mimetype: %s", string(content))
			}
		}
		if strings.HasSuffix(f.Name, "package.opf") {
			found = true
		}
	}

	if !found {
		t.Error("No valid mimetype file or package.opf found in KEPUB")
	}
}
func createMultiVolumeTestManga() md.Manga {
	manga := createTestManga()

	// Add a second volume
	vol2ID := md.NewIdentifier("2")
	vol2 := md.Volume{
		Info: md.VolumeInfo{
			Identifier: vol2ID,
		},
		Chapters: map[md.Identifier]md.Chapter{},
	}

	// Add a chapter to the second volume
	chap2ID := md.NewIdentifier("2-1")
	chap2 := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chap2ID,
			Title:            "Chapter 2",
			VolumeIdentifier: vol2ID,
		},
		Pages: map[int]image.Image{
			0: createTestImage(1000, 1400, color.White),
			1: createTestImage(1000, 1400, color.White),
		},
	}

	vol2.Chapters[chap2ID] = chap2
	manga.Volumes[vol2ID] = vol2

	return manga
}
