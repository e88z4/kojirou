// Package output provides format output types and interfaces for ebook formats
package output

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/leotaku/kojirou/cmd/formats/kepubconv"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/mobi"
)

// FormatOutput represents the output of a format generator
type FormatOutput interface {
	// Extension returns the file extension for this format (without dot)
	Extension() string
	// GetBytes returns the bytes of the generated ebook
	GetBytes() ([]byte, error)
}

// MobiOutput wraps a mobi.Book to implement FormatOutput
type MobiOutput struct {
	*mobi.Book
}

func NewMobiOutput(book *mobi.Book) MobiOutput {
	return MobiOutput{Book: book}
}

func (m MobiOutput) Extension() string {
	return "azw3"
}

func (m MobiOutput) GetBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.Realize().Write(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetCoverImage returns the cover image if one exists
func (m MobiOutput) GetCoverImage() image.Image {
	return m.CoverImage
}

// GetThumbFilename returns the thumbnail filename for Kindle devices
func (m MobiOutput) GetThumbFilename() string {
	return m.Book.GetThumbFilename()
}

// EpubWriter exposes Write methods for epub file
type EpubWriter interface {
	Write(io.Writer) error
}

// EpubOutput wraps an epub.Epub to implement FormatOutput
type EpubOutput struct {
	*epub.Epub
}

func NewEpubOutput(epub *epub.Epub) EpubOutput {
	return EpubOutput{Epub: epub}
}

func (e EpubOutput) Extension() string {
	return "epub"
}

func (e EpubOutput) GetBytes() ([]byte, error) {
	tempFile, err := os.CreateTemp("", "epub-*.epub")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write to temp file since go-epub requires a filename
	if err := e.Write(tempFile.Name()); err != nil {
		return nil, fmt.Errorf("write epub: %w", err)
	}

	// Read back the file
	return os.ReadFile(tempFile.Name())
}

// KepubOutput wraps an epub.Epub to implement FormatOutput
type KepubOutput struct {
	*epub.Epub
}

func NewKepubOutput(epub *epub.Epub) KepubOutput {
	return KepubOutput{Epub: epub}
}

func (k KepubOutput) Extension() string {
	return "kepub.epub"
}

func (k KepubOutput) GetBytes() ([]byte, error) {
	return kepubconv.ConvertToKEPUB(k.Epub, "", 0)
}
