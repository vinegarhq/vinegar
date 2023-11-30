//go:build amd64

package sysinfo

import (
	"os"

	"golang.org/x/sys/cpu"
)

type CPUInfo struct {
	Name            string
	AVX             bool
	SplitLockDetect bool
}

var (
	Kernel    string
	CPU       CPUInfo
	Cards     []Card
	Distro    string
	InFlatpak bool
)

func init() {
	Kernel = getKernel()

	CPU.AVX = cpu.X86.HasAVX
	CPU.Name, CPU.SplitLockDetect = cpuModel()

	Cards = getCards()
	Distro = getDistro()

	_, err := os.Stat("/.flatpak-info")
	InFlatpak = err == nil
}
