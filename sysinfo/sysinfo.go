//go:build amd64

package sysinfo

import (
	"os"
)

var (
	Kernel    string
	Cards     []Card
	Distro    string
	InFlatpak bool
)

func init() {
	Kernel = getKernel()
	Cards = getCards()
	Distro = getDistro()

	_, err := os.Stat("/.flatpak-info")
	InFlatpak = err == nil
}
