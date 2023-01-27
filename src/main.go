// User should not have to interact with this program at any time.

package main

import (
	"path/filepath"
	"archive/zip"
	"github.com/adrg/xdg"
	"github.com/nocrusts/vinegar/src/downloader"
	"log"
	"os"
	"sync"
)

// Constants
const ROBLOXPLAYERLAUNCHERURL = "https://www.roblox.com/download/client"
const RBXFPSUNLOCKERURL = "https://github.com/axstin/rbxfpsunlocker/releases/download/v4.4.4/rbxfpsunlocker-x64.zip"

//func build_prefix(mode string) (err error) {
//
//}

func main() {
	//Set up synchronization wait group
	wg := new(sync.WaitGroup)

	// Discover if EXEs are available.
	vinegarDataHome := (filepath.Join(xdg.DataHome, "/vinegar"))
	log.Println("Searching for XDG_DATA_HOME")
	os.MkdirAll(filepath.Join(xdg.DataHome, "/vinegar/executables"), 0755)
	log.Println("Searching for executables...")

	// Check if Launcher exists
	if _, err := os.Stat(filepath.Join(vinegarDataHome, "/executables/RobloxPlayerLauncher.exe")); err == nil {
		log.Println("Found Roblox Player Launcher!")
	} else {
		log.Println("Couldn't find launcher, installing from web.")
		wg.Add(1)
		go downloader.Download(ROBLOXPLAYERLAUNCHERURL, filepath.Join(vinegarDataHome, "/executables/RobloxPlayerLauncher.exe"), wg)
	}

	// Check if Unlocker exists
	if _, err := os.Stat(filepath.Join(vinegarDataHome, "/executables/rbxfpsunlocker.exe")); err == nil {
		log.Println("Found FPS unlocker!")
	} else {
		log.Println("Couldn't find unlocker, installing from web.")
		wg.Add(1)
		go func() {
			downloader.Download(RBXFPSUNLOCKERURL, filepath.Join(vinegarDataHome, "/executables/rbxfpsunlocker.zip"), wg)
			//TODO: Extract zip
			archive, err := zip.OpenReader(filepath.Join(vinegarDataHome, "/executables/rbxfpsunlocker.zip")

		} ()

	}


	wg.Wait()
}
