package main

import (
	"fmt"
	"path"
	"runtime/debug"

	"github.com/apprehensions/rbxweb/clientsettings"
	"github.com/apprehensions/wine"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/sysinfo"
)

func PrintSysinfo(cfg *config.Config) {
	playerPfx := wine.New(BinaryPrefixDir(clientsettings.WindowsPlayer), cfg.Player.WineRoot)
	studioPfx := wine.New(BinaryPrefixDir(clientsettings.WindowsStudio64), cfg.Studio.WineRoot)

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
  * Supports AVX: %t
  * Supports split lock detection: %t
* Kernel: %s
* Wine (Player): %s
* Wine (Studio): %s
`

	fmt.Printf(info,
		Version, revision,
		sysinfo.Distro,
		sysinfo.CPU.Name,
		sysinfo.CPU.AVX, sysinfo.CPU.SplitLockDetect,
		sysinfo.Kernel,
		playerPfx.Version(),
		studioPfx.Version(),
	)

	if sysinfo.InFlatpak {
		fmt.Println("* Flatpak: [x]")
	}

	fmt.Println("* Cards:")
	for i, c := range sysinfo.Cards {
		fmt.Printf("  * Card %d: %s %s %s\n", i, c.Driver, path.Base(c.Device), c.Path)
	}
}
