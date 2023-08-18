package dirs

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	Cache     = filepath.Join(xdg.CacheHome, "aubun")
	Config    = filepath.Join(xdg.ConfigHome, "aubun")
	Data      = filepath.Join(xdg.DataHome, "aubun")
	Overlay   = filepath.Join(Config, "overlay")
	Downloads = filepath.Join(Cache, "downloads")
	Logs      = filepath.Join(Cache, "logs")
	Prefix    = filepath.Join(Data, "prefix")
	Versions  = filepath.Join(Data, "versions")
)

func Mkdir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}
