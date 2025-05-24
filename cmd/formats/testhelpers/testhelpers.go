// Package testhelpers provides common testing utilities for the Kojirou project
//
// This package contains shared helper functions for creating test data, primarily
// for testing the various format generators. It helps reduce code duplication
// across test files and provides consistent test data.
//
// Key functions include:
//   - CreateTestManga: Creates a standard test manga with volumes and chapters
//   - CreateTestImage: Generates test images with specified dimensions
//   - CreateWidePageTestManga: Creates a manga with wide pages for testing splitting functionality
//   - CreateInvalidImageManga: Creates a manga with nil images to test error handling
//
// Example usage:
//
//	// In a test file
//	func TestSomeFunction(t *testing.T) {
//	    manga := testhelpers.CreateTestManga()
//	    // Use the test manga in your test
//	}
package testhelpers

import (
	"image"
	"image/color"

	md "github.com/leotaku/kojirou/mangadex"
)

// CreateTestManga creates a basic test manga with standard structure
//
// Returns a manga object with:
//   - Two volumes with proper identifiers
//   - Each volume containing 1-2 chapters
//   - Standard metadata (title, author, etc.)
//   - Sample pages in each chapter
//
// This function is useful for general-purpose testing of manga processing code.
func CreateTestManga() md.Manga {
	manga := md.Manga{
		Info: md.MangaInfo{
			Title:   "Test Manga",
			ID:      "test-manga-id",
			Authors: []string{"Test Author"},
		},
		Volumes: map[md.Identifier]md.Volume{},
	}

	// Create two volumes
	vol1ID := md.NewIdentifier("1")
	vol2ID := md.NewIdentifier("2")

	vol1 := md.Volume{
		Info: md.VolumeInfo{
			Identifier: vol1ID,
		},
		Chapters: map[md.Identifier]md.Chapter{},
	}

	vol2 := md.Volume{
		Info: md.VolumeInfo{
			Identifier: vol2ID,
		},
		Chapters: map[md.Identifier]md.Chapter{},
	}

	// Add chapters to volumes
	chap1ID := md.NewIdentifier("1-1")
	chap1 := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chap1ID,
			Title:            "Chapter 1",
			VolumeIdentifier: vol1ID,
		},
		Pages: map[int]image.Image{
			0: CreateTestImage(1000, 1500, color.White),
			1: CreateTestImage(1000, 1500, color.White),
		},
	}
	vol1.Chapters[chap1ID] = chap1

	chap2ID := md.NewIdentifier("2-1")
	chap2 := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chap2ID,
			Title:            "Chapter 2",
			VolumeIdentifier: vol2ID,
		},
		Pages: map[int]image.Image{
			0: CreateTestImage(1000, 1500, color.White),
			1: CreateTestImage(1000, 1500, color.White),
		},
	}
	vol2.Chapters[chap2ID] = chap2

	// Add volumes to manga
	manga.Volumes[vol1ID] = vol1
	manga.Volumes[vol2ID] = vol2

	return manga
}

// CreateTestImage creates a test image with specified dimensions and background color
//
// Parameters:
//   - width: The width of the image in pixels
//   - height: The height of the image in pixels
//   - bgColor: The background color of the image
//
// Returns an image.Image with the specified dimensions and background color,
// plus a simple pattern to make it distinguishable.
//
// This function is useful for creating test images of various sizes and aspects
// to test image processing functions.
func CreateTestImage(width, height int, bgColor color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with background color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw a simple pattern to make it more recognizable
	for y := 0; y < height; y += 10 {
		for x := 0; x < width; x += 10 {
			if (x/10+y/10)%2 == 0 {
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}

	return img
}

// CreateWidePageTestManga creates a manga with wide pages for testing
//
// Returns a manga object with:
//   - Standard volumes and chapters structure
//   - Several wide pages with varying aspect ratios (e.g., 2:1, 3:1)
//   - Mixed with standard portrait-orientation pages
//
// This function is particularly useful for testing wide page handling strategies
// such as splitting, preserving, or scaling pages with different aspect ratios.
func CreateWidePageTestManga() md.Manga {
	manga := CreateTestManga()

	// Add a wide page chapter
	volID := md.NewIdentifier("1")
	chapID := md.NewIdentifier("wide")

	// Create a chapter with wide pages
	wideChapter := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chapID,
			Title:            "Wide Page Chapter",
			VolumeIdentifier: volID,
		},
		Pages: map[int]image.Image{
			0: CreateTestImage(1000, 1500, color.White), // Normal
			1: CreateTestImage(2000, 1000, color.White), // Wide 2:1
			2: CreateTestImage(3000, 1000, color.White), // Very wide 3:1
			3: CreateTestImage(1000, 1500, color.White), // Normal
		},
	}

	// Add to volume if it exists, otherwise create a new volume
	vol, exists := manga.Volumes[volID]
	if exists {
		vol.Chapters[chapID] = wideChapter
		manga.Volumes[volID] = vol
	} else {
		newVol := md.Volume{
			Info: md.VolumeInfo{
				Identifier: volID,
			},
			Chapters: map[md.Identifier]md.Chapter{
				chapID: wideChapter,
			},
		}
		manga.Volumes[volID] = newVol
	}

	return manga
}

// CreateInvalidImageManga creates a manga with nil images
func CreateInvalidImageManga() md.Manga {
	manga := CreateTestManga()

	// Find first chapter and add a nil image
	for volID, vol := range manga.Volumes {
		for chapID, chap := range vol.Chapters {
			// Add a nil image to the chapter
			chap.Pages[99] = nil
			vol.Chapters[chapID] = chap
			manga.Volumes[volID] = vol
			return manga // Return after modifying the first chapter found
		}
	}

	return manga
}

// CreateSpecialCharTitleManga creates a manga with special characters in the title
func CreateSpecialCharTitleManga() md.Manga {
	manga := CreateTestManga()
	manga.Info.Title = "Special & < > Characters: Test"
	return manga
}

// CreateLargeImageManga creates a manga with very large images
func CreateLargeImageManga() md.Manga {
	manga := CreateTestManga()

	// Find first chapter and add large images
	for volID, vol := range manga.Volumes {
		for chapID, chap := range vol.Chapters {
			// Add large images to the chapter
			chap.Pages[0] = CreateTestImage(3000, 4000, color.White)
			chap.Pages[1] = CreateTestImage(4000, 5000, color.White)
			vol.Chapters[chapID] = chap
			manga.Volumes[volID] = vol
			return manga // Return after modifying the first chapter found
		}
	}

	return manga
}
