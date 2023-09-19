package state

import (
	"log"
	"os"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
)

func CleanPackages(pfx *wine.Prefix) error {
	pkgs, err := Packages(pfx)
	if err != nil {
		return err
	}

	log.Println("Checking for unused cached package files")

	return util.WalkDirExcluded(dirs.Downloads, pkgs, func(path string) error {
		log.Printf("Removing unused package %s", path)
		return os.Remove(filepath.Join(dirs.Downloads, path))
	})
}

func CleanVersions(pfx *wine.Prefix) error {
	var versionsDir = dirs.GetVersionsPath(pfx)

	vers, err := Versions(pfx)
	if err != nil {
		return err
	}

	log.Println("Checking for unused version directories")

	return util.WalkDirExcluded(versionsDir, vers, func(path string) error {
		log.Printf("Removing unused version directory %s", path)
		return os.RemoveAll(filepath.Join(versionsDir, path))
	})
}
