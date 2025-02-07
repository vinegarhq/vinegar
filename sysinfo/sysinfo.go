// Package sysinfo provides global top-level variables about
// the running host machine.
package sysinfo

import (
	"os"

	"golang.org/x/sys/cpu"
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

	CPU = Processor{
		Name: "unknown cpu",
		AVX: cpu.X86.HasAVX,
	}

	if n := cpuName(); n != "" {
		CPU.Name = n
	}

	Cards = getCards()
	Distro = getDistro()

	_, err := os.Stat("/.flatpak-info")
	InFlatpak = err == nil
}
