// Copyright vinegar-development 2023

package main

import (
	"log"
	"os/user"
	"path/filepath"
)

var programDirs = defProgramDirs()

func defProgramDirs() []string {
	user, err := user.Current()
	Errc(err)

	var programDirPaths = []string{
		filepath.Join("users", user.Username, "AppData/Local"),
		"Program Files",
	}

	if Config.Env["WINEARCH"] == "win64" {
		programDirPaths = append(programDirPaths, "Program Files (x86)")
	}

	return programDirPaths
}

// Kill the wineprefix, required for autokill, and
// sometimes fixes Flatpak wine crashes.
func PfxKill() {
	log.Println("Killing wineprefix")
	_ = Exec("wineserver", false, "-k")
}
