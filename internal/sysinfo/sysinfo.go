// Package sysinfo provides basic information about the running host machine.
package sysinfo

import (
	"debug/elf"
	"io"
	"os"
)

var (
	Cards   []Card
	Flatpak bool
	LibC    string
)

func init() {
	Cards = getCards()

	_, err := os.Stat("/.flatpak-info")
	Flatpak = err == nil

	f, _ := elf.Open("/proc/self/exe")
	for _, prog := range f.Progs {
		if prog.Type != elf.PT_INTERP {
			continue
		}
		b, _ := io.ReadAll(prog.Open())
		LibC = string(b)
	}
}
