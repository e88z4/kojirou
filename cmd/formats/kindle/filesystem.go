package kindle

import (
	"errors"
	"fmt"
	"image/jpeg"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/leotaku/kojirou/cmd/formats/output"
	"github.com/leotaku/kojirou/cmd/formats/progress"
	"github.com/leotaku/kojirou/cmd/formats/util"
	md "github.com/leotaku/kojirou/mangadex"
)

type NormalizedDirectory struct {
	bookDirectory      string
	thumbnailDirectory string
}

func NewNormalizedDirectory(target, title string, kindleFolder bool) NormalizedDirectory {
	title = util.SanitizePOSIXName(title)
	title = strings.ReplaceAll(title, ":", "_")
	title = strings.ReplaceAll(title, " ", "_") // Remove spaces for POSIX compliance
	title = strings.Trim(title, ".")            // Remove trailing/leading dots
	if title == "" || title == "." || title == ".." {
		title = "untitled"
	}
	switch {
	case kindleFolder && target == "":
		return NormalizedDirectory{
			bookDirectory:      path.Join("kindle", "documents", title),
			thumbnailDirectory: path.Join("kindle", "system", "thumbnails"),
		}
	case kindleFolder:
		return NormalizedDirectory{
			bookDirectory:      path.Join(target, "documents", title),
			thumbnailDirectory: path.Join(target, "system", "thumbnails"),
		}
	case target == "":
		return NormalizedDirectory{
			bookDirectory: title,
		}
	default:
		return NormalizedDirectory{
			bookDirectory: target,
		}
	}
}

func (n *NormalizedDirectory) Has(identifier md.Identifier) bool {
	// Check for any supported format
	exts := []string{".azw3", ".epub", ".kepub.epub"}
	base := identifier.StringFilled(4, 2, false)
	for _, ext := range exts {
		if exists(path.Join(n.bookDirectory, base+ext)) {
			return true
		}
	}
	return false
}

// HasWithExtension checks if a file with the specified identifier and extension exists
func (n *NormalizedDirectory) HasWithExtension(identifier md.Identifier, extension string) bool {
	filename := identifier.StringFilled(4, 2, false) + "." + extension
	return exists(path.Join(n.bookDirectory, filename))
}

// Path returns the normalized path for a volume with the given identifier and extension
func (n *NormalizedDirectory) Path(identifier md.Identifier, extension string) string {
	if n.bookDirectory == "" {
		return ""
	}
	filename := identifier.StringFilled(4, 2, false) + "." + extension
	return path.Join(n.bookDirectory, filename)
}

// WriteFormat writes the output to the appropriate file based on its extension
func (n *NormalizedDirectory) WriteFormat(identifier md.Identifier, out output.FormatOutput, p progress.Progress) error {
	if n.bookDirectory == "" {
		return fmt.Errorf("unsupported configuration: no book output")
	}

	// Get the path for this format
	filepath := n.Path(identifier, out.Extension())

	f, err := create(filepath)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer f.Close()

	data, err := out.GetBytes()
	if err != nil {
		return fmt.Errorf("get bytes: %w", err)
	}

	if _, err := p.NewProxyWriter(f).Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	// Handle thumbnail for MOBI/AZW3 files
	if mobi, ok := out.(*output.MobiOutput); ok && n.thumbnailDirectory != "" {
		coverImage := mobi.GetCoverImage()
		if coverImage != nil {
			f, err := create(path.Join(n.thumbnailDirectory, mobi.GetThumbFilename()))
			if err != nil {
				return fmt.Errorf("create thumbnail: %w", err)
			}
			defer f.Close()

			if err := jpeg.Encode(p.NewProxyWriter(f), coverImage, nil); err != nil {
				return fmt.Errorf("write thumbnail: %w", err)
			}
		}
	}

	return nil
}

// WriteEpub writes an EPUB format output to the appropriate file
func (n *NormalizedDirectory) WriteEpub(identifier md.Identifier, epub *output.EpubOutput, p progress.Progress) error {
	return n.WriteFormat(identifier, epub, p)
}

// WriteKepub writes a KEPUB format output to the appropriate file
func (n *NormalizedDirectory) WriteKepub(identifier md.Identifier, kepub *output.KepubOutput, p progress.Progress) error {
	return n.WriteFormat(identifier, kepub, p)
}

// WriteMobi writes a MOBI/AZW3 format output to the appropriate file
func (n *NormalizedDirectory) WriteMobi(identifier md.Identifier, mobi *output.MobiOutput, p progress.Progress) error {
	return n.WriteFormat(identifier, mobi, p)
}

// GetExistingFormats returns a map of format extensions to file paths for a given identifier
func (n *NormalizedDirectory) GetExistingFormats(identifier md.Identifier) map[string]string {
	result := make(map[string]string)
	exts := []string{"azw3", "epub", "kepub.epub"}
	base := identifier.StringFilled(4, 2, false)

	for _, ext := range exts {
		filepath := path.Join(n.bookDirectory, base+"."+ext)
		if exists(filepath) {
			result[ext] = filepath
		}
	}

	return result
}

func pathnameFromTitle(filename string) string {
	return util.SanitizePOSIXName(filename)
}

func exists(pathname string) bool {
	_, err := os.Stat(pathname)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if errors.Is(err, fs.ErrExist) {
		return true
	} else if err != nil {
		return false
	} else {
		return true
	}
}

func create(pathname string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(pathname), os.ModePerm); err != nil {
		return nil, fmt.Errorf("directory: %w", err)
	}
	if f, err := os.Create(pathname); err != nil {
		return nil, fmt.Errorf("file: %w", err)
	} else {
		return f, nil
	}
}
