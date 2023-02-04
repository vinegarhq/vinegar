// Copyright vinegar-development 2023

package vinegar

import (
	"log"
	"regexp"
	"net/url"
	"os"
	"path/filepath"
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

func RobloxLaunch(dirs *Dirs, exe string, url string, arg string) {
	if RobloxFind(dirs, false, exe) == "" {
		installerPath := filepath.Join(dirs.Cache, "rbxinstall.exe")
		Download(url, installerPath)
		Exec(dirs, "wine", installerPath, "-fast")
		Errc(os.RemoveAll(installerPath))
	}

	log.Println("Launching", exe)
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
