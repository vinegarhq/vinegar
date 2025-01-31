package state

import (
	"log/slog"
	"os"

	"github.com/vinegarhq/vinegar/internal/dirs"
)

// CleanPackages removes all cached package downloads in dirs.Downloads
// that aren't held in the state's Binary packages.
func (s *State) CleanPackages() error {
	return dirs.WalkForExcluded(dirs.Downloads, s.Studio.Packages, func(path string) error {
		slog.Info("Cleaning up unused cached package", "path", path)
		return os.Remove(path)
	})
}

// CleanPackages removes all Binary versions that aren't
// held in the state's Binary packages.
func (s *State) CleanVersions() error {
	return dirs.WalkForExcluded(dirs.Versions, []string{s.Studio.Version}, func(path string) error {
		slog.Info("Cleaning up unused version directory", "path", path)
		return os.RemoveAll(path)
	})
}
