package main

import (
	"fmt"
	"path/filepath"
	"runtime/debug"

	"github.com/apprehensions/wine"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/sysinfo"
)

func printSysinfo(cfg *config.Config) {
	path := filepath.Join(dirs.Prefixes, "studio")
	pfx := wine.New(path, cfg.Studio.WineRoot)

	var revision string
	bi, _ := debug.ReadBuildInfo()
	for _, bs := range bi.Settings {
		if bs.Key == "vcs.revision" {
			revision = fmt.Sprintf("(%s)", bs.Value)
		}
	}

	info := `* Vinegar: %s %s
* Distro: %s
* Processor: %s
* Kernel: %s
* Wine: %s
`

	fmt.Printf(info,
		Version, revision,
		sysinfo.Distro,
		sysinfo.CPU.Name,
		sysinfo.Kernel,
		pfx.Version(),
	)

	if sysinfo.InFlatpak {
		fmt.Println("* Flatpak: [x]")
	}

	fmt.Println("* Cards:")
	for i, c := range sysinfo.Cards {
		fmt.Printf("  * Card %d: %s %s %s\n", i, c.Driver, filepath.Base(c.Device), c.Path)
	}
}
