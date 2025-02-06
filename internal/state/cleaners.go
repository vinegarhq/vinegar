package state

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/vinegarhq/vinegar/internal/dirs"
)

func firstExcluded(dir string, included []string, onExcluded func(string) error) error {
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

// CleanPackages removes all cached package downloads in dirs.Downloads
// that aren't held in the state's Binary packages.
func (s *State) CleanPackages() error {
	return firstExcluded(dirs.Downloads, s.Studio.Packages, func(path string) error {
		slog.Info("Cleaning up unused cached package", "path", path)
		return os.Remove(path)
	})
}

// CleanPackages removes all Binary versions that aren't
// held in the state's Binary packages.
func (s *State) CleanVersions() error {
	return firstExcluded(dirs.Versions, []string{s.Studio.Version}, func(path string) error {
		slog.Info("Cleaning up unused version directory", "path", path)
		return os.RemoveAll(path)
	})
}
