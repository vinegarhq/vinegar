package dirs

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	Cache     = filepath.Join(xdg.CacheHome, "vinegar")
	Config    = filepath.Join(xdg.ConfigHome, "vinegar")
	Data      = filepath.Join(xdg.DataHome, "vinegar")
	Overlays  = filepath.Join(Config, "overlays")
	Downloads = filepath.Join(Cache, "downloads")
	Logs      = filepath.Join(Cache, "logs")
	Prefixes  = filepath.Join(Data, "prefixes")
	Versions  = filepath.Join(Data, "versions")
)

var (
	StatePath  = filepath.Join(Data, "state.json")
	ConfigPath = filepath.Join(Config, "config.toml")
	WinePath   = filepath.Join(Data, "kombucha")
)
