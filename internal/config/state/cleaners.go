package state

import (
	"os"
	"log"
	"path/filepath"

	"github.com/vinegarhq/aubun/internal/dirs"
	"github.com/vinegarhq/aubun/util"
)

func CleanPackages() error {
	pkgs, err := Packages()
	if err != nil {
		return err
	}

	log.Println("Checking for unused cached package files")

	return util.WalkDirExcluded(dirs.Downloads, pkgs, func(path string) error {
		log.Printf("Removing unused package %s", path)
		return os.Remove(filepath.Join(dirs.Downloads, path))
	})
}

func CleanVersions() error {
	vers, err := Versions()
	if err != nil {
		return err
	}

	log.Println("Checking for unused version directories")

	return util.WalkDirExcluded(dirs.Versions, vers, func(path string) error {
		log.Printf("Removing unused version directory %s", path)
		return os.RemoveAll(filepath.Join(dirs.Versions, path))
	})
}


