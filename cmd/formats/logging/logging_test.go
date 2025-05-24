package logging

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/leotaku/kojirou/cmd/formats"
)

func TestFormatLogging(t *testing.T) {
	// Save and restore stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	// Test with color disabled to make output predictable
	EnableColor(false)

	// Test FormatInfo
	FormatInfo(formats.FormatEpub, "Testing info logging")

	// Test FormatSuccess
	FormatSuccess(formats.FormatMobi, "Testing success logging")

	// Test FormatError
	FormatError(formats.FormatKepub, errors.New("test error"))

	// Debug logging is off by default
	FormatDebug(formats.FormatEpub, "This should not appear")

	// Enable debug logging and test
	EnableDebug(true)
	FormatDebug(formats.FormatEpub, "This should appear")

	// Close writer to be able to read from the pipe
	w.Close()

	// Read output from pipe
	out, _ := io.ReadAll(r)

	// Verify output contains expected content
	output := string(out)
	t.Log("Captured output:", output)

	// Check for expected content
	expectedStrings := []string{
		"[epub] Testing info logging",
		"[mobi] Testing success logging",
		"[kepub] Error: test error",
		"[epub] DEBUG: This should appear",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't", expected)
		}
	}

	// This should NOT be in the output since debug was initially disabled
	unexpected := "[epub] DEBUG: This should not appear"
	if contains(output, unexpected) {
		t.Errorf("Output unexpectedly contained %q", unexpected)
	}
}

func TestTimedOperation(t *testing.T) {
	// Save and restore stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	// Enable debug mode and disable color for testing
	EnableDebug(true)
	EnableColor(false)

	// Test successful operation
	err := TimedOperation(formats.FormatEpub, "test operation", func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected nil error from successful operation, got %v", err)
	}

	// Test failed operation
	testErr := errors.New("operation failed")
	err = TimedOperation(formats.FormatMobi, "failing operation", func() error {
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}

	// Close writer to be able to read from the pipe
	w.Close()

	// Read output from pipe
	out, _ := io.ReadAll(r)

	// Verify output contains expected content
	output := string(out)
	t.Log("Captured output:", output)

	// Check for expected content
	expectedStrings := []string{
		"[epub] DEBUG: Starting test operation",
		"[epub] DEBUG: Completed test operation in",
		"[mobi] DEBUG: Starting failing operation",
		"[mobi] Error: failing operation: operation failed",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't", expected)
		}
	}
}

func TestFormatSummary(t *testing.T) {
	// Save and restore stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	// Disable color for testing
	EnableColor(false)

	// Create a test status map
	formatStatuses := map[formats.FormatType]string{
		formats.FormatEpub:  "Success",
		formats.FormatMobi:  "Error: something went wrong",
		formats.FormatKepub: "Skipped (already exists)",
	}

	// Call FormatSummary
	FormatSummary(formatStatuses)

	// Close writer to be able to read from the pipe
	w.Close()

	// Read output from pipe
	out, _ := io.ReadAll(r)

	// Verify output contains expected content
	output := string(out)
	t.Log("Captured output:", output)

	// Check for expected content
	expectedStrings := []string{
		"✓ Success: epub",
		"↷ Skipped: kepub",
		"✗ Errors: mobi",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't", expected)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && s != substr && strings.Contains(s, substr)
}
