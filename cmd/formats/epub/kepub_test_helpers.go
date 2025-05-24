// Package epub provides EPUB and KEPUB conversion functionality
package epub

import (
	"image"
	"image/color"

	md "github.com/leotaku/kojirou/mangadex"
)

// createTestManga creates a test manga for use in testing
func createTestManga() md.Manga {
	manga := md.Manga{
		Info: md.MangaInfo{
			Title:   "Test Manga",
			ID:      "test-manga-id",
			Authors: []string{"Test Author"},
		},
		Volumes: map[md.Identifier]md.Volume{},
	}

	// Create a volume
	volID := md.NewIdentifier("1")
	vol := md.Volume{
		Info: md.VolumeInfo{
			Identifier: volID,
		},
		Chapters: map[md.Identifier]md.Chapter{},
	}

	// Create a chapter
	chapID := md.NewIdentifier("1-1")
	chap := md.Chapter{
		Info: md.ChapterInfo{
			Identifier:       chapID,
			Title:            "Chapter 1",
			VolumeIdentifier: volID,
		},
		Pages: map[int]image.Image{
			0: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
			1: image.NewRGBA(image.Rect(0, 0, 1000, 1400)),
		},
	}

	vol.Chapters[chapID] = chap
	manga.Volumes[volID] = vol

	return manga
}

// createKepubTestImage creates a test image with the specified dimensions and color
func createKepubTestImage(width, height int, bgColor color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	return img
}

// createKepubWidePageTestManga creates a manga with wide pages for testing
func createKepubWidePageTestManga() md.Manga {
	manga := createTestManga()

	// Find the first volume and chapter
	for volID, vol := range manga.Volumes {
		for chapID, chap := range vol.Chapters {
			// Add a wide page
			widePage := image.NewRGBA(image.Rect(0, 0, 2000, 1400))
			chap.Pages[2] = widePage
			vol.Chapters[chapID] = chap
		}
		manga.Volumes[volID] = vol
		break
	}

	return manga
}
