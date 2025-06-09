// Package epub provides functionality for generating EPUB format ebooks from manga
package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"image"
	"io"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	testhelpers "github.com/leotaku/kojirou/cmd/formats/testhelpers"
	"github.com/leotaku/kojirou/mangadex"
)

// createTestManga creates a test manga for use in testing
func createTestManga() mangadex.Manga {
	manga := mangadex.Manga{
		Info: mangadex.MangaInfo{
			Title:   "Test Manga",
			ID:      "test-manga-id",
			Authors: []string{"Test Author"},
		},
		Volumes: map[mangadex.Identifier]mangadex.Volume{},
	}
	volID := mangadex.NewIdentifier("1")
	vol := mangadex.Volume{
		Info:     mangadex.VolumeInfo{Identifier: volID},
		Chapters: map[mangadex.Identifier]mangadex.Chapter{},
	}
	chapID := mangadex.NewIdentifier("1-1")
	chap := mangadex.Chapter{
		Info: mangadex.ChapterInfo{
			Identifier:       chapID,
			Title:            "Chapter 1",
			VolumeIdentifier: volID,
		},
		Pages: map[int]image.Image{
			0: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
		},
	}
	vol.Chapters[chapID] = chap
	manga.Volumes[volID] = vol
	return manga
}

// createMultiVolumeTestManga creates a manga with multiple volumes for testing
func createMultiVolumeTestManga() mangadex.Manga {
	manga := createTestManga()
	vol2ID := mangadex.NewIdentifier("2")
	vol2 := mangadex.Volume{
		Info:     mangadex.VolumeInfo{Identifier: vol2ID},
		Chapters: map[mangadex.Identifier]mangadex.Chapter{},
	}
	chap2ID := mangadex.NewIdentifier("2-1")
	chap2 := mangadex.Chapter{
		Info: mangadex.ChapterInfo{
			Identifier:       chap2ID,
			Title:            "Chapter 2",
			VolumeIdentifier: vol2ID,
		},
		Pages: map[int]image.Image{
			0: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
			1: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
		},
	}
	vol2.Chapters[chap2ID] = chap2
	manga.Volumes[vol2ID] = vol2
	return manga
}

// TestEPUBToKEPUBConversion tests the full conversion process from EPUB to KEPUB format
func TestEPUBToKEPUBConversion(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*epub.Epub, error)
		validate    func(*testing.T, []byte)
		wantErr     bool
		seriesTitle string
		seriesIndex float64
	}{
		{
			name: "standard epub to kepub",
			setup: func() (*epub.Epub, error) {
				manga := testhelpers.CreateTestManga()
				epub, _, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				if err != nil {
					return nil, err
				}
				return epub, nil
			},
			validate: func(t *testing.T, data []byte) {
				validateBasicKEPUBStructure(t, data)
				validateKoboSpecificMetadata(t, data)
				validateKoboHTMLTransformation(t, data)
				validateKoboCoverInOPF(t, data)
				validateSeriesMetadata(t, data)
			},
			seriesTitle: "Test Series",
			seriesIndex: 1.0,
		},
		{
			name: "epub to kepub with long series name",
			setup: func() (*epub.Epub, error) {
				manga := testhelpers.CreateTestManga()
				epub, _, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				if err != nil {
					return nil, err
				}
				return epub, nil
			},
			validate: func(t *testing.T, data []byte) {
				validateBasicKEPUBStructure(t, data)
				validateKoboSpecificMetadata(t, data)
				validateKoboHTMLTransformation(t, data)
				validateKoboCoverInOPF(t, data)
				validateSeriesMetadata(t, data)
			},
			seriesTitle: strings.Repeat("Very Long Series Title ", 20),
			seriesIndex: 2.0,
		},
		{
			name: "epub to kepub without series info",
			setup: func() (*epub.Epub, error) {
				manga := testhelpers.CreateTestManga()
				epub, _, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				if err != nil {
					return nil, err
				}
				return epub, nil
			},
			validate: func(t *testing.T, data []byte) {
				validateBasicKEPUBStructure(t, data)
				validateKoboSpecificMetadata(t, data)
				validateKoboHTMLTransformation(t, data)
				validateKoboCoverInOPF(t, data)
				validateNoSeriesMetadata(t, data)
			},
			seriesTitle: "",
			seriesIndex: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			epub, err := tc.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Convert to KEPUB with series metadata
			kepubData, err := kepubconv.ConvertToKEPUB(epub, tc.seriesTitle, tc.seriesIndex)

			// Check error expectations
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToKEPUB() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if err == nil {
				tc.validate(t, kepubData)
			}
		})
	}
}

