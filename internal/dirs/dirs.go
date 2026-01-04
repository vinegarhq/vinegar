package dirs

import (
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

var (
	Cache      = filepath.Join(xdg.CacheHome, "vinegar")
	Config     = filepath.Join(xdg.ConfigHome, "vinegar")
	Data       = filepath.Join(xdg.DataHome, "vinegar")
	Overlays   = filepath.Join(Config, "overlays")
	Logs       = filepath.Join(Cache, "logs")
	Prefixes   = filepath.Join(Data, "prefixes")
	Deployment = filepath.Join(Data, "deployment")
)

var (
	StatePath   = filepath.Join(Data, "state.json")
	ConfigPath  = filepath.Join(Config, "config.toml")
	WinePath    = filepath.Join(Data, "kombucha")
	AppDataPath = filepath.Join(Data, "appdata")
)

func Windows(name string) string {
	// You never know.
	if !filepath.IsAbs(name) {
		panic("dirs: unhandled local path")
	}
	return "Z:" + strings.ReplaceAll(name, "/", "\\")
}
