package main

import (
	"errors"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Roblox struct {
	File       string
	Path       string
	Version    string
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

func (r *Roblox) Install(directories map[string]string) {
	var pkgmanif PackageManifest

	PfxInit()
	log.Println("Installing Roblox", r.Version)
	pkgmanif.Version = r.Version
	pkgmanif.Construct()
	pkgmanif.DownloadVerifyExtractAll(r.VersionDir, directories)
	r.AppSettings()
}

func (r *Roblox) Setup(directories map[string]string) {
	r.VersionDir = filepath.Join(Dirs.Versions, r.Version)

	log.Println("Checking for Roblox", r.File, r.Version)

	if _, err := os.Stat(r.VersionDir); errors.Is(err, os.ErrNotExist) {
		r.Install(directories)
	} else if err != nil {
		log.Fatal(err)
	}

	exePath, err := filepath.Abs(filepath.Join(r.VersionDir, r.File))
	if err != nil {
		log.Fatal(err)
	}

	DxvkStrap()

	r.Path = exePath
}

// THANKS PIZZABOXER
//
// Hack to parse Roblox.com's given arguments from RobloxPlayerLauncher to
// RobloxPlayerBeta This function is mainly a hack to take place of what the
// launcher would do, and would fork for RobloxPlayerBeta.
func BrowserArgsParse(launchURI string) []string {
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

func (r *Roblox) GetNewestLogFile() string {
	return NewestFile(filepath.Join(AppDataDir, "Local", "Roblox", "logs", "*.log"))
}

func (r *Roblox) Execute(args []string) {
	log.Println("Launching", r.File, r.Version)

	args = append([]string{"wine", r.Path}, args...)

	if Config.Launcher != "" {
		args = append([]string{Config.Launcher}, args...)
	}

	log.Println(args)
	robloxCmd := exec.Command(args[0], args[1:]...)

	logFile := LogFile(r.File)
	log.Println("Wine log file:", logFile.Name())
	robloxCmd.Stderr = logFile
	robloxCmd.Stdout = logFile
	err := robloxCmd.Run()

	log.Println("Roblox log file:", r.GetNewestLogFile())

	// exit code 41 is a false alarm.
	if err != nil && err.Error() != "exit status 41" {
		log.Fatalf("roblox exec err: %s", err)
	}
}

func RobloxPlayer(args ...string) {
	directories := map[string]string{
		"RobloxApp.zip":                 "",
		"shaders.zip":                   "shaders",
		"ssl.zip":                       "ssl",
		"content-avatar.zip":            "content/avatar",
		"content-configs.zip":           "content/configs",
		"content-fonts.zip":             "content/fonts",
		"content-sky.zip":               "content/sky",
		"content-sounds.zip":            "content/sounds",
		"content-textures2.zip":         "content/textures",
		"content-models.zip":            "content/models",
		"content-textures3.zip":         "PlatformContent/pc/textures",
		"content-terrain.zip":           "PlatformContent/pc/terrain",
		"content-platform-fonts.zip":    "PlatformContent/pc/fonts",
		"extracontent-luapackages.zip":  "ExtraContent/LuaPackages",
		"extracontent-translations.zip": "ExtraContent/translations",
		"extracontent-models.zip":       "ExtraContent/models",
		"extracontent-textures.zip":     "ExtraContent/textures",
		"extracontent-places.zip":       "ExtraContent/places",
	}

	var rblx Roblox
	rblx.File = "RobloxPlayerBeta.exe"
	rblx.Version = GetLatestVersion("version")
	rblx.Setup(directories)
	rblx.ApplyFFlags("Client")

	if strings.HasPrefix(strings.Join(args, " "), "roblox-player:1+launchmode:") {
		args = BrowserArgsParse(args[0])
	}

	rblx.Execute(args)
	PfxKill()
}

func RobloxStudio(args ...string) {
	directories := map[string]string{
		"ApplicationConfig.zip":        "ApplicationConfig",
		"BuiltInPlugins.zip":           "BuiltInPlugins",
		"BuiltInStandalonePlugins.zip": "BuiltInStandalonePlugins",
		"Plugins.zip":                  "Plugins",
		"Qml.zip":                      "Qml",
		"StudioFonts.zip":              "StudioFonts",
		// "WebView2.zip": "",
		// "WebView2RuntimeInstaller.zip": "",
		"RobloxStudio.zip":                "",
		"Libraries.zip":                   "",
		"LibrariesQt5.zip":                "",
		"content-avatar.zip":              "content/avatar",
		"content-configs.zip":             "content/configs",
		"content-fonts.zip":               "content/fonts",
		"content-models.zip":              "content/models",
		"content-qt_translations.zip":     "content/qt_translations",
		"content-sky.zip":                 "content/sky",
		"content-sounds.zip":              "content/sounds",
		"shaders.zip":                     "shaders",
		"ssl.zip":                         "ssl",
		"content-textures2.zip":           "content/textures",
		"content-textures3.zip":           "PlatformContent/pc/textures",
		"content-studio_svg_textures.zip": "content/studio_svg_textures",
		"content-terrain.zip":             "PlatformContent/pc/terrain",
		"content-platform-fonts.zip":      "PlatformContent/pc/fonts",
		"content-api-docs.zip":            "content/api_docs",
		"extracontent-scripts.zip":        "ExtraContent/scripts",
		"extracontent-luapackages.zip":    "ExtraContent/LuaPackages",
		"extracontent-translations.zip":   "ExtraContent/translations",
		"extracontent-models.zip":         "ExtraContent/models",
		"extracontent-textures.zip":       "ExtraContent/textures",
		"redist.zip":                      "",
	}

	var rblx Roblox
	rblx.File = "RobloxStudioBeta.exe"
	rblx.Version = GetLatestVersion("versionQTStudio")
	Config.Dxvk = false // Dxvk doesnt work under studio

	rblx.Setup(directories)
	rblx.ApplyFFlags("Studio")
	rblx.Execute(args)

	PfxKill()
}
