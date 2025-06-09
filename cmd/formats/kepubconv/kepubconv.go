package kepubconv

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/kojirou/cmd/formats/util"
	"golang.org/x/net/html"
)

// KEPUBExtension is the standard extension for Kobo KEPUB files
const KEPUBExtension = ".kepub.epub"

// ConvertToKEPUB transforms a standard EPUB object into a Kobo-compatible KEPUB.
func ConvertToKEPUB(epubBook *epub.Epub, seriesTitle string, seriesIndex float64) ([]byte, error) {
	var retErr error
	// Input validation
	if epubBook == nil {
		return nil, errors.New("nil EPUB object provided")
	}
	if !hasSections(epubBook) {
		return nil, errors.New("empty EPUB: no content sections found")
	}

	// Create a temporary directory for processing
	tempDir, err := os.MkdirTemp("", "kepub-conversion")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := util.ForceRemoveAll(tempDir); err != nil && retErr == nil {
			retErr = err
		}
	}()

	// Create necessary CSS files for EPUB write operation
	// The go-epub library may look in several places for CSS files
	cssContent := "body { margin: 0; padding: 0; } img { display: block; max-width: 100%; height: auto; }"

	// Create a style.css file in multiple possible locations
	for _, dir := range []string{"css", "001", ""} {
		styleDir := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(styleDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create style directory %s: %w", styleDir, err)
		}
		cssPath := filepath.Join(styleDir, "style.css")
		if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write CSS file %s: %w", cssPath, err)
		}
	}

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
	if err := processEPUBForKobo(extractDir, seriesTitle, seriesIndex); err != nil {
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
	kepubData, err := os.ReadFile(kepubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read KEPUB data: %w", err)
	}

	// Clean up: Remove debug output directory if it exists
	debugOutdir := "/home/felix/src/kojirou/kepub_debug_tmp"
	_ = os.RemoveAll(debugOutdir)

	return kepubData, retErr
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
func processEPUBForKobo(extractDir string, seriesTitle string, seriesIndex float64) error {
	// 1. Inject Kobo-specific metadata into OPF files (recursive)
	opfFiles := []string{}
	if err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".opf") {
			opfFiles = append(opfFiles, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to walk for OPF files: %w", err)
	}
	for _, opfFile := range opfFiles {
		data, err := os.ReadFile(opfFile)
		if err != nil {
			return fmt.Errorf("failed to read OPF file: %w", err)
		}
		output := injectKoboMetadata(data, seriesTitle, seriesIndex)
		// --- Ensure cover image is first in manifest and referenced in metadata ---
		output, err = ensureKoboCoverInOPF(output)
		if err != nil {
			return fmt.Errorf("failed to ensure Kobo cover in OPF: %w", err)
		}
		if err := os.WriteFile(opfFile, output, 0644); err != nil {
			return fmt.Errorf("failed to write modified OPF file: %w", err)
		}
	}

	// 2. Add Kobo-specific attributes to HTML/XHTML files (recursive)
	htmlFiles := []string{}
	if err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".html") || strings.HasSuffix(strings.ToLower(path), ".xhtml")) {
			htmlFiles = append(htmlFiles, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to walk for HTML/XHTML files: %w", err)
	}
	for _, htmlFile := range htmlFiles {
		data, err := os.ReadFile(htmlFile)
		if err != nil {
			return fmt.Errorf("failed to read HTML/XHTML file: %w", err)
		}
		modifiedData := addKoboAttributes(data)
		if err := os.WriteFile(htmlFile, modifiedData, 0644); err != nil {
			return fmt.Errorf("failed to write modified HTML/XHTML file: %w", err)
		}
	}

	return nil
}

