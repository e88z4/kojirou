// Modified helper functions that use the shared testhelpers package
// These functions replace the existing ones in epub_format_test.go

package epub

// How to migrate to shared test helpers:
//
// 1. Add import for the testhelpers package:
//    "github.com/leotaku/kojirou/cmd/formats/testhelpers"
//
// 2. Replace existing helper functions with calls to shared implementations:
//
// func createTestManga() md.Manga {
//     return testhelpers.CreateTestManga()
// }
//
// func createTestImage(width, height int, bgColor color.Color) image.Image {
//     return testhelpers.CreateTestImage(width, height, bgColor)
// }
//
// func createWidePageTestManga() md.Manga {
//     return testhelpers.CreateWidePageTestManga()
// }
//
// func createInvalidImageManga() md.Manga {
//     return testhelpers.CreateInvalidImageManga()
// }
//
// func createSpecialCharTitleManga() md.Manga {
//     return testhelpers.CreateSpecialCharTitleManga()
// }
//
// func createLargeImageManga() md.Manga {
//     return testhelpers.CreateLargeImageManga()
// }
//
// 3. Keep unique helper functions that aren't in the shared package:
//
// func createEmptyVolumeManga() md.Manga { ... }
// func createNoTitleManga() md.Manga { ... }
// func createMixedLanguageManga() md.Manga { ... }
