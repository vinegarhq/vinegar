package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var LocalProgramDir = defLocalProgramDir()

func defLocalProgramDir() string {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dir := filepath.Join(Dirs.Pfx, "drive_c", "users", user.Username, "AppData/Local", "vinegar")

	CreateDirs(dir)

	return dir
}

func PfxInit() {
	if _, err := os.Stat(filepath.Join(Dirs.Pfx, "drive_c", "windows")); err == nil {
		return
	}

	log.Println("Initializing wineprefix")

	if err := Exec("wineboot", "", "-i"); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting wineprefix version to", Config.Version)

	if err := Exec("winecfg", "", "/v", Config.Version); err != nil {
		log.Fatal(err)
	}
}

func PfxKill() {
	log.Println("Killing wineprefix")

	_ = Exec("wineserver", "", "-k")
}
