// Copyright vinegar-development 2023

package vinegar

import (
	"log"
	"fmt"
	"regexp"
	"net/url"
	"os"
	"path/filepath"
)

const (
	FPSUNLOCKERHASH = "sha256:050fe7c0127dbd4fdc0cecf3ba46248ba7e14d37edba1a54eac40602c130f2f8"
	FPSUNLOCKERLITURL = "https://github.com/axstin/rbxfpsunlocker/releases/download/v4.4.4/rbxfpsunlocker-x64.zip"

	// go-getter functionality
	FPSUNLOCKERURL = FPSUNLOCKERLITURL + "?checksum=" + FPSUNLOCKERHASH
)

func RobloxFind(dirs *Dirs, giveDir bool, exe string) (string) {
	var final string
	user := os.Getenv("USER")

	if user == "" {
		log.Fatal("Failed to get $USER variable")
	}

	var programDirs = []string {
		filepath.Join("users", user, "AppData/Local"),
		"Program Files (x86)",
	}

	for _, programDir := range programDirs {
		versionDir, err := os.Open(filepath.Join(dirs.Pfx, "drive_c", programDir, "Roblox/Versions"))

		if os.IsNotExist(err) {
			continue
		}

		versionDirs, err := versionDir.Readdir(0)

		if err != nil {
			continue
		}

		for _, v := range versionDirs {
			checkExe, err := os.Open(filepath.Join(versionDir.Name(), v.Name(), exe))

			// os.IsNotExist does not work here
			if err != nil {
				continue
			}

			defer checkExe.Close()

			if giveDir == false {	
				final = checkExe.Name()
			} else {
				final = filepath.Join(versionDir.Name(), v.Name())
			}

			break
		}
		defer versionDir.Close()
	}
	return final
}

func RbxFpsUnlocker(dirs *Dirs) {
	fpsUnlockerPath := InitExec(dirs, "rbxfpsunlocker.exe", FPSUNLOCKERURL, "FPS Unlocker")

	var settings = []string {
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

	settingsFile, err := os.Create(filepath.Join(dirs.Cache, "settings"))
	Errc(err)
	defer settingsFile.Close()

	// FIXME: compare settings file, to check if user has modified the settings file
	log.Println("Writing custom rbxfpsunlocker settings to", settingsFile.Name())
	for _, setting := range settings {
		_, err := fmt.Fprintln(settingsFile, setting + "\r")
		Errc(err)
	}

	log.Println("Launching FPS Unlocker")
	Exec(dirs, "wine", fpsUnlockerPath)
}

func RobloxLaunch(dirs *Dirs, exe string, url string, what string, arg string) {
	if RobloxFind(dirs, false, exe) == "" {
		installerPath := InitExec(dirs, exe, url, what)
		Exec(dirs, "wine", installerPath, "-fast")
		Errc(os.RemoveAll(installerPath))
	}

	log.Println("Launching", what)
	Exec(dirs, "wine", RobloxFind(dirs, false, exe), arg)
	// RbxFpsUnlocker(dirs)
}

func BrowserArgsParse(arg string) (string) {
	rbxArgs := regexp.MustCompile("[\\:\\,\\+\\s]+").Split(arg, -1)

	/*
	 * roblox-player 1 launchmode play gameinfo 
	 * {authticket} launchtime {launchtime} placelauncherurl 
	 * {placelauncherurl} browsertrackerid {browsertrackerid}
	 * robloxLocale {rloc} gameLocale {gloc} channel
	 */

	log.Println(rbxArgs)
 	placeLauncherUrlDecoded, err := url.QueryUnescape(rbxArgs[9])
	Errc(err)
	log.Println(placeLauncherUrlDecoded)

	/* 
	 * RobloxPlayerLauncher will parse these and forward them
	 * to RobloxPlayerBeta, due to limitations of Go, we do these
	 * ourselves.
	 */

	/* 
	 * forwarded command line from RobloxPlayerLauncher as of 2022-02-03
	 *   RobloxPlayerBeta -t {authticket} -j {placelauncherurl} -b 0 --launchtime={launchtime} --rloc {rloc} --gloc {gloc}
	 */

	return "--app " + "-t " + rbxArgs[5] + " -j " + placeLauncherUrlDecoded + " -b 0 --launchtime=" + rbxArgs[7] + " --rloc " + rbxArgs[13] + " --gloc " + rbxArgs[15]
}
