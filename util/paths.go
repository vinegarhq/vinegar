package util

import (
	"os"
	"path/filepath"
)

func UserDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share"), nil
}

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
