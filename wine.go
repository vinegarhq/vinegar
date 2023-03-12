package main

import (
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var programDirs = defProgramDirs()

func defProgramDirs() []string {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dirs := []string{
		filepath.Join("users", user.Username, "AppData/Local"),
		"Program Files",
		"Program Files (x86)",
	}

	for i, dir := range dirs {
		dirs[i] = filepath.Join(Dirs.Pfx, "drive_c", dir)
	}

	return dirs
}

func EdgeDirSet(perm fs.FileMode, create bool) {
	for _, programDir := range programDirs {
		EdgeDir := filepath.Join(programDir, "Microsoft", "EdgeUpdate")

		if create {
			CheckDirs(DirMode, filepath.Dir(EdgeDir))
			CheckDirs(perm, EdgeDir)

			continue
		}

		if _, err := os.Stat(EdgeDir); err == nil {
			if err := os.Chmod(EdgeDir, perm); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func PfxInit() {
	if _, err := os.Stat(filepath.Join(Dirs.Pfx, "drive_c", "windows")); err == nil {
		return
	}

	log.Println("Initializing wineprefix")

	if err := Exec("wineboot", false, "-i"); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting wineprefix version to", Config.Version)

	if err := Exec("winecfg", false, "/v", Config.Version); err != nil {
		log.Fatal(err)
	}
}

func PfxKill() {
	log.Println("Killing wineprefix")

	_ = Exec("wineserver", false, "-k")
}
