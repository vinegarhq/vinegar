// Copyright vinegar-development 2023

package vinegar

import (
	"log"
	"regexp"
	"net/url"
	"os"
	"path/filepath"
)

// Search for Roblox's Version directories for a given exe, when
// giveDir is passed, it will give the exe's base directory instead of the 
// full path of the final Roblox executable.
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

// Launch the given Roblox executable, finding it from RobloxFind().
// When it is not found, it is fetched and installed.
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

// Hack to parse Roblox.com's given arguments from RobloxPlayerLauncher to RobloxPlayerBeta
// This function is mainly a hack to take place of what the launcher would do, and would fork
// for RobloxPlayerBeta, but due to unsolved sandboxing issues, we do these ourselves.
func BrowserArgsParse(arg string) (string) {
	// roblox-player 1 launchmode play gameinfo 
	// {authticket} launchtime {launchtime} placelauncherurl 
	// {placelauncherurl} browsertrackerid {browsertrackerid}
	// robloxLocale {rloc} gameLocale {gloc} channel
	rbxArgs := regexp.MustCompile("[\\:\\,\\+\\s]+").Split(arg, -1)

	log.Println(rbxArgs)
 	placeLauncherUrlDecoded, err := url.QueryUnescape(rbxArgs[9])
	Errc(err)
	log.Println(placeLauncherUrlDecoded)

	// Forwarded command line as of 2023-02-03:
	// RobloxPlayerBeta -t {authticket} -j {placelauncherurl} -b 0 --launchtime={launchtime} --rloc {rloc} --gloc {gloc}
	return "--app " + "-t " + rbxArgs[5] + " -j " + placeLauncherUrlDecoded + " -b 0 --launchtime=" + rbxArgs[7] + " --rloc " + rbxArgs[13] + " --gloc " + rbxArgs[15]
}