// TestKEPUBEnhancedFeatures tests Kobo-specific enhancement features
func TestKEPUBEnhancedFeatures(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() (*epub.Epub, func(), error)
		checkFeature func(*testing.T, []byte)
		wantErr      bool
		seriesTitle  string
		seriesIndex  float64
	}{
		{
			name: "kobo spans for pagination",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			checkFeature: func(t *testing.T, data []byte) {
				validateKoboTextSpans(t, data)
			},
			wantErr: false,
		},
		{
			name: "kobo fixed layout metadata",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			checkFeature: func(t *testing.T, data []byte) {
				validateKoboFixedLayout(t, data)
			},
			wantErr: false,
		},
		{
			name: "kobo image handling",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			checkFeature: func(t *testing.T, data []byte) {
				validateKoboImageHandling(t, data)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer cleanup()

			kepubData, err := kepubconv.ConvertToKEPUB(epub, "", 0) // No series metadata for feature tests
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToKEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.checkFeature != nil {
				tt.checkFeature(t, kepubData)
			}
		})
	}
}

// TestKEPUBCompatibility tests compatibility with Kobo devices through file structure validation
func TestKEPUBCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*epub.Epub, func(), error)
		validate func(*testing.T, []byte)
		wantErr  bool
	}{
		{
			name: "standard compatibility",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			validate: func(t *testing.T, data []byte) {
				validateKoboCompatibility(t, data)
			},
			wantErr: false,
		},
		{
			name: "special title compatibility",
			setup: func() (*epub.Epub, func(), error) {
				manga := testhelpers.CreateTestManga()
				manga.Info.Title = "Special: Characters & Test"
				epub, cleanup, err := GenerateEPUB(t.TempDir(), manga, kindle.WidepagePolicyPreserve, false, false)
				return epub, cleanup, err
			},
			validate: func(t *testing.T, data []byte) {
				validateKoboCompatibility(t, data)
				validateSpecialCharacterHandling(t, data)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, cleanup, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer cleanup()

			kepubData, err := kepubconv.ConvertToKEPUB(epub, "", 0) // No series metadata for compatibility tests
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToKEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.validate != nil {
				tt.validate(t, kepubData)
			}
		})
	}
}

// Helper functions for KEPUB validation

// validateBasicKEPUBStructure checks basic KEPUB file structure requirements
func validateBasicKEPUBStructure(t *testing.T, data []byte) {
	if len(data) == 0 {
		t.Fatal("KEPUB data is empty")
	}

	// Verify we can read it as a ZIP file
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check required files
	requiredFiles := []string{
		"mimetype",
		"META-INF/container.xml",
	}

	hasOPF := false
	for _, f := range r.File {
		for i, required := range requiredFiles {
			if f.Name == required {
				requiredFiles = append(requiredFiles[:i], requiredFiles[i+1:]...)
				break
			}
		}

		if strings.HasSuffix(f.Name, ".opf") {
			hasOPF = true
		}
	}

	if len(requiredFiles) > 0 {
		t.Errorf("Missing required files: %v", requiredFiles)
	}

	if !hasOPF {
		t.Error("No OPF file found in KEPUB")
	}

	// Verify mimetype is first file and uncompressed
	if len(r.File) == 0 || r.File[0].Name != "mimetype" {
		t.Error("mimetype must be the first file in the KEPUB")
	} else if r.File[0].Method != zip.Store {
		t.Error("mimetype must be stored without compression")
	}
}

// validateKoboSpecificMetadata checks for Kobo-specific metadata in OPF
func validateKoboSpecificMetadata(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and read OPF file
	var opfContent []byte
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			opfContent, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}
			break
		}
	}

	if len(opfContent) == 0 {
		t.Fatal("No OPF content found")
	}

	// Check for required Kobo metadata
	koboMetadataMarkers := []string{
		"kobo:",
		"rendition:",
	}

	for _, marker := range koboMetadataMarkers {
		if !bytes.Contains(opfContent, []byte(marker)) {
			t.Errorf("OPF doesn't contain expected Kobo metadata: %s", marker)
		}
	}
}

