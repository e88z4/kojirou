package util

import (
	"strings"
)

// SanitizePOSIXName replaces or removes characters not allowed in POSIX file and folder names
func SanitizePOSIXName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\x00", "_")
	name = replacer.Replace(name)
	name = strings.Trim(name, " .")
	if name == "" || name == "." || name == ".." {
		name = "untitled"
	}
	return name
}
