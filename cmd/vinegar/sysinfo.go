package main

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/vinegarhq/vinegar/sysinfo"
)

func (ui *ui) DebugInfo() string {
	var revision string
	bi, _ := debug.ReadBuildInfo()
	for _, bs := range bi.Settings {
		if bs.Key == "vcs.revision" {
			revision = fmt.Sprintf("(%s)", bs.Value)
		}
	}

	var b strings.Builder

	inst := "source"
	if sysinfo.InFlatpak {
		inst = "flatpak"
	}

	info := `* Vinegar: %s %s
* Distro: %s
* Processor: %s
* Kernel: %s
* Wine: %s
* Installation: %s
`

	fmt.Fprintf(&b, info,
		Version, revision,
		sysinfo.Distro,
		sysinfo.CPU.Name,
		sysinfo.Kernel,
		ui.pfx.Version(),
		inst,
	)

	fmt.Fprintln(&b, "* Cards:")
	for i, c := range sysinfo.Cards {
		fmt.Fprintf(&b, "  * Card %d: %s %s %s\n",
			i, c.Driver, filepath.Base(c.Device), c.Path)
	}

	return b.String()
}
