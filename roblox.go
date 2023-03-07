package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const RCOURL = "https://raw.githubusercontent.com/L8X/Roblox-Client-Optimizer/main/ClientAppSettings.json"

// Loops over the global program directories, searching for Roblox's
// version directory with a match of the given executable:
//
//	programdir/Roblox/Versions/version-XXXXXXXX/exe
//
// For Roblox Studio, this is simply at the root of the versions directory.
// giveDir is used mainly to get the location of the executable's directory,
// which can be used to place the FFlags file into.
func RobloxFind(giveDir bool, exe string) string {
	for _, programDir := range programDirs {
		versionDir := filepath.Join(programDir, "Roblox/Versions")

		// Studio
		rootExe := filepath.Join(versionDir, exe)
		if _, e := os.Stat(rootExe); e == nil {
			if !giveDir {
				return rootExe
			}

			return versionDir
		}

		// Glob for programdir/Roblox/Versions/version-*/exe
		versionExe, _ := filepath.Glob(filepath.Join(versionDir, "*", exe))

		if versionExe == nil {
			continue
		}

		if !giveDir {
			return versionExe[0]
		}

		return filepath.Dir(versionExe[0])
	}

	return ""
}

// Simply download the given URL and execute it, which is used
// only for installing Roblox, as the default behavior of launching
// Roblox's launchers is to always install it. Afterwards, we remove
// the installer as we no longer need it.
func RobloxInstall(url string) error {
	log.Println("Installing Roblox")

	installerPath := filepath.Join(Dirs.Cache, "rbxinstall.exe")

	if err := Download(url, installerPath); err != nil {
		return err
	}

	if err := Exec("wine", true, installerPath); err != nil {
		return err
	}

	if err := os.RemoveAll(installerPath); err != nil {
		return err
	}

	return nil
}

// Validate the given renderer, and apply it to the given map (fflags);
// It will also disable every other renderer.
func RobloxSetRenderer(renderer string, fflags map[string]interface{}) {
	possibleRenderers := []string{
		"OpenGL",
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	validRenderer := false

	for _, r := range possibleRenderers {
		if renderer == r {
			validRenderer = true
		}
	}

	if !validRenderer {
		log.Fatal("invalid renderer, must be one of:", possibleRenderers)
	}

	for _, r := range possibleRenderers {
		isRenderer := r == renderer
		fflags["FFlagDebugGraphicsPrefer"+r] = isRenderer
		fflags["FFlagDebugGraphicsDisable"+r] = !isRenderer
	}
}

// Create the fflags directory;which sebsequently contains the FFlags
// file, which is also created immediately afterwards, and if the configuration
// specifies to use RCO, the RCO FFlags file overrwrites the FFlags file
// entirely. Checking if the file is empty is necessary to prevent JSON
// Unmarshalling errors, for when RCO hasn't been applied.
// Afterwards we set the user-set FFlags, which is then idented to look pretty
// and then written to the fflags file.
func RobloxApplyFFlags(app string, dir string) error {
	fflags := make(map[string]interface{})

	if app == "Studio" {
		return nil
	}

	fflagsDir := filepath.Join(dir, app+"Settings")
	CheckDirs(DirMode, fflagsDir)

	fflagsFile, err := os.Create(filepath.Join(fflagsDir, app+"AppSettings.json"))
	if err != nil {
		return err
	}

	if Config.ApplyRCO {
		log.Println("Applying RCO FFlags")

		if err := Download(RCOURL, fflagsFile.Name()); err != nil {
			return err
		}
	}

	fflagsFileContents, err := os.ReadFile(fflagsFile.Name())
	if err != nil {
		return err
	}

	if string(fflagsFileContents) != "" {
		if err := json.Unmarshal(fflagsFileContents, &fflags); err != nil {
			return err
		}
	}

	log.Println("Applying custom FFlags")

	RobloxSetRenderer(Config.Renderer, fflags)

	for fflag, value := range Config.FFlags {
		fflags[fflag] = value
	}

	fflagsJSON, err := json.MarshalIndent(fflags, "", "  ")
	if err != nil {
		return err
	}

	if _, err := fflagsFile.Write(fflagsJSON); err != nil {
		return err
	}

	return nil
}

// Create or set Microsoft's Edge installation directories to be read only,
// since Studio for whatever reason will install it, though it is not needed.
// Check if the versions (with the exe) exists, if not install the Roblox Client,
// which brings with it the Studio installer in the root of the versions directory.
// However, if roblox has still not been found, exit immediately. Apply fflags,
// Install DXVK when specified in configuration.
// Additionally, launch wine with the specified 'launcher' in the configuration
// when set, for example: <launcher> wine ../RobloxPlayerLauncher.exe
// The arguments seen below is a result of Go's slice arrays and wanting to add
// custom arguments, so i simply chose to append what we need to the array instead.
// When enabled in configuration, we also wait for Roblox to exit (queried with the given string)
//
//	RobloxPlayerLauncher.exe -> RobloxPlayerLau
//	RobloxPlayerLauncher.exe -> RobloxPlayer + Bet
//
// Linux's procfs 'comm' file has a maximum string length of 16 characters, which is
// why using sliced [:15] is preffered to provided directly.
func RobloxLaunch(exe string, app string, args ...string) {
	EdgeDirSet(DirROMode, true)

	if RobloxFind(false, exe) == "" {
		if err := RobloxInstall("https://www.roblox.com/download/client"); err != nil {
			log.Fatal("failed to install roblox: ", err)
		}
	}

	robloxRoot := RobloxFind(true, exe)

	if robloxRoot == "" {
		log.Fatal("failed to find roblox")
	}

	DxvkStrap()

	if err := RobloxApplyFFlags(app, robloxRoot); err != nil {
		log.Fatal("failed to apply fflags: ", err)
	}

	log.Println("Launching", exe)
	args = append([]string{filepath.Join(robloxRoot, exe)}, args...)

	prog := "wine"

	if Config.Launcher != "" {
		args = append([]string{"wine"}, args...)
		prog = Config.Launcher
	}

	if err := Exec(prog, true, args...); err != nil {
		log.Fatal("roblox exec err: ", err)
	}

	if Config.AutoKillPfx {
		CommLoop(exe[:15])
		CommLoop(exe[:12] + "Bet")
		PfxKill()
	}
}

func RobloxStudio(args ...string) {
	exe := "RobloxStudioLauncherBeta.exe"

	// Protocol URI, Launcher cannot be used
	if len(args) < 1 {
		args = []string{"-ide"}
	} else if len(args[0]) > 12 && args[0][:13] == "roblox-studio" {
		exe = "RobloxStudioBeta.exe"
	}

	RobloxLaunch(exe, "Studio", args...)
}
