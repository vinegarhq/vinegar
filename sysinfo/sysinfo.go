// Package sysinfo provides global top-level variables about
// the running host machine.
package sysinfo

import (
	"os"
)

var (
	Kernel    string
	CPU       Processor
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
