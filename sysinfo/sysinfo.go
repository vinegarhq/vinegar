//go:build amd64

package sysinfo

import (
	"os"

	"golang.org/x/sys/cpu"
)

var (
	Kernel    string
	CPU       string
	Cards     []Card
	Distro    string
	HasAVX    = cpu.X86.HasAVX
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
