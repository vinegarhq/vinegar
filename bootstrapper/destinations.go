package bootstrapper

type PackageDirectories map[string]string

func (bt BinaryType) Directories() PackageDirectories {
	switch bt {
	case Player:
		return PlayerDirectories
	case Studio:
		return StudioDirectories
	}

	return nil
}

/* github.com/pizzaboxer/bloxstrap/blob/main/Bloxstrap/Bootstrapper.cs */
var PlayerDirectories = PackageDirectories{
	"RobloxApp.zip":                 "",
	"shaders.zip":                   "shaders/",
	"ssl.zip":                       "ssl/",
	"WebView2.zip":                  "",
	"WebView2RuntimeInstaller.zip":  "WebView2RuntimeInstaller",
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

/* github.com/MaximumADHD/Roblox-Studio-Mod-Manager/blob/main/Config/KnownRoots.json */
/* jq 'reduce to_entries[] as $item ({}; . + {($item.key): $item.value.ExtractTo})' */
var StudioDirectories = PackageDirectories{
	"BuiltInPlugins.zip":              "BuiltInPlugins",
	"ApplicationConfig.zip":           "ApplicationConfig",
	"BuiltInStandalonePlugins.zip":    "BuiltInStandalonePlugins",
	"content-qt_translations.zip":     "content/qt_translations",
	"content-platform-fonts.zip":      "PlatformContent/pc/fonts",
	"content-terrain.zip":             "PlatformContent/pc/terrain",
	"content-textures3.zip":           "PlatformContent/pc/textures",
	"extracontent-translations.zip":   "ExtraContent/translations",
	"extracontent-luapackages.zip":    "ExtraContent/LuaPackages",
	"extracontent-textures.zip":       "ExtraContent/textures",
	"extracontent-scripts.zip":        "ExtraContent/scripts",
	"extracontent-models.zip":         "ExtraContent/models",
	"content-sky.zip":                 "content/sky",
	"content-fonts.zip":               "content/fonts",
	"content-avatar.zip":              "content/avatar",
	"content-models.zip":              "content/models",
	"content-sounds.zip":              "content/sounds",
	"content-configs.zip":             "content/configs",
	"content-api-docs.zip":            "content/api_docs",
	"content-textures2.zip":           "content/textures",
	"content-studio_svg_textures.zip": "content/studio_svg_textures",
	"Qml.zip":                         "Qml",
	"ssl.zip":                         "ssl",
	"Plugins.zip":                     "Plugins",
	"shaders.zip":                     "shaders",
	"StudioFonts.zip":                 "StudioFonts",
	"redist.zip":                      "",
	"WebView2.zip":                    "",
	"Libraries.zip":                   "",
	"LibrariesQt5.zip":                "",
	"RobloxStudio.zip":                "",
	"WebView2RuntimeInstaller.zip":    "",
}

var ExcludedPackages = []string{
	"RobloxPlayerLauncher.exe",
	"WebView2RuntimeInstaller.zip",
}

func IsExcluded(name string) bool {
	for _, ex := range ExcludedPackages {
		if name == ex {
			return true
		}
	}

	return false
}
