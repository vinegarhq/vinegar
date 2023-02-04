// Copyright vinegar-development 2023

package vinegar

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	FPSUNLOCKERHASH = "sha256:050fe7c0127dbd4fdc0cecf3ba46248ba7e14d37edba1a54eac40602c130f2f8"
	FPSUNLOCKERURL  = "https://github.com/axstin/rbxfpsunlocker/releases/download/v4.4.4/rbxfpsunlocker-x64.zip"
)

// Launch or automatically install axstin's rbxfpsunlocker.
// This function will also create it's own settings for rbxfpsunlocker, for
// faster or cleaner startup.
func RbxFpsUnlocker() {
	fpsUnlockerPath := filepath.Join(Dirs.Data, "rbxfpsunlocker.exe")
	_, err := os.Stat(fpsUnlockerPath)

	if os.IsNotExist(err) {
		fpsUnlockerZip := filepath.Join(Dirs.Cache, "rbxfpsunlocker.zip")
		log.Println("Installing rbxfpsunlocker")
		Download(fpsUnlockerZip, FPSUNLOCKERURL)
		Unzip(fpsUnlockerZip, fpsUnlockerPath)
	}

	var settings = []string{
		"UnlockClient=true",
		"UnlockStudio=true",
		"FPSCapValues=[30.000000, 60.000000, 75.000000, 120.000000, 144.000000, 165.000000, 240.000000, 360.000000]",
		"FPSCapSelection=0",
		"FPSCap=0.000000",
		"CheckForUpdates=false",
		"NonBlockingErrors=true",
		"SilentErrors=true",
		"QuickStart=true",
	}

	settingsFile, err := os.Create(filepath.Join(Dirs.Cache, "settings"))
	Errc(err)
	defer settingsFile.Close()

	// FIXME: compare settings file, to check if user has modified the settings file
	log.Println("Writing custom rbxfpsunlocker settings to", settingsFile.Name())
	for _, setting := range settings {
		_, err := fmt.Fprintln(settingsFile, setting+"\r")
		Errc(err)
	}

	log.Println("Launching FPS Unlocker")
	Exec("wine", fpsUnlockerPath)

	// Since this file is always overwritten, just remove it anyway.
	Errc(os.RemoveAll(settingsFile.Name()))
}
