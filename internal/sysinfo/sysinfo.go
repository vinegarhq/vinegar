// Package sysinfo provides basic information about the running host machine.
package sysinfo

import (
	"debug/elf"
	"io"
	"log/slog"
	"os"

	"github.com/jaypipes/pcidb"
)

var (
	Cards   []Card
	Flatpak bool
	LibC    string
)

func init() {
	Cards = getCards()

	pci, err := pcidb.New()
	if err == nil {
		for i, c := range Cards {
			if v, ok := pci.Vendors[c.Vendor]; ok {
				Cards[i].Vendor = v.Name
			}
			if p, ok := pci.Products[c.Vendor+c.Product]; ok {
				Cards[i].Product = p.Name
			}
		}
	} else {
		slog.Error("Failed to load PCI Database", "err", err)
	}

	_, err = os.Stat("/.flatpak-info")
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
