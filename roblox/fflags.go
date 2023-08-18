package roblox

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
)

type FFlags map[string]interface{}

func (f *FFlags) Apply(versionDir string) error {
	dir := filepath.Join(versionDir, "ClientSettings")
	path := filepath.Join(dir, "ClientAppSettings.json")

	// If the fflags are empty, the FFlags file's contents will be 'null'
	if len(*f) == 0 {
		return nil
	}

	log.Printf("Applying custom FFlags")

	err := os.Mkdir(dir, 0o755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer file.Close()

	fflags, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(fflags)
	if err != nil {
		return err
	}

	return nil
}