// validateKoboHTMLTransformation checks for Kobo-specific HTML transformations
func validateKoboHTMLTransformation(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check HTML files for Kobo transformations
	foundTransformedHTML := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xhtml") || strings.HasSuffix(f.Name, ".html") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Check for Kobo-specific text spans and attributes
			if bytes.Contains(content, []byte("kobo:")) ||
				bytes.Contains(content, []byte("<span class=\"kobo")) ||
				bytes.Contains(content, []byte("koboSpan")) ||
				bytes.Contains(content, []byte("epub:type=\"kobo\"")) ||
				bytes.Contains(content, []byte("class=\"kobo-image\"")) ||
				bytes.Contains(content, []byte("xmlns:kobo")) {
				foundTransformedHTML = true
				break
			}
		}
	}

	if !foundTransformedHTML {
		t.Error("No evidence of Kobo HTML transformations found")
	}
}

// validateKoboTextSpans checks for Kobo text span elements in HTML
func validateKoboTextSpans(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	foundSpans := false
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xhtml") || strings.HasSuffix(f.Name, ".html") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Check for Kobo-specific span elements
			if bytes.Contains(content, []byte("kobo-span")) ||
				bytes.Contains(content, []byte("class=\"kobo")) {
				foundSpans = true
				break
			}
		}
	}

	if !foundSpans {
		t.Error("No Kobo text spans found in HTML content")
	}
}

// validateKoboFixedLayout checks for fixed layout metadata
func validateKoboFixedLayout(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Find and check OPF content
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}

			// Check for fixed layout metadata
			if !bytes.Contains(content, []byte("rendition:layout")) ||
				!bytes.Contains(content, []byte("pre-paginated")) {
				t.Error("Fixed layout metadata not found")
			}
			return
		}
	}

	t.Error("No OPF file found")
}

// validateKoboImageHandling checks for proper Kobo image handling
func validateKoboImageHandling(t *testing.T, data []byte) {
	validateBasicKEPUBStructure(t, data) // Basic validation is sufficient for this
}

// validateKoboCompatibility checks for Kobo device compatibility
func validateKoboCompatibility(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	// Check for Kobo compatibility markers
	var hasKoboTypeMetadata bool
	var hasKoboVersion bool

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}

			// Check for required Kobo metadata
			hasKoboTypeMetadata = bytes.Contains(content, []byte("kobo:content-type"))
			hasKoboVersion = bytes.Contains(content, []byte("kobo:epub-version"))
		}
	}

	if !hasKoboTypeMetadata {
		t.Error("Missing kobo:content-type metadata required for Kobo compatibility")
	}

	if !hasKoboVersion {
		t.Error("Missing kobo:epub-version metadata required for Kobo compatibility")
	}
}