// injectKoboMetadata adds Kobo-specific metadata to the OPF XML content.
func injectKoboMetadata(data []byte, seriesTitle string, seriesIndex float64) []byte {
	opf := string(data)
	// 1. Inject Kobo/rendition namespaces into <package ...>
	packageRe := regexp.MustCompile(`(?s)<package([^>]*)>`)
	opf = packageRe.ReplaceAllStringFunc(opf, func(pkgTag string) string {
		// Always add Kobo/rendition namespaces if not present
		if !strings.Contains(pkgTag, "xmlns:rendition") {
			pkgTag = strings.Replace(pkgTag, ">", ` xmlns:rendition="http://www.idpf.org/vocab/rendition/#">`, 1)
		}
		if !strings.Contains(pkgTag, "xmlns:kobo") {
			pkgTag = strings.Replace(pkgTag, ">", ` xmlns:kobo="http://kobobooks.com/ns/kobo">`, 1)
		}
		return pkgTag
	})

	// 2. Insert required meta tags as direct children of <metadata>, but only if not already present
	requiredMeta := []struct{ keyType, key, content string }{
		{"property", "kobo:content-type", "comic"},
		{"property", "kobo:epub-version", "3.0"},
		{"property", "rendition:layout", "pre-paginated"},
		{"property", "rendition:orientation", "portrait"},
		{"property", "rendition:spread", "none"},
		{"property", "rendition:flow", "paginated"},
		{"property", "dcterms:modified", time.Now().UTC().Format("2006-01-02T15:04:05Z")},
		{"property", "page-progression-direction", "rtl"},
	}

	// Add Calibre series metadata if series title is provided
	if seriesTitle != "" {
		requiredMeta = append(requiredMeta,
			struct{ keyType, key, content string }{
				keyType: "name",
				key:     "calibre:series",
				content: seriesTitle,
			},
			struct{ keyType, key, content string }{
				keyType: "name",
				key:     "calibre:series_index",
				content: fmt.Sprintf("%.1f", seriesIndex),
			},
		)
	}

	// Check which metadata is already present
	present := map[string]bool{}
	metaRe := regexp.MustCompile(`<meta[^>]+(?:property|name)="([^"]+)"[^>]*/?>`)
	for _, m := range metaRe.FindAllStringSubmatch(opf, -1) {
		present[m[1]] = true
	}

	// Build the metadata insert string with proper escaping
	var metaInsert strings.Builder
	for _, m := range requiredMeta {
		if !present[m.key] {
			metaInsert.WriteString(`<meta `)
			metaInsert.WriteString(m.keyType)
			metaInsert.WriteString(`="`)
			metaInsert.WriteString(xmlEscape(m.key))
			metaInsert.WriteString(`" content="`)
			metaInsert.WriteString(xmlEscape(m.content))
			metaInsert.WriteString(`"/>`)
		}
	}

	// Insert the new metadata before closing </metadata> tag
	metadataCloseRe := regexp.MustCompile(`(?s)(</metadata>)`)
	if metaInsert.Len() > 0 {
		opf = metadataCloseRe.ReplaceAllString(opf, metaInsert.String()+"$1")
	}
	return []byte(opf)
}

// xmlEscape escapes special characters for XML attributes and text content
func xmlEscape(s string) string {
	var buf strings.Builder
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// addKoboAttributes adds Kobo-specific attributes to HTML content.
func addKoboAttributes(data []byte) []byte {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return data // Return original data if parsing fails
	}

	// Ensure Kobo and epub namespaces on <html>
	var ensureNamespaces func(*html.Node)
	ensureNamespaces = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "html" {
			// Remove all existing Kobo/epub namespace attributes (by Key or Namespace)
			newAttrs := make([]html.Attribute, 0, len(n.Attr))
			for _, attr := range n.Attr {
				if (attr.Key == "xmlns:kobo" || attr.Key == "xmlns:epub") || (attr.Namespace == "xmlns" && (attr.Key == "kobo" || attr.Key == "epub")) {
					continue
				}
				newAttrs = append(newAttrs, attr)
			}
			// Always add Kobo and epub namespaces as the first attributes, in canonical order
			attrsWithNS := []html.Attribute{
				{Key: "xmlns:kobo", Val: "http://kobobooks.com/ns/kobo"},
				{Key: "xmlns:epub", Val: "http://www.idpf.org/2007/ops"},
			}
			attrsWithNS = append(attrsWithNS, newAttrs...)
			n.Attr = attrsWithNS
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ensureNamespaces(c)
		}
	}

	// Unique span ID counter
	spanIDCounter := 1
	imgIDCounter := 1

	// Helper to wrap direct text node children in Kobo spans
	wrapTextNodes := func(parent *html.Node) {
		var next *html.Node
		for c := parent.FirstChild; c != nil; c = next {
			next = c.NextSibling
			if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
				span := &html.Node{
					Type: html.ElementNode,
					Data: "span",
					Attr: []html.Attribute{
						{Key: "class", Val: "koboSpan"},
						{Key: "id", Val: fmt.Sprintf("kobo-span-%d", spanIDCounter)},
					},
				}
				spanIDCounter++
				textCopy := &html.Node{Type: html.TextNode, Data: c.Data}
				span.AppendChild(textCopy)
				parent.InsertBefore(span, c)
				parent.RemoveChild(c)
			}
		}
	}

	var modifyNode func(*html.Node)
	modifyNode = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "p" || n.Data == "div") {
			wrapTextNodes(n)
		}
		if n.Type == html.ElementNode && n.Data == "img" {
			hasEpubType := false
			hasID := false
			hasClass := false
			for i, attr := range n.Attr {
				if attr.Key == "epub:type" {
					hasEpubType = true
					n.Attr[i].Val = "kobo"
				}
				if attr.Key == "id" {
					hasID = true
				}
				if attr.Key == "class" {
					if !strings.Contains(attr.Val, "kobo-image") {
						n.Attr[i].Val = attr.Val + " kobo-image"
					}
					hasClass = true
				}
			}
			if !hasEpubType {
				n.Attr = append(n.Attr, html.Attribute{Key: "epub:type", Val: "kobo"})
			}
			if !hasID {
				n.Attr = append(n.Attr, html.Attribute{Key: "id", Val: fmt.Sprintf("kobo_img_%d", imgIDCounter)})
				imgIDCounter++
			}
			if !hasClass {
				n.Attr = append(n.Attr, html.Attribute{Key: "class", Val: "kobo-image"})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			modifyNode(c)
		}
	}

	ensureNamespaces(doc)
	modifyNode(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return data // Return original data if rendering fails
	}
	return buf.Bytes()
}

