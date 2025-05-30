package kepubconv

import (
	"os"
	"path/filepath"
)

// ForceRemoveAll recursively removes a directory tree, forcibly changing permissions if needed.
func ForceRemoveAll(path string) error {
	// Remove files first
	if err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		mode := info.Mode()
		if mode.IsDir() {
			return nil
		}
		if mode&0200 == 0 {
			_ = os.Chmod(p, 0666)
		}
		_ = os.Remove(p)
		return nil
	}); err != nil {
		return err
	}
	// Remove directories (bottom-up)
	if err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			_ = os.Chmod(p, 0777)
			if p != path {
				_ = os.Remove(p)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	_ = os.Chmod(path, 0777)
	return os.Remove(path)
}
