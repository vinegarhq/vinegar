package bootstrapper

import (
	"github.com/vinegarhq/vinegar/roblox"
)

// PackageDirectories is a map of where Binary packages should go.
type PackageDirectories map[string]string

// BinaryDirectories retrieves the PackageDirectories for the given [roblox.BinaryType].
func BinaryDirectories(t roblox.BinaryType) PackageDirectories {
	switch t {
	case roblox.Player:
		return PlayerDirectories
	case roblox.Studio:
		return StudioDirectories
	}

	return nil
}

// PlayerDirectories is retrieved from [Bloxstrap].
//
// [Bloxstrap]: https://github.com/pizzaboxer/bloxstrap/blob/main/Bloxstrap/Bootstrapper.cs
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

// StudioDirectories is retrieved from [Roblox-Studio-Mod-Manager].
//
// [Roblox-Studio-Mod-Manager]: https://github.com/MaximumADHD/Roblox-Studio-Mod-Manager/blob/main/Config/KnownRoots.json
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
