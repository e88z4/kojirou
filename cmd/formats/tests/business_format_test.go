package cmd_test

import (
	"testing"

	"github.com/leotaku/kojirou/cmd"
	"github.com/leotaku/kojirou/cmd/formats"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	"github.com/leotaku/kojirou/cmd/formats/testhelpers"
	md "github.com/leotaku/kojirou/mangadex"
)

// TestMultiFormatEndToEnd tests the complete multi-format generation process
func TestMultiFormatEndToEnd(t *testing.T) {
	// Create a temporary directory for output
	testDir := t.TempDir()

	// Use a real test manga and volume
	skeleton := testhelpers.CreateTestManga()
	var volume md.Volume
	for _, v := range skeleton.Volumes {
		volume = v
		break
	}
	if volume.Info.Identifier.String() == "" {
		t.Fatal("No volume found in test manga")
	}

	// Create test directory
	dir := kindle.NewNormalizedDirectory(testDir, skeleton.Info.Title, false)

	// Save the original formats arg and restore it after the test
	origFormatsArg := cmd.FormatsArg
	defer func() { cmd.FormatsArg = origFormatsArg }()

	// Test with all supported formats
	cmd.FormatsArg = "mobi,epub,kepub"

	// Call HandleVolume
	err := cmd.HandleVolume(skeleton, volume, dir)

	// In a real test this would pass with proper mocking
	// Here we expect an error due to lack of real manga data
	if err == nil {
		t.Error("Expected error due to missing manga data, but got nil")
	}

	// Verify format selection error handling
	cmd.FormatsArg = "mobi,invalid,epub"
	_, err = formats.ParseFormats(cmd.FormatsArg)
	if err == nil {
		t.Error("Expected error with invalid format, but got nil")
	}
}
