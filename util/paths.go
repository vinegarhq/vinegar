package util

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func WalkDirExcluded(dir string, included []string, onExcluded func(string) error) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

find:
	for _, file := range files {
		for _, inc := range included {
			if file.Name() == inc {
				continue find
			}
		}

		if err := onExcluded(file.Name()); err != nil {
			return err
		}
	}

	return nil
}

func FindTimeFile(dir string, comparison *time.Time) (string, error) {
	var name string

	err := filepath.Walk(dir, func(p string, i fs.FileInfo, err error) error {
		if i.ModTime().After(*comparison) {
			name = p
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if name == "" {
		return "", os.ErrNotExist
	}

	return name, nil
}
