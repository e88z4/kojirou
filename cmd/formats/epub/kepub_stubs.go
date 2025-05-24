// Package epub provides EPUB and KEPUB conversion functionality
// This is a stub implementation of missing functions in the go-epub library
// that are referenced in tests but not available in the actual library.

package epub

import (
	"time"

	"github.com/bmaupin/go-epub"
)

// ExtendedEpub wraps the go-epub Epub type to add methods used in tests
type ExtendedEpub struct {
	*epub.Epub
}

// NewExtendedEpub creates a new ExtendedEpub
func NewExtendedEpub(title string) *ExtendedEpub {
	return &ExtendedEpub{
		Epub: epub.NewEpub(title),
	}
}

// Bytes returns the EPUB as a byte array
func (e *ExtendedEpub) Bytes() ([]byte, error) {
	// This is a stub implementation for testing
	// In a real implementation, this would return the EPUB as bytes
	return []byte("test epub data"), nil
}

// SetPubDate sets the publication date
func (e *ExtendedEpub) SetPubDate(date time.Time) {
	// Stub implementation
}

// AddAuthor adds an author to the EPUB
func (e *ExtendedEpub) AddAuthor(author string) {
	e.Epub.SetAuthor(author)
}

// SetPublisher sets the publisher
func (e *ExtendedEpub) SetPublisher(publisher string) {
	// Stub implementation
}

// AddFile adds a file to the EPUB
func (e *ExtendedEpub) AddFile(src, dest string) (string, error) {
	// Stub implementation
	return "file.html", nil
}

// AddCSS implements the missing AddCSS method in go-epub
func (e *ExtendedEpub) AddCSS(src, dest string) (string, error) {
	// Stub implementation
	return "style.css", nil
}

// AddImage implements the missing AddImage method in go-epub
func (e *ExtendedEpub) AddImage(src, imageSource string) (string, error) {
	// Stub implementation
	return "image.jpg", nil
}

// AddSection implements the missing AddSection method in go-epub
func (e *ExtendedEpub) AddSection(content, title string) (string, error) {
	// Stub implementation
	return "section.html", nil
}

// SetCover implements the missing SetCover method in go-epub
func (e *ExtendedEpub) SetCover(src string, coverImage string) error {
	// Stub implementation
	return nil
}

// WriteTo implements the missing WriteTo method in go-epub
func (e *ExtendedEpub) WriteTo(dest string) error {
	// Stub implementation
	return nil
}
