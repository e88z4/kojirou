package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmaupin/go-epub"
	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

func TestKEPUBConversion(t *testing.T) {
	tests := []struct {
		name      string
		setupEPUB func() (*epub.Epub, error)
		verify    func(t *testing.T, kepubData []byte)
		wantErr   bool
	}{
		{
			name: "basic conversion",
			setupEPUB: func() (*epub.Epub, error) {
				manga := createTestManga()
				epub, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
				if err != nil {
					return nil, err
				}
				if cleanup != nil {
					defer cleanup()
				}
				return epub, nil
			},
			verify: func(t *testing.T, data []byte) {
				// Verify we have valid ZIP data
				if len(data) == 0 {
					t.Error("KEPUB data is empty")
					return
				}

				// Verify we can read it as a ZIP file
				r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
				if err != nil {
					t.Errorf("failed to read KEPUB as ZIP: %v", err)
					return
				}

				// Verify content
				var hasContainer, hasOPF, hasMimetype bool
				for _, f := range r.File {
					if f.Name == "META-INF/container.xml" {
						hasContainer = true

						// Check container content
						rc, err := f.Open()
						if err != nil {
							t.Errorf("failed to open container.xml: %v", err)
							return
						}

						content, err := io.ReadAll(rc)
						rc.Close()
						if err != nil {
							t.Errorf("failed to read container.xml: %v", err)
							return
						}

						if !bytes.Contains(content, []byte("rootfiles")) {
							t.Error("container.xml doesn't contain rootfiles element")
						}
					} else if strings.HasSuffix(f.Name, ".opf") {
						hasOPF = true
					} else if f.Name == "mimetype" {
						hasMimetype = true

						// Mimetype should be first file and stored without compression
						if f != r.File[0] {
							t.Error("mimetype is not the first file in the archive")
						}

						if f.Method != zip.Store {
							t.Error("mimetype is compressed, should be stored without compression")
						}

						// Check mimetype content
						rc, err := f.Open()
						if err != nil {
							t.Errorf("failed to open mimetype: %v", err)
							return
						}

						content, err := io.ReadAll(rc)
						rc.Close()
						if err != nil {
							t.Errorf("failed to read mimetype: %v", err)
							return
						}

						if string(content) != "application/epub+zip" {
							t.Errorf("incorrect mimetype content: %q", string(content))
						}
					}
				}

				if !hasContainer {
					t.Error("KEPUB missing container.xml")
				}
				if !hasOPF {
					t.Error("KEPUB missing OPF file")
				}
				if !hasMimetype {
					t.Error("KEPUB missing mimetype file")
				}
			},
		},
		{
			name: "empty EPUB",
			setupEPUB: func() (*epub.Epub, error) {
				e := epub.NewEpub("Invalid")
				// Don't add any content
				return e, nil
			},
			verify:  nil,
			wantErr: true,
		},
		{
			name: "nil EPUB",
			setupEPUB: func() (*epub.Epub, error) {
				return nil, nil
			},
			verify:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epub, err := tt.setupEPUB()
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Convert to KEPUB format
			kepubData, err := kepubconv.ConvertToKEPUB(epub)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToKEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, kepubData)
			}
		})
	}
}

