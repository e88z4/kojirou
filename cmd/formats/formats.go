// Package formats provides format-specific functionality for different ebook formats
package formats

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/mobi"
)

// FormatType represents the type of ebook format to generate
type FormatType string

const (
	// FormatMobi represents the Kindle MOBI format
	FormatMobi FormatType = "mobi"
	// FormatEpub represents the standard EPUB format
	FormatEpub FormatType = "epub"
	// FormatKepub represents the Kobo-specific EPUB format
	FormatKepub FormatType = "kepub"
)

// String returns the string representation of the format type
func (f FormatType) String() string {
	return string(f)
}

// FormatOutput represents the output of a format generator
type FormatOutput interface {
	// Extension returns the file extension for this format (without dot)
	Extension() string
	// Bytes returns the bytes of the generated ebook
	Bytes() ([]byte, error)
}

// MobiOutput wraps a mobi.Book to implement FormatOutput
type MobiOutput struct {
	*mobi.Book
}

func (m MobiOutput) Extension() string {
	return "azw3"
}

func (m MobiOutput) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := m.Realize().Write(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EpubOutput wraps an epub.Epub to implement FormatOutput
type EpubOutput struct {
	*epub.Epub
}

func (e EpubOutput) Extension() string {
	return "epub"
}

func (e EpubOutput) Bytes() ([]byte, error) {
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

func (k KepubOutput) Extension() string {
	return "kepub.epub"
}

func (k KepubOutput) Bytes() ([]byte, error) {
	// TODO: Convert EPUB to Kobo EPUB format
	tempFile, err := os.CreateTemp("", "kepub-*.epub")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write to temp file since go-epub requires a filename
	if err := k.Write(tempFile.Name()); err != nil {
		return nil, fmt.Errorf("write kepub: %w", err)
	}

	// Read back the file
	return os.ReadFile(tempFile.Name())
}

// ParseFormats converts a comma-separated string of format names into a slice of FormatType
func ParseFormats(formatStr string) ([]FormatType, error) {
	parts := strings.Split(formatStr, ",")
	formats := make([]FormatType, 0, len(parts))

	for _, part := range parts {
		format := FormatType(strings.TrimSpace(strings.ToLower(part)))
		switch format {
		case FormatMobi, FormatEpub, FormatKepub:
			formats = append(formats, format)
		default:
			return nil, fmt.Errorf("unsupported format: %s", part)
		}
	}

	if len(formats) == 0 {
		return nil, fmt.Errorf("no valid formats specified")
	}

	return formats, nil
}