// hasSections checks if the EPUB has any sections using reflection.
func hasSections(epubBook *epub.Epub) bool {
	v := reflect.ValueOf(epubBook).Elem()
	field := v.FieldByName("sections")
	if !field.IsValid() {
		return false
	}
	return field.Len() > 0
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

	// 1. Write mimetype file first, uncompressed
	mimetypePath := filepath.Join(extractDir, "mimetype")
	mimetypeInfo, err := os.Stat(mimetypePath)
	if err != nil {
		return fmt.Errorf("mimetype file missing: %w", err)
	}
	mimetypeHeader, err := zip.FileInfoHeader(mimetypeInfo)
	if err != nil {
		return fmt.Errorf("failed to create mimetype header: %w", err)
	}
	mimetypeHeader.Name = "mimetype"
	mimetypeHeader.Method = zip.Store // No compression

	mimetypeWriter, err := zipWriter.CreateHeader(mimetypeHeader)
	if err != nil {
		return fmt.Errorf("failed to create mimetype entry: %w", err)
	}
	mimetypeFile, err := os.Open(mimetypePath)
	if err != nil {
		return fmt.Errorf("failed to open mimetype: %w", err)
	}
	if _, err := io.Copy(mimetypeWriter, mimetypeFile); err != nil {
		mimetypeFile.Close()
		return fmt.Errorf("failed to write mimetype: %w", err)
	}
	mimetypeFile.Close()

	// 2. Write all other files (skip mimetype)
	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(extractDir, path)
		if err != nil {
			return err
		}
		if info.IsDir() || relPath == "mimetype" {
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

// ensureKoboCoverInOPF ensures the cover image is the first item in the manifest and referenced in <meta name="cover" content="cover"/>.
func ensureKoboCoverInOPF(opfData []byte) ([]byte, error) {
	type item struct {
		ID         string `xml:"id,attr"`
		Href       string `xml:"href,attr"`
		MediaType  string `xml:"media-type,attr"`
		Properties string `xml:"properties,attr,omitempty"`
	}

	type manifest struct {
		XMLName xml.Name `xml:"manifest"`
		Items   []item   `xml:"item"`
	}

	type meta struct {
		Name     string `xml:"name,attr,omitempty"`
		Content  string `xml:"content,attr,omitempty"`
		Property string `xml:"property,attr,omitempty"`
	}

	type metadata struct {
		XMLName xml.Name `xml:"metadata"`
		Metas   []meta   `xml:"meta"`
	}

	type opfPackage struct {
		XMLName  xml.Name `xml:"package"`
		Metadata metadata `xml:"metadata"`
		Manifest manifest `xml:"manifest"`
	}

	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return opfData, err
	}

	// Find cover image using strict priority order:
	coverIdx := -1
	// 1. Properties contains "cover-image"
	for i, it := range pkg.Manifest.Items {
		if strings.Contains(it.Properties, "cover-image") && strings.HasPrefix(it.MediaType, "image/") {
			coverIdx = i
			break
		}
	}
	// 2. id="cover" and image/*
	if coverIdx == -1 {
		for i, it := range pkg.Manifest.Items {
			if it.ID == "cover" && strings.HasPrefix(it.MediaType, "image/") {
				coverIdx = i
				break
			}
		}
	}
	// 3. href contains 'cover' and image/*
	if coverIdx == -1 {
		for i, it := range pkg.Manifest.Items {
			if strings.Contains(strings.ToLower(it.Href), "cover") && strings.HasPrefix(it.MediaType, "image/") {
				coverIdx = i
				break
			}
		}
	}
	// 4. meta[name=cover] content reference
	if coverIdx == -1 {
		var coverId string
		for _, m := range pkg.Metadata.Metas {
			if m.Name == "cover" {
				coverId = m.Content
				break
			}
		}
		if coverId != "" {
			for i, it := range pkg.Manifest.Items {
				if it.ID == coverId && strings.HasPrefix(it.MediaType, "image/") {
					coverIdx = i
					break
				}
			}
		}
	}
	// 5. first image/* item
	if coverIdx == -1 {
		for i, it := range pkg.Manifest.Items {
			if strings.HasPrefix(it.MediaType, "image/") {
				coverIdx = i
				break
			}
		}
	}

	// Always move the cover to the first position if found
	if coverIdx >= 0 && len(pkg.Manifest.Items) > 0 {
		coverItem := pkg.Manifest.Items[coverIdx]
		// Remove the cover item from its current position
		pkg.Manifest.Items = append(pkg.Manifest.Items[:coverIdx], pkg.Manifest.Items[coverIdx+1:]...)
		// Insert at the front
		pkg.Manifest.Items = append([]item{coverItem}, pkg.Manifest.Items...)
	}

	// Ensure cover id is "cover" and has cover-image property
	if len(pkg.Manifest.Items) > 0 {
		pkg.Manifest.Items[0].ID = "cover"
		if pkg.Manifest.Items[0].Properties == "" {
			pkg.Manifest.Items[0].Properties = "cover-image"
		} else if !strings.Contains(pkg.Manifest.Items[0].Properties, "cover-image") {
			pkg.Manifest.Items[0].Properties += " cover-image"
		}
	}

	// Ensure <meta name="cover" content="cover"/> exists
	hasCoverMeta := false
	for _, m := range pkg.Metadata.Metas {
		if m.Name == "cover" && m.Content == "cover" {
			hasCoverMeta = true
			break
		}
	}
	if !hasCoverMeta {
		pkg.Metadata.Metas = append([]meta{{Name: "cover", Content: "cover"}}, pkg.Metadata.Metas...)
	}

	// Marshal back to XML with proper formatting
	out, err := xml.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return opfData, err
	}

	// Add XML declaration and remove unnecessary whitespace
	out = append([]byte(xml.Header), out...)
	out = regexp.MustCompile(`>\s+<`).ReplaceAll(out, []byte(">\n<"))

	// --- Manual metadata serialization to guarantee all <meta> are escaped ---
	metadataStart := []byte("<metadata>")
	metadataEnd := []byte("</metadata>")
	var metaItems []string
	for _, m := range pkg.Metadata.Metas {
		attrs := []string{}
		if m.Name != "" {
			attrs = append(attrs, "name=\""+xmlEscape(m.Name)+"\"")
		}
		if m.Property != "" {
			attrs = append(attrs, "property=\""+xmlEscape(m.Property)+"\"")
		}
		if m.Content != "" {
			attrs = append(attrs, "content=\""+xmlEscape(m.Content)+"\"")
		}
		metaItems = append(metaItems, "  <meta "+strings.Join(attrs, " ")+"/>")
	}
	metadataBlock := string(metadataStart) + "\n" + strings.Join(metaItems, "\n") + "\n" + string(metadataEnd)
	out = regexp.MustCompile(`<metadata[\s\S]*?</metadata>`).ReplaceAll(out, []byte(metadataBlock))

	// --- Manual manifest serialization to guarantee <item id="cover" ...> is first ---
	manifestStart := []byte("<manifest>")
	manifestEnd := []byte("</manifest>")
	var manifestItems []string
	for _, it := range pkg.Manifest.Items {
		attrs := []string{
			"id=\"" + xmlEscape(it.ID) + "\"",
			"href=\"" + xmlEscape(it.Href) + "\"",
			"media-type=\"" + xmlEscape(it.MediaType) + "\"",
		}
		if it.Properties != "" {
			attrs = append(attrs, "properties=\""+xmlEscape(it.Properties)+"\"")
		}
		manifestItems = append(manifestItems, "  <item "+strings.Join(attrs, " ")+"/>")
	}
	// Reorder so <item id="cover" ...> is first
	coverIdx = -1
	for i, line := range manifestItems {
		if strings.Contains(line, "id=\"cover\"") {
			coverIdx = i
			break
		}
	}
	if coverIdx > 0 {
		cover := manifestItems[coverIdx]
		manifestItems = append(manifestItems[:coverIdx], manifestItems[coverIdx+1:]...)
		manifestItems = append([]string{cover}, manifestItems...)
	}
	manifestBlock := string(manifestStart) + "\n" + strings.Join(manifestItems, "\n") + "\n" + string(manifestEnd)
	out = regexp.MustCompile(`<manifest[\s\S]*?</manifest>`).ReplaceAll(out, []byte(manifestBlock))

	// Write debug output for inspection
	// _ = os.WriteFile("/home/felix/src/kojirou/cmd/formats/epub/_debug_last_opf.xml", out, 0644)

	return out, nil
}
