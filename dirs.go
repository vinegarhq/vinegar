package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Directories struct {
	Cache     string
	Config    string
	Cwd       string
	Data      string
	Downloads string
	Log       string
	Pfx       string
}

var (
	Dirs                  = defDirs()
	DirMode   fs.FileMode = 0755
	DirROMode fs.FileMode = 0664
	FileMode  fs.FileMode = 0644
)

// Function to declare the Directories struct with the default
// values. We prefer the XDG Variables over the default values, since in
// sandboxed environments such as Flatpak, it will set those variables with
// the appropriate sandboxed values.
func defDirs() Directories {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to get home directory")
	}

	xdgDirs := map[string]string{
		"XDG_CACHE_HOME":  filepath.Join(homeDir, ".cache"),
		"XDG_CONFIG_HOME": filepath.Join(homeDir, ".config"),
		"XDG_DATA_HOME":   filepath.Join(homeDir, ".local", "share"),
	}

	for varName := range xdgDirs {
		if value, ok := os.LookupEnv(varName); ok {
			xdgDirs[varName] = value
		}
	}

	dirs := Directories{
		Cache:  filepath.Join(xdgDirs["XDG_CACHE_HOME"], "vinegar"),
		Config: filepath.Join(xdgDirs["XDG_CONFIG_HOME"], "vinegar"),
		Data:   filepath.Join(xdgDirs["XDG_DATA_HOME"], "vinegar"),
	}

	dirs.Cwd = filepath.Join(dirs.Cache, "cwd")
	dirs.Downloads = filepath.Join(dirs.Cache, "downloads")
	dirs.Log = filepath.Join(dirs.Cache, "logs")
	dirs.Pfx = filepath.Join(dirs.Data, "pfx")

	CreateDirs(dirs.Cwd, dirs.Downloads, dirs.Log, dirs.Pfx)

	return dirs
}

func CreateDirs(dirs ...string) {
	for _, dir := range dirs {
		// Don't do anything if the directory does exist, to just save some logging
		if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
			continue
		}

		log.Println("Creating directory:", dir)

		if err := os.MkdirAll(dir, DirMode); err != nil {
			log.Fatal(err)
		}
	}
}

func SetPermDirs(perm fs.FileMode, dirs ...string) {
	for _, dir := range dirs {
		info, err := os.Stat(dir)

		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			log.Fatal(err)
		}

		// The given permissions will always return if it is a file
		// ---------- or a directory d---------, we simply get the string
		// and remove the first character, which says if it is a file
		// or a directory, since all we care about is the read & write permissions.
		if info.Mode().String()[1:] != perm.String()[1:] {
			log.Println("Setting directory", info.Name(), "permissions to", perm)

			if err := os.Chmod(dir, perm); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func DeleteDirs(dir ...string) {
	for _, d := range dir {
		log.Println("Deleting dir:", d)

		if err := os.RemoveAll(d); err != nil {
			log.Fatal(err)
		}
	}
}
