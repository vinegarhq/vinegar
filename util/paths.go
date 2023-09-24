package util

import (
	"os"
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
