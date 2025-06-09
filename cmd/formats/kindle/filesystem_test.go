package kindle

import (
	"os"
	"path"
	"strings"
	"testing"

	md "github.com/leotaku/kojirou/mangadex"
)

func TestNormalizedDirectoryPath(t *testing.T) {
	// Setup temporary test directory
	testDir := t.TempDir()

	// Create a normalized directory
	dir := NewNormalizedDirectory(testDir, "Test Manga", false)

	// Test identifiers
	identifier := md.NewIdentifier("1.5")

	// Test Path method with different extensions
	epubPath := dir.Path(identifier, "epub")
	expectedEpubPath := path.Join(testDir, "0001.05.epub")
	if epubPath != expectedEpubPath {
		t.Errorf("Path for EPUB incorrect, got: %s, want: %s", epubPath, expectedEpubPath)
	}

	kepubPath := dir.Path(identifier, "kepub.epub")
	expectedKepubPath := path.Join(testDir, "0001.05.kepub.epub")
	if kepubPath != expectedKepubPath {
		t.Errorf("Path for KEPUB incorrect, got: %s, want: %s", kepubPath, expectedKepubPath)
	}

	mobiPath := dir.Path(identifier, "azw3")
	expectedMobiPath := path.Join(testDir, "0001.05.azw3")
	if mobiPath != expectedMobiPath {
		t.Errorf("Path for MOBI incorrect, got: %s, want: %s", mobiPath, expectedMobiPath)
	}
}

func TestHasWithExtension(t *testing.T) {
	// Setup temporary test directory
	testDir := t.TempDir()

	// Create a normalized directory
	dir := NewNormalizedDirectory(testDir, "Test Manga", false)

	// Test identifiers
	identifier := md.NewIdentifier("1.5")

	// Create test file
	epubPath := path.Join(testDir, "0001.05.epub")
	err := os.WriteFile(epubPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test Has and HasWithExtension methods
	if !dir.Has(identifier) {
		t.Error("Has should return true when a file exists")
	}

	if !dir.HasWithExtension(identifier, "epub") {
		t.Error("HasWithExtension should return true for existing EPUB")
	}

	if dir.HasWithExtension(identifier, "kepub.epub") {
		t.Error("HasWithExtension should return false for non-existing KEPUB")
	}
}

func TestGetExistingFormats(t *testing.T) {
	// Setup temporary test directory
	testDir := t.TempDir()

	// Create a normalized directory
	dir := NewNormalizedDirectory(testDir, "Test Manga", false)

	// Test identifiers
	identifier := md.NewIdentifier("1.5")

	// Create test files
	epubPath := path.Join(testDir, "0001.05.epub")
	if err := os.WriteFile(epubPath, []byte("epub test"), 0644); err != nil {
		t.Fatalf("Failed to create EPUB test file: %v", err)
	}

	kepubPath := path.Join(testDir, "0001.05.kepub.epub")
	if err := os.WriteFile(kepubPath, []byte("kepub test"), 0644); err != nil {
		t.Fatalf("Failed to create KEPUB test file: %v", err)
	}

	// Test GetExistingFormats
	formats := dir.GetExistingFormats(identifier)

	if len(formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(formats))
	}

	if formats["epub"] != epubPath {
		t.Errorf("Expected epub path %s, got %s", epubPath, formats["epub"])
	}

	if formats["kepub.epub"] != kepubPath {
		t.Errorf("Expected kepub path %s, got %s", kepubPath, formats["kepub.epub"])
	}

	if _, exists := formats["azw3"]; exists {
		t.Error("AZW3 format should not exist")
	}
}

func TestPOSIXComplianceForKindlePaths(t *testing.T) {
	testDir := t.TempDir()
	specialTitle := "Test/Series: 01. " // includes slash, colon, space, dot
	identifier := md.NewIdentifier("1.5")

	dir := NewNormalizedDirectory(testDir, specialTitle, true)
	// Only check the last element (directory name) for forbidden characters
	bookDirBase := path.Base(dir.bookDirectory)
	forbidden := []string{"/", ":", " ", "."}
	for _, f := range forbidden {
		if strings.Contains(bookDirBase, f) {
			t.Errorf("bookDirectory base contains forbidden character '%s': %s", f, bookDirBase)
		}
	}
	// File name should be POSIX compliant (excluding extension)
	mobiPath := dir.Path(identifier, "azw3")
	fileBase := path.Base(mobiPath)
	fileName := fileBase
	if idx := strings.Index(fileBase, "."); idx != -1 {
		fileName = fileBase[:idx]
	}
	for _, f := range forbidden {
		if strings.Contains(fileName, f) {
			t.Errorf("MOBI file name contains forbidden character '%s': %s", f, fileName)
		}
	}
	// Should not be empty or reserved
	if bookDirBase == "" || bookDirBase == "." || bookDirBase == ".." {
		t.Errorf("bookDirectory base is reserved or empty: %s", bookDirBase)
	}
}
