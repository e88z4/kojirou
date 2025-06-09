package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKoboFolderModeOutput(t *testing.T) {
	tempDir := t.TempDir()
	series := "Test/Series: 01"
	volume := "01/02"
	filename := sanitizePOSIXName(series) + " v" + sanitizePOSIXName(volume) + ".kepub.epub"
	outputDir := filepath.Join(tempDir, "KoboBooks", sanitizePOSIXName(series))
	outputPath := filepath.Join(outputDir, filename)

	// Simulate writing a KEPUB file in Kobo folder mode
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	data := []byte("kepub test data")
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		t.Fatalf("failed to write kepub: %v", err)
	}

	// Check that the file exists and is POSIX compliant
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("expected file at %s, got error: %v", outputPath, err)
	}
	if strings.Contains(filename, "/") || strings.Contains(filename, "\x00") || strings.HasPrefix(filename, ".") || filename == "" {
		t.Errorf("filename is not POSIX compliant: %q", filename)
	}
}
