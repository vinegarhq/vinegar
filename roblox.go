package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func RobloxWriteAppSettings(versionDir string) {
	log.Printf("Writing %s AppSettings file", filepath.Base(versionDir))

	appSettingsFile, err := os.Create(filepath.Join(versionDir, "AppSettings.xml"))
	if err != nil {
		log.Fatal(err)
	}

	appSettings := `
<?xml version="1.0" encoding="UTF-8"?>
<Settings>
        <ContentFolder>content</ContentFolder>
        <BaseUrl>http://www.roblox.com</BaseUrl>
</Settings>
`

	if _, err = appSettingsFile.WriteString(appSettings[1:]); err != nil {
		log.Fatal(err)
	}
}

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

func RobloxPlayerLatestVersion(channel string) string {
	log.Println("Getting latest Roblox client version")

	// thanks pizzaboxer
	resp, err := http.Get("https://setup.rbxcdn.com" + channel + "/version")
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	resp.Body.Close()

	return string(body)
}

// THANKS PIZZABOXER
//
// Hack to parse Roblox.com's given arguments from RobloxPlayerLauncher to
// RobloxPlayerBeta This function is mainly a hack to take place of what the
// launcher would do, and would fork for RobloxPlayerBeta.
func BrowserArgsParse(launchURI string) (string, []string) {
	var channel string

	args := make([]string, 0)
	URIKeyArg := map[string]string{
		"launchmode":       "--",
		"gameinfo":         "-t ",
		"placelauncherurl": "-j ",
		"launchtime":       "--launchtime=",
		"browsertrackerid": "-b ",
		"robloxLocale":     "--rloc ",
		"gameLocale":       "--gloc ",
		"channel":          "-channel ",
	}

	for _, uri := range strings.Split(launchURI, "+") {
		parts := strings.Split(uri, ":")

		if URIKeyArg[parts[0]] == "" || parts[1] == "" {
			continue
		}

		if parts[0] == "launchmode" && parts[1] == "play" {
			parts[1] = "app"
		}

		if parts[0] == "channel" {
			channel = "/channel/" + strings.ToLower(parts[1])
		}

		if parts[0] == "placelauncherurl" {
			urlDecoded, err := url.QueryUnescape(parts[1])

			if err != nil {
				log.Fatal(err)
			}

			parts[1] = urlDecoded
		}

		args = append(args, URIKeyArg[parts[0]]+parts[1])
	}

	return channel, args
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

	if Config.AutoKill {
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
		exe = RobloxFind(false, "RobloxStudioBeta.exe")
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

	if strings.HasPrefix(strings.Join(args, " "), "roblox-player:1+launchmode:") {
		channel, _args := BrowserArgsParse(args[0])

		if filepath.Base(root) == RobloxPlayerLatestVersion(channel) {
			exe = filepath.Join(root, "RobloxPlayerBeta.exe")

			args = _args
		}
	}

	RobloxLaunch(exe, args...)
}
