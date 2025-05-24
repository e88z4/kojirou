// Package formats provides format-specific functionality for different ebook formats
package formats

import (
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

// FormatGenerator defines an interface for generating different ebook formats
type FormatGenerator interface {
	Generate(manga md.Manga, widepage kindle.WidepagePolicy, autocrop bool, leftToRight bool) (FormatOutput, error)
}
