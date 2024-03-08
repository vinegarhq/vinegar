package main

import (
	"fmt"
	"log"
	"path"
	"runtime/debug"

	"github.com/apprehensions/rbxweb/clientsettings"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/sysinfo"
	"github.com/vinegarhq/vinegar/wine"
)

func PrintSysinfo(cfg *config.Config) {
	playerPfx, err := wine.New(BinaryPrefixDir(clientsettings.WindowsPlayer), cfg.Player.WineRoot)
	if err != nil {
		log.Fatalf("player prefix: %s", err)
	}

	studioPfx, err := wine.New(BinaryPrefixDir(clientsettings.WindowsStudio64), cfg.Studio.WineRoot)
	if err != nil {
		log.Fatalf("studio prefix: %s", err)
	}

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
