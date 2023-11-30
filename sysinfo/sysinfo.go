//go:build amd64

package sysinfo

import (
	"os"
)

var (
	Kernel    string
	CPU       cpu
	Cards     []Card
	Distro    string
	InFlatpak bool
)

func init() {
	Kernel = getKernel()
	CPU = getCPU()
	Cards = getCards()
	Distro = getDistro()

	_, err := os.Stat("/.flatpak-info")
	InFlatpak = err == nil
}
