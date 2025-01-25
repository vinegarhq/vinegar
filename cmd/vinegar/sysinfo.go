package main

import (
	"fmt"
	"path"
	"runtime/debug"

	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/sysinfo"
)

func PrintSysinfo(cfg *config.Config) {
	b, _ := NewBinary(cfg)

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
		b.Prefix.Version(),
	)

	if sysinfo.InFlatpak {
		fmt.Println("* Flatpak: [x]")
	}

	fmt.Println("* Cards:")
	for i, c := range sysinfo.Cards {
		fmt.Printf("  * Card %d: %s %s %s\n", i, c.Driver, path.Base(c.Device), c.Path)
	}
}
