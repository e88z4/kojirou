// Package epub provides OPF file processing functions for KEPUB conversion
package epub

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// OPFProcessor handles modification of OPF files for Kobo compatibility
type OPFProcessor struct {
	Package *OPFPackage
}

// OPFPackage represents the root element of an OPF file
type OPFPackage struct {
	XMLName         xml.Name        `xml:"package"`
	Version         string          `xml:"version,attr"`
	UniqueID        string          `xml:"unique-identifier,attr"`
	Metadata        OPFMetadata     `xml:"metadata"`
	Manifest        OPFManifest     `xml:"manifest"`
	Spine           OPFSpine        `xml:"spine"`
	NamespacePrefix string          `xml:",attr"`
	NamespaceMap    map[string]bool // Used to track namespaces
}

// OPFMetadata represents the metadata section of an OPF file
type OPFMetadata struct {
	XMLName    xml.Name        `xml:"metadata"`
	Identifier []OPFIdentifier `xml:"identifier"`
	Title      []string        `xml:"title"`
	Language   []string        `xml:"language"`
	Meta       []OPFMeta       `xml:"meta"`
	Creator    []OPFCreator    `xml:"creator"`
}

// OPFIdentifier represents an identifier element
type OPFIdentifier struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

// OPFMeta represents a meta element
type OPFMeta struct {
	Name     string `xml:"name,attr,omitempty"`
	Content  string `xml:"content,attr"`
	Property string `xml:"property,attr,omitempty"`
}

// OPFCreator represents a creator element
type OPFCreator struct {
	Role  string `xml:"role,attr,omitempty"`
	Value string `xml:",chardata"`
}

// OPFManifest represents the manifest section
type OPFManifest struct {
	XMLName xml.Name  `xml:"manifest"`
	Items   []OPFItem `xml:"item"`
}

// OPFItem represents an item in the manifest
type OPFItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

// OPFSpine represents the spine section
type OPFSpine struct {
	XMLName  xml.Name     `xml:"spine"`
	Toc      string       `xml:"toc,attr"`
	PageDir  string       `xml:"page-progression-direction,attr,omitempty"`
	ItemRefs []OPFItemRef `xml:"itemref"`
}

// OPFItemRef represents an itemref in the spine
type OPFItemRef struct {
	IDRef      string `xml:"idref,attr"`
	Linear     string `xml:"linear,attr,omitempty"`
	Properties string `xml:"properties,attr,omitempty"`
}

// NewOPFProcessor creates a new OPF processor from file content
func NewOPFProcessor(content []byte) (*OPFProcessor, error) {
	var pkg OPFPackage
	err := xml.Unmarshal(content, &pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OPF: %w", err)
	}

	// Initialize the namespace map
	pkg.NamespaceMap = make(map[string]bool)

	return &OPFProcessor{
		Package: &pkg,
	}, nil
}

// AddKoboMetadata adds Kobo-specific metadata to the OPF file
func (p *OPFProcessor) AddKoboMetadata() {
	// Add Kobo metadata items
	p.Package.Metadata.Meta = append(p.Package.Metadata.Meta, OPFMeta{
		Name:    "kobo.displayed.title",
		Content: strings.Join(p.Package.Metadata.Title, " - "),
	})

	// Check if we have a title and add it as Kobo content title
	if len(p.Package.Metadata.Title) > 0 {
		p.Package.Metadata.Meta = append(p.Package.Metadata.Meta, OPFMeta{
			Property: "title",
			Content:  p.Package.Metadata.Title[0],
		})
	}

	// Add other Kobo-specific metadata
	p.Package.Metadata.Meta = append(p.Package.Metadata.Meta,
		OPFMeta{Name: "fixed-layout", Content: "true"},
		OPFMeta{Name: "orientation-lock", Content: "portrait"},
		OPFMeta{Name: "book-type", Content: "comic"},
		OPFMeta{Name: "zero-margin", Content: "true"},
	)
}

// Serialize converts the OPF object back to XML
func (p *OPFProcessor) Serialize() ([]byte, error) {
	output, err := xml.MarshalIndent(p.Package, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize OPF: %w", err)
	}

	// Add XML declaration
	output = append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"), output...)
	return output, nil
}
