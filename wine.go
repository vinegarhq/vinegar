package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

var AppDataDir = defAppDataDir()

func defAppDataDir() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf("could not get current user: %s", err)
	}

	return filepath.Join(Dirs.Pfx, "drive_c", "users", user.Username, "AppData")
}

func PfxInit() {
	// Initialize the wineprefix only when the 'windows' directory is happy.
	if _, err := os.Stat(filepath.Join(Dirs.Pfx, "drive_c", "windows")); err == nil {
		return
	}

	log.Println("Initializing wineprefix")

	if err := exec.Command("wineboot", "-i").Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting wineprefix version to", Config.Version)

	if err := exec.Command("winecfg", "/v", Config.Version).Run(); err != nil {
		log.Fatal(err)
	}
}

func PfxKill() {
	log.Println("Killing wineprefix")

	_ = exec.Command("wineserver", "-k").Run()
}
