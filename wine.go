package main

import (
	"log"
	"os"
	"path/filepath"
)

func PfxInit() {
	// Initialize the wineprefix only when the 'windows' directory is happy.
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
