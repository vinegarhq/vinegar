package sysinfo

import (
	"os"
)

var (
	Kernel    kernel
	CPU       string
	Cards     []card
	Distro    distro
	InFlatpak bool
)

func init() {
	Kernel = getKernel()
	CPU = cpuModel()
	Cards = getCards()
	Distro = getDistro()

	_, err := os.Stat("/.flatpak-info")
	InFlatpak = err == nil
}
