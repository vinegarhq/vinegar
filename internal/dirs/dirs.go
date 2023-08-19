package dirs

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	Cache      = filepath.Join(xdg.CacheHome, "aubun")
	Config     = filepath.Join(xdg.ConfigHome, "aubun")
	Data       = filepath.Join(xdg.DataHome, "aubun")
	Overlay    = filepath.Join(Config, "overlay")
	Downloads  = filepath.Join(Cache, "downloads")
	Logs       = filepath.Join(Cache, "logs")
	Prefix     = filepath.Join(Data, "prefix")
	PrefixData = filepath.Join(Prefix, "aubun")
	Versions   = filepath.Join(PrefixData, "versions")
)

func Mkdirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}
