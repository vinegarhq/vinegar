package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Directories struct {
	Cache  string
	Config string
	Cwd    string
	Data   string
	Pfx    string
	Log    string
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

	dirs.Pfx = filepath.Join(dirs.Data, "pfx")
	dirs.Cwd = filepath.Join(dirs.Cache, "cwd")
	dirs.Log = filepath.Join(dirs.Cache, "logs")

	CheckDirs(DirMode, dirs.Cache, dirs.Config, dirs.Data, dirs.Pfx, dirs.Cwd, dirs.Log)

	return dirs
}

func CheckDirs(perm fs.FileMode, dirs ...string) {
	for _, dir := range dirs {
		info, err := os.Stat(dir)

		// Create the directory only when it does not exist.
		// Since logging is preferred this is also preferred.
		if errors.Is(err, os.ErrNotExist) {
			log.Println("Creating", perm, "dir:", dir)

			if err := os.MkdirAll(dir, perm); err != nil {
				log.Fatal(err)
			}

			continue
		}

		// The given permissions will always return if it is a file
		// ---------- or a directory d---------, we simply get the string
		// and remove the first character, which says if it is a file
		// or a directory, since all we care about is the read & write permissions.
		if err == nil && info.Mode().String()[1:] != perm.String()[1:] {
			log.Println("Setting dir", dir, "permissions to", perm)

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
