package dirs

import (
	"os"
	"path/filepath"
	"slices"
)

// WalkForExcluded will walk the file tree located at dir, calling
// onExcluded for every file or directory that does not have a name in the given included.
func WalkForExcluded(dir string, included []string, onExcluded func(string) error) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if slices.Contains(included, file.Name()) {
			continue
		}

		if err := onExcluded(filepath.Join(dir, file.Name())); err != nil {
			return err
		}
	}

	return nil
}
