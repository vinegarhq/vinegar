package dirs

import (
	"os"
	"path/filepath"
	"slices"
)

// WalkForExcluded will walk the file tree located at dir, calling
// onExcluded for every file or directory that does not have a name in the given included.
func WalkForExcluded(dir string, included []string, onExcluded func(string) error) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if slices.Contains(included, info.Name()) || info.IsDir() {
			return nil
		}

		return onExcluded(path)
	})
}
