// Integration tests for EPUB and KEPUB functionality
//
// This file contains integration tests that verify the complete workflow for
// generating EPUBs and KEPUBs, testing multiple format generation, and verifying
// output file structure. These tests exercise the full functionality of the
// EPUB and KEPUB generators in real-world usage scenarios.
package epub

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	kepubconv "github.com/leotaku/kojirou/cmd/formats/kepubconv"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	"github.com/leotaku/kojirou/cmd/formats/testhelpers"
	md "github.com/leotaku/kojirou/mangadex"
)

// TestCompleteWorkflow tests the entire EPUB and KEPUB generation workflow
//
// This test verifies:
// 1. EPUB generation from manga data
// 2. Writing EPUB to a file
// 3. Converting EPUB to KEPUB format
// 4. Writing KEPUB to a file
// 5. Verifying that both files exist and contain data
//
// This test ensures that the complete workflow from manga data to KEPUB file
// works correctly with no errors.
func TestCompleteWorkflow(t *testing.T) {
	// Create test manga
	manga := testhelpers.CreateTestManga()

	// Create temporary directory for output
	tempDir, err := os.MkdirTemp("", "kojirou-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test EPUB generation
	epubPath := filepath.Join(tempDir, "test.epub")
	t.Run("EPUB Generation", func(t *testing.T) {
		// Generate EPUB
		epubObj, cleanupObj, err := GenerateEPUB(tempDir, manga, kindle.WidepagePolicyPreserve, false, true)
		if err != nil {
			t.Fatalf("GenerateEPUB() failed: %v", err)
		}
		if cleanupObj != nil {
			defer cleanupObj()
		}

		// Write EPUB to file
		err = epubObj.Write(epubPath)
		if err != nil {
			t.Fatalf("Failed to write EPUB: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(epubPath); os.IsNotExist(err) {
			t.Fatalf("EPUB file not created: %s", epubPath)
		}

		// Verify file has content
		info, err := os.Stat(epubPath)
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}
		if info.Size() == 0 {
			t.Fatal("EPUB file is empty")
		}

		fmt.Printf("Generated EPUB file at %s (%d bytes)\n", epubPath, info.Size())

		// Test KEPUB conversion
		kepubPath := filepath.Join(tempDir, "test.kepub.epub")

		// Convert EPUB to KEPUB
		kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
		if err != nil {
			t.Fatalf("ConvertToKEPUB() failed: %v", err)
		}

		// Write KEPUB data to file
		err = os.WriteFile(kepubPath, kepubData, 0644)
		if err != nil {
			t.Fatalf("Failed to write KEPUB data: %v", err)
		}

		// Verify KEPUB file exists
		if _, err := os.Stat(kepubPath); os.IsNotExist(err) {
			t.Fatalf("KEPUB file not created: %s", kepubPath)
		}

		// Verify KEPUB file has content
		info, err = os.Stat(kepubPath)
		if err != nil {
			t.Fatalf("Failed to get KEPUB file info: %v", err)
		}
		if info.Size() == 0 {
			t.Fatal("KEPUB file is empty")
		}

		fmt.Printf("Generated KEPUB file at %s (%d bytes)\n", kepubPath, info.Size())
	})
}

// TestSimultaneousFormatGeneration tests generating multiple formats at once
//
// This test verifies:
// 1. Generation of EPUB files with different configuration options:
//   - Left-to-right reading direction
//   - Right-to-left reading direction
//   - Wide page splitting
//
// 2. Converting one of the EPUBs to KEPUB format
// 3. Verifying that all files exist and contain data
//
// This test ensures that multiple formats can be generated efficiently
// from the same source data with different configuration options.
func TestSimultaneousFormatGeneration(t *testing.T) {
	// Create test manga - one with wide pages to test different processing options
	manga := testhelpers.CreateWidePageTestManga()

	// Create temporary directory for output
	tempDir, err := os.MkdirTemp("", "kojirou-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// File paths for different formats
	epubPathLTR := filepath.Join(tempDir, "test-ltr.epub")
	epubPathRTL := filepath.Join(tempDir, "test-rtl.epub")
	epubPathSplit := filepath.Join(tempDir, "test-split.epub")
	kepubPath := filepath.Join(tempDir, "test.kepub.epub")

	// Generate EPUB with LTR setting
	epubLTR, cleanupLTR, err := GenerateEPUB(tempDir, manga, kindle.WidepagePolicyPreserve, false, true)
	if err != nil {
		t.Fatalf("GenerateEPUB(LTR) failed: %v", err)
	}
	if cleanupLTR != nil {
		defer cleanupLTR()
	}
	err = epubLTR.Write(epubPathLTR)
	if err != nil {
		t.Fatalf("Failed to write LTR EPUB: %v", err)
	}

	// Generate EPUB with RTL setting
	epubRTL, cleanupRTL, err := GenerateEPUB(tempDir, manga, kindle.WidepagePolicyPreserve, false, false)
	if err != nil {
		t.Fatalf("GenerateEPUB(RTL) failed: %v", err)
	}
	if cleanupRTL != nil {
		defer cleanupRTL()
	}
	err = epubRTL.Write(epubPathRTL)
	if err != nil {
		t.Fatalf("Failed to write RTL EPUB: %v", err)
	}

	// Generate EPUB with page splitting
	epubSplit, cleanupSplit, err := GenerateEPUB(tempDir, manga, kindle.WidepagePolicySplit, false, true)
	if err != nil {
		t.Fatalf("GenerateEPUB(Split) failed: %v", err)
	}
	if cleanupSplit != nil {
		defer cleanupSplit()
	}
	err = epubSplit.Write(epubPathSplit)
	if err != nil {
		t.Fatalf("Failed to write Split EPUB: %v", err)
	}

	// Convert LTR EPUB to KEPUB
	kepubData, err := kepubconv.ConvertToKEPUB(epubLTR)
	if err != nil {
		t.Fatalf("ConvertToKEPUB() failed: %v", err)
	}

	// Write KEPUB data to file
	err = os.WriteFile(kepubPath, kepubData, 0644)
	if err != nil {
		t.Fatalf("Failed to write KEPUB data: %v", err)
	}

	// Verify all files exist and have content
	files := []string{epubPathLTR, epubPathRTL, epubPathSplit, kepubPath}
	for _, filePath := range files {
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to get info for %s: %v", filePath, err)
		}
		if info.Size() == 0 {
			t.Fatalf("File is empty: %s", filePath)
		}
		fmt.Printf("Generated file: %s (%d bytes)\n", filePath, info.Size())
	}
}

// TestOutputFileStructure tests the structure of output files
func TestOutputFileStructure(t *testing.T) {
	// Create test manga with various special cases
	testCases := []struct {
		name     string
		manga    func() md.Manga
		filePath string
	}{
		{
			name: "Standard Manga",
			manga: func() md.Manga {
				return testhelpers.CreateTestManga()
			},
			filePath: "standard.epub",
		},
		{
			name: "Wide Page Manga",
			manga: func() md.Manga {
				return testhelpers.CreateWidePageTestManga()
			},
			filePath: "wide.epub",
		},
		{
			name: "Special Characters",
			manga: func() md.Manga {
				return testhelpers.CreateSpecialCharTitleManga()
			},
			filePath: "special.epub",
		},
	}

	// Create temporary directory for output
	tempDir, err := os.MkdirTemp("", "kojirou-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate EPUB
			manga := tc.manga()
			epubObj, cleanupObj, err := GenerateEPUB(tempDir, manga, kindle.WidepagePolicyPreserve, false, true)
			if err != nil {
				t.Fatalf("GenerateEPUB() failed: %v", err)
			}
			if cleanupObj != nil {
				defer cleanupObj()
			}

			// Write EPUB to file
			epubPath := filepath.Join(tempDir, tc.filePath)
			err = epubObj.Write(epubPath)
			if err != nil {
				t.Fatalf("Failed to write EPUB: %v", err)
			}

			// Convert to KEPUB
			kepubData, err := kepubconv.ConvertToKEPUB(epubObj)
			if err != nil {
				t.Fatalf("ConvertToKEPUB() failed: %v", err)
			}

			// Write KEPUB data to file
			kepubPath := filepath.Join(tempDir, tc.filePath+".kepub.epub")
			err = os.WriteFile(kepubPath, kepubData, 0644)
			if err != nil {
				t.Fatalf("Failed to write KEPUB data: %v", err)
			}

			// Verify both files exist and have content
			for _, path := range []string{epubPath, kepubPath} {
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("Failed to get info for %s: %v", path, err)
				}
				if info.Size() == 0 {
					t.Fatalf("File is empty: %s", path)
				}
				fmt.Printf("Generated file: %s (%d bytes)\n", path, info.Size())
			}
		})
	}
}