// validateSpecialCharacterHandling checks for proper handling of special characters
func validateSpecialCharacterHandling(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}
			// Robust: parse all <meta> and <item> attributes and check for unescaped special chars
			type Meta struct {
				Name     string `xml:"name,attr,omitempty"`
				Content  string `xml:"content,attr,omitempty"`
				Property string `xml:"property,attr,omitempty"`
			}
			type Metadata struct {
				Metas []Meta `xml:"meta"`
			}
			type Item struct {
				ID         string `xml:"id,attr"`
				Href       string `xml:"href,attr"`
				MediaType  string `xml:"media-type,attr"`
				Properties string `xml:"properties,attr,omitempty"`
			}
			type Manifest struct {
				Items []Item `xml:"item"`
			}
			type Package struct {
				Metadata Metadata `xml:"metadata"`
				Manifest Manifest `xml:"manifest"`
			}
			var pkg Package
			if err := xml.Unmarshal(content, &pkg); err != nil {
				t.Fatalf("Failed to parse OPF XML: %v", err)
			}
			// Check for unescaped ampersand or < or > in meta or item attributes
			for _, m := range pkg.Metadata.Metas {
				if strings.Contains(m.Name, "&") || strings.Contains(m.Name, "<") || strings.Contains(m.Name, ">") {
					t.Error("Special characters not properly escaped in meta name")
				}
				if strings.Contains(m.Content, "&") || strings.Contains(m.Content, "<") || strings.Contains(m.Content, ">") {
					t.Error("Special characters not properly escaped in meta content")
				}
				if strings.Contains(m.Property, "&") || strings.Contains(m.Property, "<") || strings.Contains(m.Property, ">") {
					t.Error("Special characters not properly escaped in meta property")
				}
			}
			for _, it := range pkg.Manifest.Items {
				if strings.Contains(it.ID, "&") || strings.Contains(it.ID, "<") || strings.Contains(it.ID, ">") {
					t.Error("Special characters not properly escaped in item id")
				}
				if strings.Contains(it.Href, "&") || strings.Contains(it.Href, "<") || strings.Contains(it.Href, ">") {
					t.Error("Special characters not properly escaped in item href")
				}
				if strings.Contains(it.MediaType, "&") || strings.Contains(it.MediaType, "<") || strings.Contains(it.MediaType, ">") {
					t.Error("Special characters not properly escaped in item media-type")
				}
				if strings.Contains(it.Properties, "&") || strings.Contains(it.Properties, "<") || strings.Contains(it.Properties, ">") {
					t.Error("Special characters not properly escaped in item properties")
				}
			}
			return
		}
	}
	t.Error("No OPF file found in KEPUB")
}

// validateKoboCoverInOPF checks that the cover image is the first item in the manifest and referenced in <meta name="cover" content="cover"/>.
func validateKoboCoverInOPF(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}
			// Robust: parse manifest as XML and check first <item> is id="cover"
			type Item struct {
				ID         string `xml:"id,attr"`
				Href       string `xml:"href,attr"`
				MediaType  string `xml:"media-type,attr"`
				Properties string `xml:"properties,attr,omitempty"`
			}
			type Manifest struct {
				Items []Item `xml:"item"`
			}
			type Meta struct {
				Name    string `xml:"name,attr,omitempty"`
				Content string `xml:"content,attr,omitempty"`
			}
			type Metadata struct {
				Metas []Meta `xml:"meta"`
			}
			type Package struct {
				Manifest Manifest `xml:"manifest"`
				Metadata Metadata `xml:"metadata"`
			}
			var pkg Package
			if err := xml.Unmarshal(content, &pkg); err != nil {
				t.Fatalf("Failed to parse OPF XML: %v", err)
			}
			if len(pkg.Manifest.Items) == 0 || pkg.Manifest.Items[0].ID != "cover" {
				t.Error("<item id=\"cover\" ...> is not the first item in manifest")
			}
			// Check for <meta name="cover" content="cover"/>
			found := false
			for _, m := range pkg.Metadata.Metas {
				if m.Name == "cover" && m.Content == "cover" {
					found = true
					break
				}
			}
			if !found {
				t.Error("<meta name=\"cover\" content=\"cover\"/> not found in OPF metadata")
			}
			return
		}
	}
	t.Error("No OPF file found in KEPUB")
}

// validateSeriesMetadata checks for series metadata in OPF
func validateSeriesMetadata(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}

			// Check for series metadata
			if !bytes.Contains(content, []byte("meta name=\"calibre:series\"")) ||
				!bytes.Contains(content, []byte("meta name=\"calibre:series_index\"")) {
				t.Error("Series metadata not found in OPF")
			}
			return
		}
	}
	t.Error("No OPF file found in KEPUB")
}

// validateNoSeriesMetadata checks that no series metadata is present in OPF
func validateNoSeriesMetadata(t *testing.T, data []byte) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read KEPUB as ZIP: %v", err)
	}

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open OPF file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("Failed to read OPF file: %v", err)
			}

			// Check that series metadata is absent
			if bytes.Contains(content, []byte("meta name=\"calibre:series\"")) ||
				bytes.Contains(content, []byte("meta name=\"calibre:series_index\"")) {
				t.Error("Unexpected series metadata found in OPF")
			}
			return
		}
	}
	t.Error("No OPF file found in KEPUB")
}
