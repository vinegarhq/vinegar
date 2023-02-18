// Copyright vinegar-development 2023

package main

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path/filepath"
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

			if !giveDir {
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
	installerPath := filepath.Join(Dirs.Cache, "rbxinstall.exe")
	Download(url, installerPath)
	Exec("wine", true, installerPath)
	Errc(os.RemoveAll(installerPath))
}

// Write to the FFlags with the configuration's preferred renderer
// and FFlags.
func RobloxApplyFFlags(dir string) {
	flags := make(map[string]interface{})

	log.Println("Applying FFlags")

	fflagsDir := filepath.Join(dir, "ClientSettings")
	CheckDirs(0755, fflagsDir)

	// Create an empty fflags file
	fflagsFile, err := os.Create(filepath.Join(fflagsDir, "ClientAppSettings.json"))
	Errc(err)

	if Config.ApplyRCO {
		ApplyRCOFFlags(fflagsFile.Name())
	}

	// Apply our renderers overrides
	possibleRenderers := []string{
		"OpenGL",
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	for _, rend := range possibleRenderers {
		isRenderer := rend == Config.Renderer
		Config.FFlags["FFlagDebugGraphicsPrefer"+rend] = isRenderer
		Config.FFlags["FFlagDebugGraphicsDisable"+rend] = !isRenderer
	}

	// Read the file
	fflags, err := os.ReadFile(fflagsFile.Name())
	Errc(err)
	json.Unmarshal(fflags, &flags)

	// Now let's add our own fflags
	for flag, value := range Config.FFlags {
		flags[flag] = value
	}

	// Finally, write the file
	finalFFlagsFile, err := json.MarshalIndent(flags, "", "  ")
	log.Println(fflagsFile.Name())

	Errc(err)
	Errc(os.WriteFile(fflagsFile.Name(), finalFFlagsFile, 0644))
}

func EdgeDirSet(perm uint32, create bool) {
	user, err := user.Current()
	Errc(err)

	var programDirs = []string{
		filepath.Join("users", user.Username, "AppData/Local"),
		"Program Files (x86)",
		"Program Files",
	}

	for _, programDir := range programDirs {
		EdgeDir := filepath.Join(Dirs.Pfx, "drive_c", programDir, "Microsoft", "EdgeUpdate")

		if _, err := os.Stat(EdgeDir); os.IsNotExist(err) && create {
			CheckDirs(0755, filepath.Dir(EdgeDir))
			CheckDirs(perm, EdgeDir)
		} else if os.IsExist(err) {
			err := os.Chmod(EdgeDir, os.FileMode(perm))
			if err != nil && create {
				Errc(err)
			}

			log.Println("Setting", EdgeDir, "to", os.FileMode(perm))
		} else {
			Errc(err)
		}
	}
}

// Launch the given Roblox executable, finding it from RobloxFind().
// When it is not found, it is fetched and installed. additionally,
// pass vinegar's command line with the Roblox executable pre-appended.
func RobloxLaunch(exe, url string, installFFlagPlayer bool, args ...string) {
	EdgeDirSet(0644, true)

	if RobloxFind(false, exe) == "" {
		RobloxInstall(url)
	}

	// Wait for Roblox{Studio,Player}Launcher
	// to finish installing, as sometimes that could happen
	CommLoop(exe[:15])

	robloxRoot := RobloxFind(true, exe)

	if installFFlagPlayer {
		RobloxApplyFFlags(robloxRoot)
	}

	log.Println("Launching", exe)
	args = append([]string{filepath.Join(robloxRoot, exe)}, args...)

	if Config.GameMode {
		args = append([]string{"wine"}, args...)
		Exec("gamemoderun", true, args...)
	} else {
		Exec("wine", true, args...)
	}

	if Config.AutoRFPSU {
		RbxFpsUnlocker()
	}
}
