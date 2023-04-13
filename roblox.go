package main

import (
	"log"
	"net/url"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Roblox struct {
	File string
	Path string
	Version string
	VersionDir string
}

func (r *Roblox) AppSettings() {
	log.Printf("Writing %s AppSettings file", filepath.Base(r.VersionDir))

	appSettingsFile, err := os.Create(filepath.Join(r.VersionDir, "AppSettings.xml"))
	if err != nil {
		log.Fatal(err)
	}

	appSettings := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n" +
                   "<Settings>\r\n" +
                   "        <ContentFolder>content</ContentFolder>\r\n" +
                   "        <BaseUrl>http://www.roblox.com</BaseUrl>\r\n" +
                   "</Settings>\r\n"

	if _, err = appSettingsFile.WriteString(appSettings); err != nil {
		log.Fatal(err)
	}
}

func (r *Roblox) Install() {
	var pkgmanif PackageManifest

	log.Println("Installing Roblox", r.Version)

	pkgmanif.Version = r.Version
	pkgmanif.Construct()
	pkgmanif.DownloadVerifyAll()
	pkgmanif.ExtractAll(ClientPackageDirectories())
	r.AppSettings()
}

func (r *Roblox) Setup() {
	r.VersionDir = filepath.Join(LocalProgramDir, r.Version)

	log.Println("Checking for Roblox", r.File, r.Version)

	if _, err := os.Stat(r.VersionDir); errors.Is(err, os.ErrNotExist) {
		r.Install()
	} else if err != nil {
		log.Fatal(err)
	}

	exePath, err := filepath.Abs(filepath.Join(r.VersionDir, r.File))
	if err != nil {
		log.Fatal(err)
	}
	
	r.Path = exePath
}

// THANKS PIZZABOXER
//
// Hack to parse Roblox.com's given arguments from RobloxPlayerLauncher to
// RobloxPlayerBeta This function is mainly a hack to take place of what the
// launcher would do, and would fork for RobloxPlayerBeta.
func BrowserArgsParse(launchURI string) []string {
//	var channel string

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

//		if parts[0] == "channel" {
//			channel = "/channel/" + strings.ToLower(parts[1])
//		}

		if parts[0] == "placelauncherurl" {
			urlDecoded, err := url.QueryUnescape(parts[1])

			if err != nil {
				log.Fatal(err)
			}

			parts[1] = urlDecoded
		}

		args = append(args, URIKeyArg[parts[0]]+parts[1])
	}

	return args
}

func (r *Roblox) Execute(args ...string) {
	log.Println("Launching", r.File, r.Version)
	log.Println(r.Path)
	args = append([]string{r.Path}, args...)

	prog := "wine"

	if Config.Launcher != "" {
		args = append([]string{"wine"}, args...)
		prog = Config.Launcher
	}

	err := Exec(prog, false, args...)

	// exit code 41 is a false alarm.
	if err != nil && err.Error() != "exit status 41" {
		log.Fatal("roblox exec err: ", err)
	}

//	if Config.AutoKill {
//		// Probably wouldn't want to use the full path of the EXE.
//		exeName := filepath.Base(exe)
//		CommLoop(exeName[:15])
//		CommLoop(exeName[:12] + "Bet")
//		PfxKill()
//	}
}

// Handler for launching Roblox Studio, this is required since roblox-studio-auth
// passes special arguments meant for RobloxStudioBeta.
// Aditionally pass -ide if needed, and disable DXVK.
//func RobloxStudio(args ...string) {
//	exe := RobloxSetup("RobloxStudioLauncherBeta.exe")
//
//	// Protocol URI, Launcher cannot be used
//	if len(args) < 1 {
//		args = []string{"-ide"}
//	} else if strings.HasPrefix(strings.Join(args, " "), "roblox-studio") {
//		exe = RobloxFind(false, "RobloxStudioBeta.exe")
//	}
//
//	// DXVK does not work under studio.
//	Config.Dxvk = false
//
//	RobloxLaunch(exe, args...)
// 	log.Println(args)
// }

// Handler for launching Roblox Player, which checks if we are being launched
// from the browser, and check if we have the latest Roblox version, if we do,
// use RobloxPlayerBeta directly, which then we would need to format the Roblox
// arguments ourselves, which is RobloxPlayerLauncher's job.
// Using RobloxPlayerBeta can shave some time off when launching Roblox.
//func RobloxPlayer(version string, args ...string) {
//	exe := filepath.Join(LocalProgramDir, version, "RobloxPlayerBeta.exe")
//	RobloxLaunch(exe, args...)
//}
