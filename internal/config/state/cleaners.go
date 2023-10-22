package state

import (
	"log"
	"os"

	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/util"
)

func CleanPackages() error {
	pkgs, err := Packages()
	if err != nil {
		return err
	}

	log.Println("Checking for unused cached package files")

	return util.WalkDirExcluded(dirs.Downloads, pkgs, func(path string) error {
		log.Printf("Removing unused package %s", path)
		return os.Remove(path)
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
		return os.RemoveAll(path)
	})
}
