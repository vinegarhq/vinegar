package main

import (
	"log"
	"os"
	"path/filepath"
)

type Directories struct {
	Cache     string
	Config    string
	Data      string
	Downloads string
	Logs      string
	Prefix    string
	Versions  string
}

var Dirs = GetDirectories()

// Function to declare the Directories struct with the default
// values. We prefer the XDG Variables over the default values, since in
// sandboxed environments such as Flatpak, it will set those variables with
// the appropriate sandboxed values.
func GetDirectories() Directories {
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

	dirs.Downloads = filepath.Join(dirs.Cache, "downloads")
	dirs.Logs = filepath.Join(dirs.Cache, "logs")
	dirs.Prefix = filepath.Join(dirs.Data, "prefix")
	dirs.Versions = filepath.Join(dirs.Data, "versions")

	return dirs
}

func DeleteDirs(dir ...string) {
	for _, d := range dir {
		if err := os.RemoveAll(d); err != nil {
			log.Fatal(err)
		}
	}
}
