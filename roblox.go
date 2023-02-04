// Copyright vinegar-development 2023

package main

import (
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
)

// Search for Roblox's Version directories for a given exe, when
// giveDir is passed, it will give the exe's base directory instead of the
// full path of the final Roblox executable.
func RobloxFind(giveDir bool, exe string) string {
	var final string

	user, err := user.Current()
	Errc(err)

	var programDirs = []string{
		filepath.Join("users", user.Username, "AppData/Local"),
		"Program Files (x86)",
		"Program Files",
	}

	for _, programDir := range programDirs {
		versionDir, err := os.Open(filepath.Join(Dirs.Pfx, "drive_c", programDir, "Roblox/Versions"))

		if os.IsNotExist(err) {
			continue
		}

		studioExe, err := os.Open(filepath.Join(versionDir.Name(), exe))

		if err == nil {
			final = studioExe.Name()
			break
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

// Technically, fetch a url's exe and launch it once. This is used
// for roblox installation since launching the program once will make
// automatically install itself.
func RobloxInstall(url string) {
	log.Println("Installing", url)
	installerPath := filepath.Join(Dirs.Cache, "rbxinstall.exe")
	Download(url, installerPath)
	Exec("wine", installerPath)
	Errc(os.RemoveAll(installerPath))
}

// Launch the given Roblox executable, finding it from RobloxFind().
// When it is not found, it is fetched and installed. additionally,
// pass vinegar's command line with the Roblox executable pre-appended.
func RobloxLaunch(exe, url string, installFFlagPlayer bool, args ...string) {
	if RobloxFind(false, exe) == "" {
		RobloxInstall(url)
	}

	robloxRoot := RobloxFind(true, exe)

	if installFFlagPlayer == true {
		ApplyRCOFFlags(robloxRoot)
	}

	args = append([]string{filepath.Join(robloxRoot, exe)}, args...)
	log.Println("Launching", exe)
	Exec("wine", args...)
}

// Hack to parse Roblox.com's given arguments from RobloxPlayerLauncher to RobloxPlayerBeta
// This function is mainly a hack to take place of what the launcher would do, and would fork
// for RobloxPlayerBeta, but due to unsolved sandboxing issues, we do these ourselves.
func BrowserArgsParse(arg string) string {
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
