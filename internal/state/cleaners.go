package state

import (
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/vinegarhq/vinegar/internal/dirs"
)

// CleanPackages removes all cached package downloads in dirs.Downloads
// that aren't held in the state's Binary packages.
func (s *State) CleanPackages() error {
	log.Println("Checking for unused cached package files")

	return walkDirExcluded(dirs.Downloads, s.Packages(), func(path string) error {
		log.Printf("Removing unused package %s", path)
		return os.Remove(path)
	})
}

// CleanPackages removes all Binary versions that aren't
// held in the state's Binary packages.
func (s *State) CleanVersions() error {
	log.Println("Checking for unused version directories")

	return walkDirExcluded(dirs.Versions, s.Versions(), func(path string) error {
		log.Printf("Removing unused version directory %s", path)
		return os.RemoveAll(path)
	})
}

// walkDirExcluded will walk the file tree located at dir, calling
// onExcluded for every file or directory that does not have a name in included.
func walkDirExcluded(dir string, included []string, onExcluded func(string) error) error {
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
