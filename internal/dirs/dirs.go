package dirs

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/vinegarhq/vinegar/wine"
)

var (
	Cache     = filepath.Join(xdg.CacheHome, "vinegar")
	Config    = filepath.Join(xdg.ConfigHome, "vinegar")
	Data      = filepath.Join(xdg.DataHome, "vinegar")
	Overlay   = filepath.Join(Config, "overlay")
	Downloads = filepath.Join(Cache, "downloads")
	Logs      = filepath.Join(Cache, "logs")
)

var (
	Prefix_Name         = "prefix"
	PrefixData_Relative = "vinegar"
	Versions_Relative   = filepath.Join(PrefixData_Relative, "versions")
)

func GetPrefixPath(name string) string {
	if name != "" {
		return filepath.Join(Data, Prefix_Name+"-"+name)
	} else {
		return filepath.Join(Data, Prefix_Name)
	}
}

func GetPrefixData(pfx *wine.Prefix) string {
	return filepath.Join(pfx.Dir, PrefixData_Relative)
}

func GetVersionsPath(pfx *wine.Prefix) string {
	return filepath.Join(pfx.Dir, Versions_Relative)
}

func Mkdirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}
