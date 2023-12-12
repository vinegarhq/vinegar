package state

import (
	"log"
	"os"

	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/util"
)

// CleanPackages removes all cached package downloads in dirs.Downloads
// that aren't held in the state's Binary packages.
func (s *State) CleanPackages() error {
	log.Println("Checking for unused cached package files")

	return util.WalkDirExcluded(dirs.Downloads, s.Packages(), func(path string) error {
		log.Printf("Removing unused package %s", path)
		return os.Remove(path)
	})
}

// CleanPackages removes all Binary versions that aren't
// held in the state's Binary packages.
func (s *State) CleanVersions() error {
	log.Println("Checking for unused version directories")

	return util.WalkDirExcluded(dirs.Versions, s.Versions(), func(path string) error {
		log.Printf("Removing unused version directory %s", path)
		return os.RemoveAll(path)
	})
}
