// User should not have to interact with this program at any time.

package main

import (
	"path/filepath"
	"github.com/adrg/xdg"
	"github.com/nocrusts/vinegar/src/downloader"
	"log"
	"os"
	"sync"
	"strings"
)

// Constants
const ROBLOXPLAYERLAUNCHERURL = "https://www.roblox.com/download/client"
const RBXFPSUNLOCKERURL = "https://github.com/axstin/rbxfpsunlocker/releases/download/v4.4.4/rbxfpsunlocker-x64.zip"

func main() {
	//Set up synchronization wait group
	wg := new(sync.WaitGroup)

	// Discover folders
	vinegarDataHome := (filepath.Join(xdg.DataHome, "/vinegar"))
	log.Println("Searching for XDG_DATA_HOME")
	os.MkdirAll(filepath.Join(xdg.DataHome, "/vinegar/executables"), 0755)
	log.Println("Searching for executables...")

	// Check if Launcher exists
	if _, err := os.Stat(filepath.Join(vinegarDataHome, "/executables/RobloxPlayerLauncher.exe")); err == nil {
		log.Println("Found Roblox Player Launcher!")
	} else {
		log.Println("Couldn't find launcher, installing...")
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Attempting to symlink from flatpak resources.")
			if err := os.Link("/app/RobloxPlayerLauncher.exe", filepath.Join(vinegarDataHome, "/executables/RobloxPlayerLauncher.exe")); err != nil {
				log.Println("Not in Flatpak!")
				downloader.Download(ROBLOXPLAYERLAUNCHERURL, filepath.Join(vinegarDataHome, "/executables/RobloxPlayerLauncher.exe"))
			}
		} ()
	}

	// Check if Unlocker exists
	if _, err := os.Stat(filepath.Join(vinegarDataHome, "/executables/rbxfpsunlocker.exe")); err == nil {
		log.Println("Found FPS unlocker!")
	} else {
		log.Println("Couldn't find unlocker, installing from web.")
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.println("Attempting to symlink from flatpak resources.")
			//not a typo, zip extraction can be done by flatpak builder.
			if err := os.Link("/app/rbxfpsunlocker.exe", filepath.Join(vinegarDataHome, "/executables/rbxfpsunlocker.exe")); err != nil {
				log.Println("Not in Flatpak!")
				downloader.Download(RBXFPSUNLOCKERURL, filepath.Join(vinegarDataHome, "/executables/rbxfpsunlocker.zip"))

			}
		} ()
	}


	wg.Wait()
}