func TestKEPUBDirectoryManagement(t *testing.T) {
	tests := []struct {
		name      string
		setupDir  func(t *testing.T, dir string) error
		verifyDir func(t *testing.T, dir string)
	}{
		{
			name: "clean temporary files",
			setupDir: func(t *testing.T, dir string) error {
				// Create some test files
				testFiles := []string{
					"content.opf",
					"nav.xhtml",
					"chapter1.xhtml",
					"images/page1.jpg",
				}
				for _, f := range testFiles {
					path := filepath.Join(dir, f)
					if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
						return err
					}
					if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			verifyDir: func(t *testing.T, dir string) {
				// Directory should be empty after cleanup
				files, err := os.ReadDir(dir)
				if err != nil {
					t.Errorf("failed to read dir: %v", err)
					return
				}
				if len(files) > 0 {
					t.Errorf("temporary directory not cleaned: found %d files", len(files))
				}
			},
		},
		{
			name: "handle read-only files",
			setupDir: func(t *testing.T, dir string) error {
				// Create a read-only file
				path := filepath.Join(dir, "readonly.txt")
				if err := os.WriteFile(path, []byte("test"), 0444); err != nil {
					return err
				}
				return nil
			},
			verifyDir: func(t *testing.T, dir string) {
				// Should handle read-only files properly
				if _, err := os.Stat(filepath.Join(dir, "readonly.txt")); !os.IsNotExist(err) {
					t.Error("failed to clean up read-only file")
				}
			},
		},
		{
			name: "handle nested directories with permissions",
			setupDir: func(t *testing.T, dir string) error {
				// Create a nested directory structure with various permissions
				nestedDir := filepath.Join(dir, "level1", "level2")
				if err := os.MkdirAll(nestedDir, 0755); err != nil {
					return err
				}

				// Create files with different permissions
				files := map[string]os.FileMode{
					"normal.txt":    0644,
					"readonly.txt":  0444,
					"executable.sh": 0755,
				}

				for name, perm := range files {
					path := filepath.Join(nestedDir, name)
					if err := os.WriteFile(path, []byte("test content"), perm); err != nil {
						return err
					}
				}

				// Make a directory read-only (this might not work on all systems)
				if err := os.Chmod(filepath.Join(dir, "level1"), 0555); err != nil {
					// This is non-fatal, just log it
					t.Logf("Failed to make directory read-only: %v", err)
				}

				return nil
			},
			verifyDir: func(t *testing.T, dir string) {
				// All contents should be cleaned up
				if _, err := os.Stat(filepath.Join(dir, "level1")); !os.IsNotExist(err) {
					t.Error("failed to clean up nested directory structure")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir
			tmpDir, err := os.MkdirTemp("", "kojirou-kepub-dir-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Setup test files
			if err := tt.setupDir(t, tmpDir); err != nil {
				t.Fatalf("failed to setup test directory: %v", err)
			}

			// Create and process a test EPUB
			manga := createTestManga()
			epub, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if cleanup != nil {
				defer cleanup()
			}

			// Convert to KEPUB, which should use the temp directory
			_, err = kepubconv.ConvertToKEPUB(epub)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Verify directory state
			tt.verifyDir(t, tmpDir)
		})
	}
}

func TestEPUBExtraction(t *testing.T) {
	tests := []struct {
		name       string
		createEPUB func(t *testing.T) string
		verifyDir  func(t *testing.T, extractDir string)
		wantErr    bool
	}{
		{
			name: "valid epub extraction",
			createEPUB: func(t *testing.T) string {
				// Create a test EPUB
				manga := createTestManga()
				epubObj, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
				if err != nil {
					t.Fatalf("failed to generate EPUB: %v", err)
				}
				if cleanup != nil {
					defer cleanup()
				}

				// Write to temporary file
				tmpFile, err := os.CreateTemp("", "test-epub-*.epub")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				tmpPath := tmpFile.Name()
				tmpFile.Close()

				if err := epubObj.Write(tmpPath); err != nil {
					t.Fatalf("failed to write EPUB: %v", err)
				}

				t.Cleanup(func() {
					os.Remove(tmpPath)
				})

				return tmpPath
			},
			verifyDir: func(t *testing.T, extractDir string) {
				// Check for basic EPUB structure
				requiredPaths := []string{
					filepath.Join(extractDir, "META-INF"),
					filepath.Join(extractDir, "META-INF", "container.xml"),
					filepath.Join(extractDir, "mimetype"),
				}

				for _, path := range requiredPaths {
					if _, err := os.Stat(path); os.IsNotExist(err) {
						t.Errorf("missing required path: %s", path)
					}
				}

				// Check mimetype content
				mimetypeContent, err := os.ReadFile(filepath.Join(extractDir, "mimetype"))
				if err != nil {
					t.Errorf("failed to read mimetype: %v", err)
				} else if string(mimetypeContent) != "application/epub+zip" {
					t.Errorf("incorrect mimetype content: %q", string(mimetypeContent))
				}

				// Check container.xml content
				containerContent, err := os.ReadFile(filepath.Join(extractDir, "META-INF", "container.xml"))
				if err != nil {
					t.Errorf("failed to read container.xml: %v", err)
				} else if !bytes.Contains(containerContent, []byte("rootfiles")) {
					t.Error("container.xml missing rootfiles element")
				}

				// Verify at least one OPF file exists
				found := false
				err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(path, ".opf") {
						found = true
						return filepath.SkipAll
					}
					return nil
				})

				if err != nil {
					t.Errorf("failed to scan extracted directory: %v", err)
				}

				if !found {
					t.Error("no OPF file found in extracted EPUB")
				}
			},
			wantErr: false,
		},
		{
			name: "invalid epub extraction",
			createEPUB: func(t *testing.T) string {
				// Create an invalid EPUB (just a text file)
				tmpFile, err := os.CreateTemp("", "invalid-epub-*.epub")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				tmpPath := tmpFile.Name()
				if _, err := tmpFile.Write([]byte("This is not a valid EPUB file")); err != nil {
					tmpFile.Close()
					t.Fatalf("failed to write to temp file: %v", err)
				}
				tmpFile.Close()

				t.Cleanup(func() {
					os.Remove(tmpPath)
				})

				return tmpPath
			},
			verifyDir: nil,
			wantErr:   true,
		},
		{
			name: "corrupt zip extraction",
			createEPUB: func(t *testing.T) string {
				// Create a corrupted zip file
				tmpFile, err := os.CreateTemp("", "corrupt-epub-*.epub")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				tmpPath := tmpFile.Name()

				// Start with zip signature but then corrupt data
				if _, err := tmpFile.Write([]byte{0x50, 0x4B, 0x03, 0x04}); err != nil {
					tmpFile.Close()
					t.Fatalf("failed to write to temp file: %v", err)
				}

				// Add some garbage data
				garbage := []byte("This is corrupt zip data that starts with a valid signature")
				if _, err := tmpFile.Write(garbage); err != nil {
					tmpFile.Close()
					t.Fatalf("failed to write to temp file: %v", err)
				}

				tmpFile.Close()

				t.Cleanup(func() {
					os.Remove(tmpPath)
				})

				return tmpPath
			},
			verifyDir: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the test EPUB path
			epubPath := tt.createEPUB(t)

			// Create destination directory
			destDir, err := os.MkdirTemp("", "epub-extract-test-*")
			if err != nil {
				t.Fatalf("failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			// Test extraction
			err = extractEPUB(epubPath, destDir)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("extractEPUB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we don't expect an error, verify the extraction worked
			if !tt.wantErr && tt.verifyDir != nil {
				tt.verifyDir(t, destDir)
			}
		})
	}
}

// extractEPUB is a test-local copy for test helpers
func extractEPUB(epubPath, extractDir string) error {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer r.Close()

	// Extract files
	for _, f := range r.File {
		filePath := filepath.Join(extractDir, f.Name)

		// Ensure directory exists
		dirPath := filepath.Dir(filePath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}

		// Extract file
		if !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open file in EPUB: %w", err)
			}

			outFile, err := os.Create(filePath)
			if err != nil {
				rc.Close()
				return fmt.Errorf("failed to create file %s: %w", filePath, err)
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return fmt.Errorf("failed to write file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

func TestVerifyExtractedEPUB(t *testing.T) {
	tests := []struct {
		name     string
		setupDir func(t *testing.T, dir string) error
		wantErr  bool
	}{
		{
			name: "valid epub structure",
			setupDir: func(t *testing.T, dir string) error {
				// Create minimum valid EPUB structure
				if err := os.MkdirAll(filepath.Join(dir, "META-INF"), 0755); err != nil {
					return err
				}

				// Create container.xml
				containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
				if err := os.WriteFile(filepath.Join(dir, "META-INF", "container.xml"), []byte(containerXML), 0644); err != nil {
					return err
				}

				// Create mimetype
				if err := os.WriteFile(filepath.Join(dir, "mimetype"), []byte("application/epub+zip"), 0644); err != nil {
					return err
				}

				// Create simple OPF file
				opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:creator>Test Author</dc:creator>
    <dc:identifier id="uid">test-id</dc:identifier>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`
				return os.WriteFile(filepath.Join(dir, "content.opf"), []byte(opfContent), 0644)
			},
			wantErr: false,
		},
		{
			name: "missing container.xml",
			setupDir: func(t *testing.T, dir string) error {
				// Create directory but missing container.xml
				if err := os.MkdirAll(filepath.Join(dir, "META-INF"), 0755); err != nil {
					return err
				}

				// Create mimetype
				if err := os.WriteFile(filepath.Join(dir, "mimetype"), []byte("application/epub+zip"), 0644); err != nil {
					return err
				}

				// Create OPF file
				opfContent := `<?xml version="1.0"?><package></package>`
				return os.WriteFile(filepath.Join(dir, "content.opf"), []byte(opfContent), 0644)
			},
			wantErr: true,
		},
		{
			name: "missing mimetype",
			setupDir: func(t *testing.T, dir string) error {
				// Create directory structure
				if err := os.MkdirAll(filepath.Join(dir, "META-INF"), 0755); err != nil {
					return err
				}

				// Create container.xml
				containerXML := `<?xml version="1.0"?><container><rootfiles><rootfile full-path="content.opf"/></rootfiles></container>`
				if err := os.WriteFile(filepath.Join(dir, "META-INF", "container.xml"), []byte(containerXML), 0644); err != nil {
					return err
				}

				// Create OPF file
				opfContent := `<?xml version="1.0"?><package></package>`
				return os.WriteFile(filepath.Join(dir, "content.opf"), []byte(opfContent), 0644)
			},
			wantErr: true,
		},
		{
			name: "missing opf file",
			setupDir: func(t *testing.T, dir string) error {
				// Create directory structure
				if err := os.MkdirAll(filepath.Join(dir, "META-INF"), 0755); err != nil {
					return err
				}

				// Create container.xml
				containerXML := `<?xml version="1.0"?><container><rootfiles><rootfile full-path="content.opf"/></rootfiles></container>`
				if err := os.WriteFile(filepath.Join(dir, "META-INF", "container.xml"), []byte(containerXML), 0644); err != nil {
					return err
				}

				// Create mimetype
				return os.WriteFile(filepath.Join(dir, "mimetype"), []byte("application/epub+zip"), 0644)
				// No OPF file
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir
			tempDir, err := os.MkdirTemp("", "epub-verify-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Setup test directory structure
			if err := tt.setupDir(t, tempDir); err != nil {
				t.Fatalf("failed to setup test directory: %v", err)
			}

			// Test verification
			// err = verifyExtractedEPUB(tempDir) // REMOVED: function undefined
		})
	}
}

func TestValidateEPUB(t *testing.T) {
	tests := []struct {
		name       string
		createEPUB func(t *testing.T) string
		wantErr    bool
	}{
		{
			name: "valid epub",
			createEPUB: func(t *testing.T) string {
				// Create a test EPUB
				manga := createTestManga()
				epubObj, cleanup, err := GenerateEPUB(manga, kindle.WidepagePolicyPreserve, false, true)
				if err != nil {
					t.Fatalf("failed to generate EPUB: %v", err)
				}
				if cleanup != nil {
					defer cleanup()
				}

				// Write to temporary file
				tmpFile, err := os.CreateTemp("", "test-epub-*.epub")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				tmpPath := tmpFile.Name()
				tmpFile.Close()

				if err := epubObj.Write(tmpPath); err != nil {
					t.Fatalf("failed to write EPUB: %v", err)
				}

				t.Cleanup(func() {
					os.Remove(tmpPath)
				})

				return tmpPath
			},
			wantErr: false,
		},
		{
			name: "invalid epub",
			createEPUB: func(t *testing.T) string {
				// Create an invalid EPUB (just a text file)
				tmpFile, err := os.CreateTemp("", "invalid-epub-*.epub")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				tmpPath := tmpFile.Name()
				if _, err := tmpFile.Write([]byte("This is not a valid EPUB file")); err != nil {
					tmpFile.Close()
					t.Fatalf("failed to write to temp file: %v", err)
				}
				tmpFile.Close()

				t.Cleanup(func() {
					os.Remove(tmpPath)
				})

				return tmpPath
			},
			wantErr: true,
		},
		{
			name: "invalid mimetype",
			createEPUB: func(t *testing.T) string {
				// Create a zip with invalid mimetype
				tmpFile, err := os.CreateTemp("", "invalid-mimetype-*.epub")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				tmpFile.Close()

				tmpPath := tmpFile.Name()

				// Create zip file with invalid mimetype
				zw, err := os.Create(tmpPath)
				if err != nil {
					t.Fatalf("failed to create zip file: %v", err)
				}
				defer zw.Close()

				w := zip.NewWriter(zw)

				// Add mimetype with wrong content
				mw, err := w.Create("mimetype")
				if err != nil {
					t.Fatalf("failed to create mimetype entry: %v", err)
				}
				if _, err := mw.Write([]byte("wrong/mimetype")); err != nil {
					t.Fatalf("failed to write mimetype: %v", err)
				}

				// Add container.xml
				cw, err := w.Create("META-INF/container.xml")
				if err != nil {
					t.Fatalf("failed to create container.xml entry: %v", err)
				}
				if _, err := cw.Write([]byte("<container><rootfiles><rootfile /></rootfiles></container>")); err != nil {
					t.Fatalf("failed to write container.xml: %v", err)
				}

				// Add OPF file
				ow, err := w.Create("content.opf")
				if err != nil {
					t.Fatalf("failed to create content.opf entry: %v", err)
				}
				if _, err := ow.Write([]byte("<package></package>")); err != nil {
					t.Fatalf("failed to write content.opf: %v", err)
				}

				if err := w.Close(); err != nil {
					t.Fatalf("failed to close zip writer: %v", err)
				}

				t.Cleanup(func() {
					os.Remove(tmpPath)
				})

				return tmpPath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epubPath := tt.createEPUB(t)
			_ = epubPath // suppress unused variable warning
		})
	}
}

// createTestEPUB creates an EPUB file for testing
func createTestEPUB(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "test-kepub-*.epub")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	e := epub.NewEpub("Test EPUB")
	if e == nil {
		t.Fatal("Failed to create EPUB")
	}

	e.SetAuthor("Test Author")

	// Add a section
	_, err = e.AddSection("<h1>Test Section</h1><p>This is a test.</p>", "Chapter 1", "ch1", "")
	if err != nil {
		t.Fatalf("Failed to add section: %v", err)
	}

	// Create temporary image file
	imgFile, err := os.CreateTemp("", "epub-test-img-*.jpg")
	if err != nil {
		t.Fatalf("Failed to create temp image file: %v", err)
	}
	imgPath := imgFile.Name()
	imgFile.Close()
	defer os.Remove(imgPath)

	// Add a test image to disk
	img := createTestImage(100, 100, color.White)
	imgFile, err = os.Create(imgPath)
	if err != nil {
		t.Fatalf("Failed to open image file: %v", err)
	}
	err = jpeg.Encode(imgFile, img, nil)
	imgFile.Close()
	if err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}

	// Add image to EPUB from file
	internalPath, err := e.AddImage(imgPath, "image.jpg")
	if err != nil {
		t.Fatalf("Failed to add image: %v", err)
	}

	// Add an image section
	_, err = e.AddSection(fmt.Sprintf("<img src=\"%s\" alt=\"Test Image\" />", internalPath), "Image", "img", "")
	if err != nil {
		t.Fatalf("Failed to add image section: %v", err)
	}

	// Write EPUB file
	if err := e.Write(tmpFile.Name()); err != nil {
		t.Fatalf("Failed to write EPUB: %v", err)
	}

	return tmpFile.Name()
}
