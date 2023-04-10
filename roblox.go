package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

	return os.RemoveAll(installerPath)
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

// Hack to parse Roblox.com's given arguments from RobloxPlayerLauncher to
// RobloxPlayerBeta This function is mainly a hack to take place of what the
// launcher would do, and would fork for RobloxPlayerBeta.
func BrowserArgsParse(args *string) {
	// roblox-player 1 launchmode play gameinfo
	// {authticket} launchtime {launchtime} placelauncherurl
	// {placelauncherurl} browsertrackerid {browsertrackerid}
	// robloxLocale {rloc} gameLocale {gloc} channel
	rbxArgs := regexp.MustCompile(`[\:\,\+\s]+`).Split(*args, -1)

	placeLauncherURLDecoded, err := url.QueryUnescape(rbxArgs[9])
	if err != nil {
		log.Fatal(err)
	}

	// Forwarded command line as of 2023-02-03:
	// RobloxPlayerBeta -t {authticket} -j {placelauncherurl} -b 0
	//   --launchtime={launchtime} --rloc {rloc} --gloc {gloc}
	*args = "--app" + " -t " + rbxArgs[5] + " -j " + placeLauncherURLDecoded + " -b 0 " +
		"--launchtime=" + rbxArgs[7] + " --rloc " + rbxArgs[13] + " --gloc " + rbxArgs[15]
}

// Create or set Microsoft's Edge installation directories to be read only,
// since Studio will install it, which is needed for the Login page, Studio's
// Edge installation is broken on Wine. Vinegar may install Edge on it's own
// in the future if agreed upon.
//
// Check if the versions (with the exe) exists, if not install the Roblox Client,
// which brings with it the Studio installer in the root of the versions directory.
// However, if roblox has still not been found, exit immediately.
// Install DXVK - if needed, then apply the FFlags.
//
// If we are going to launch the Roblox Client, we compare the version ourselves,
// and use the Roblox Launcher if it's mismatched. This is to shave off some time
// when launching Roblox, since the Launcher will always check if the installation
// is outdated, then install it if necessary, and pass to RobloxPlayerBeta.
func RobloxSetup(exe string) string {
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

	return filepath.Join(robloxRoot, exe)
}

// launch wine with the specified 'launcher' in the configuration
// when set, for example: <launcher> wine ../RobloxPlayerLauncher.exe
//
// Linux's procfs 'comm' file has a maximum string length of 16 characters,
// which is why using sliced [:15] is preffered to provided directly:
//
//	RobloxPlayerLauncher.exe -> RobloxPlayerLau
//	RobloxPlayerLauncher.exe -> RobloxPlayer + Bet -> RobloxPlayerBeta
//
// When enabled in configuration, we also wait for Roblox to exit.
func RobloxLaunch(exe string, args ...string) {
	log.Println("Launching", exe)
	args = append([]string{exe}, args...)

	prog := "wine"

	if Config.Launcher != "" {
		args = append([]string{"wine"}, args...)
		prog = Config.Launcher
	}

	err := Exec(prog, true, args...)

	// exit code 41 is a false alarm.
	if err != nil && err.Error() != "exit status 41" {
		log.Fatal("roblox exec err: ", err)
	}

	if Config.AutoKillPfx {
		// Probably wouldn't want to use the full path of the EXE.
		exeName := filepath.Base(exe)
		CommLoop(exeName[:15])
		CommLoop(exeName[:12] + "Bet")
		PfxKill()
	}
}

// Handler for launching Roblox Studio, this is required since roblox-studio-auth
// passes special arguments meant for RobloxStudioBeta.
// Aditionally pass -ide if needed, and disable DXVK.
func RobloxStudio(args ...string) {
	exe := RobloxSetup("RobloxStudioLauncherBeta.exe")

	// Protocol URI, Launcher cannot be used
	if len(args) < 1 {
		args = []string{"-ide"}
	} else if strings.HasPrefix(strings.Join(args, " "), "roblox-studio") {
		exe = "RobloxStudioBeta.exe"
	}

	// DXVK does not work under studio.
	Config.Dxvk = false

	RobloxLaunch(exe, args...)
}

// Handler for launching Roblox Player, which checks if we are being launched
// from the browser, and check if we have the latest Roblox version, if we do,
// use RobloxPlayerBeta directly, which then we would need to format the Roblox
// arguments ourselves, which is RobloxPlayerLauncher's job.
// Using RobloxPlayerBeta can shave some time off when launching Roblox.
func RobloxPlayer(args ...string) {
	exe := RobloxSetup("RobloxPlayerLauncher.exe")
	root := filepath.Dir(exe)

	if err := RobloxApplyFFlags("Client", root); err != nil {
		log.Fatal("failed to apply fflags: ", err)
	}

	if strings.HasPrefix(strings.Join(args, " "), "roblox-player:1+launchmode:play") {
		if filepath.Base(root) == RobloxPlayerLatestVersion() {
			exe = filepath.Join(root, "RobloxPlayerBeta.exe")

			BrowserArgsParse(&args[0])
		}
	}

	RobloxLaunch(exe, args...)
}
