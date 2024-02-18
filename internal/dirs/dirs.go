package dirs

import (
	"os"
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

	// Deprecated: Vinegar supports multiple wine prefixes
	Prefix = filepath.Join(Data, "prefix")
	// Deprecated: Vinegar supports multiple overlays for each Player and Studio
	Overlay = filepath.Join(Config, "overlay")
)

func Mkdirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}
