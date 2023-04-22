package main

import (
	"errors"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinegarhq/vinegar/util"
)

const RBXCDNURL = "https://setup.rbxcdn.com/"

type Roblox struct {
	BaseURL      string
	Version      string
	VersionDir   string
	PackageDests map[string]string
	Packages
}

func (r *Roblox) AppSettings() {
	log.Printf("Writing %s AppSettings file", r.Version)

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
		appSettingsFile.Close()
		log.Fatal(err)
	}
}

func (r *Roblox) SetupURL(channel string) {
	r.BaseURL = RBXCDNURL

	if channel != "" && channel != "live" && channel != "LIVE" {
		log.Printf("Warning: Using user channel %s", channel)

		r.BaseURL += "channel/" + channel + "/"
	}
}

func (r *Roblox) GetVersion(versionFile string) {
	version, err := util.URLBody(r.BaseURL + versionFile)
	if err != nil {
		log.Fatal(err)
	}

	r.Version = version
}

func (r *Roblox) Install() {
	PfxInit()
	log.Println("Installing Roblox", r.Version)

	r.Packages = GetPackages(r.BaseURL + r.Version)
	r.Packages.Download()
	r.Packages.Extract(r.VersionDir, r.PackageDests)
	r.AppSettings()
}

func (r *Roblox) Setup() {
	r.VersionDir = filepath.Join(Dirs.Versions, r.Version)

	log.Println("Checking for Roblox", r.Version)

	if _, err := os.Stat(r.VersionDir); errors.Is(err, os.ErrNotExist) {
		r.Install()
	} else if err != nil {
		log.Fatal(err)
	}

	DxvkStrap()
}

// THANKS PIZZABOXER.
func BrowserArgsParse(launchURI string) (string, []string) {
	chnl := ""
	args := make([]string, 0)
	uris := map[string]string{
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

		if uris[parts[0]] == "" || parts[1] == "" {
			continue
		}

		if parts[0] == "launchmode" && parts[1] == "play" {
			parts[1] = "app"
		}

		if parts[0] == "channel" {
			chnl = strings.ToLower(parts[1])
		}

		if parts[0] == "placelauncherurl" {
			urlDecoded, err := url.QueryUnescape(parts[1])
			if err != nil {
				log.Fatal(err)
			}

			parts[1] = urlDecoded
		}

		args = append(args, uris[parts[0]]+parts[1])
	}

	return chnl, args
}

func (r *Roblox) Execute(exe string, args []string) {
	log.Println("Launching", exe, r.Version)

	args = append([]string{"wine", filepath.Join(r.VersionDir, exe)}, args...)

	if Config.Launcher != "" {
		args = append([]string{Config.Launcher}, args...)
	}

	log.Println(args)
	robloxCmd := exec.Command(args[0], args[1:]...)

	logFile := LogFile(exe)
	robloxCmd.Stderr = logFile
	robloxCmd.Stdout = logFile
	log.Println("Wine log file:", logFile.Name())

	if err := robloxCmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func RobloxPlayer(args ...string) {
	var rblx Roblox
	var channel string

	if strings.HasPrefix(strings.Join(args, " "), "roblox-player:1+launchmode:") {
		channel, args = BrowserArgsParse(args[0])
	}

	rblx.PackageDests = PlayerPackages()
	rblx.SetupURL(channel)
	rblx.GetVersion("version")

	rblx.Setup()
	rblx.ApplyFFlags("Client")
	rblx.Execute("RobloxPlayerBeta.exe", args)
	PfxKill()
}

func RobloxStudio(args ...string) {
	var rblx Roblox

	rblx.PackageDests = StudioPackages()
	rblx.SetupURL("LIVE")
	rblx.GetVersion("versionQTStudio")
	Config.Dxvk = false // Dxvk doesnt work under Studio

	rblx.Setup()
	rblx.ApplyFFlags("Studio")
	rblx.Execute("RobloxStudioBeta.exe", args)
	PfxKill()
}
