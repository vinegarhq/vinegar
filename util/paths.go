package util

import (
	"os"
	"path/filepath"
)

// Possibly slower but will restore compatability
func slice_contains(included []string, file_name string) bool {
	for _, name := range included {
		if name == file_name {
			return true
		}
	}
	return false
}

// WalkDirExcluded will walk the file tree located at dir, calling
// onExcluded for every file or directory that does not have a name in included.
func WalkDirExcluded(dir string, included []string, onExcluded func(string) error) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if slice_contains(included, file.Name()) {
			continue
		}

		if err := onExcluded(filepath.Join(dir, file.Name())); err != nil {
			return err
		}
	}

	return nil
}
