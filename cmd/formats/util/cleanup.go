package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ForceRemoveAll recursively removes a directory tree, forcibly changing permissions if needed.
// It returns a combined error if any removals fail.
func ForceRemoveAll(path string) error {
	var errs []string
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		errs = append(errs, fmt.Sprintf("lstat %s: %v", path, err))
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	if info.IsDir() {
		d, err := os.Open(path)
		if err != nil {
			_ = os.Chmod(path, 0777)
			if remErr := os.Remove(path); remErr != nil {
				if os.IsPermission(remErr) || strings.Contains(remErr.Error(), "directory not empty") {
					_ = os.Chmod(path, 0777)
					if retryErr := os.Remove(path); retryErr != nil {
						errs = append(errs, fmt.Sprintf("remove dir %s after retry: %v", path, retryErr))
					}
				} else {
					errs = append(errs, fmt.Sprintf("remove dir %s: %v", path, remErr))
				}
			}
			return fmt.Errorf("%s", strings.Join(errs, "; "))
		}
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err != nil {
			_ = os.Chmod(path, 0777)
			if remErr := os.Remove(path); remErr != nil {
				if os.IsPermission(remErr) || strings.Contains(remErr.Error(), "directory not empty") {
					_ = os.Chmod(path, 0777)
					if retryErr := os.Remove(path); retryErr != nil {
						errs = append(errs, fmt.Sprintf("remove dir %s after retry: %v", path, retryErr))
					}
				} else {
					errs = append(errs, fmt.Sprintf("remove dir %s: %v", path, remErr))
				}
			}
			return fmt.Errorf("%s", strings.Join(errs, "; "))
		}
		for _, name := range names {
			child := filepath.Join(path, name)
			if chErr := ForceRemoveAll(child); chErr != nil {
				errs = append(errs, chErr.Error())
			}
		}
		parent := filepath.Dir(path)
		if parent != path {
			_ = os.Chmod(parent, 0777)
		}
		_ = os.Chmod(path, 0777)
		if remErr := os.Remove(path); remErr != nil {
			if os.IsPermission(remErr) || strings.Contains(remErr.Error(), "directory not empty") {
				_ = os.Chmod(path, 0777)
				if retryErr := os.Remove(path); retryErr != nil {
					errs = append(errs, fmt.Sprintf("remove dir %s after retry: %v", path, retryErr))
				}
			} else {
				errs = append(errs, fmt.Sprintf("remove dir %s: %v", path, remErr))
			}
		}
	} else {
		_ = os.Chmod(path, 0666)
		if remErr := os.Remove(path); remErr != nil {
			errs = append(errs, fmt.Sprintf("remove file %s: %v", path, remErr))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}
