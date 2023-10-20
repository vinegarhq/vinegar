package sysinfo

import (
	"os"
	"strings"
	"path"
	"path/filepath"
)

// ANY error here will be ignored. All of the filepath querying
// and such is done within /sys/, which is stored in memory.

type Card struct {
	Path     string
	Driver   string
	Index    int
	Embedded bool
}

const drmPath = "/sys/class/drm"
var embeddedDisplays = []string{"eDP", "LVDS", "DP-2"}

func Cards() (cards []Card) {
	drmCards, _ := filepath.Glob(path.Join(drmPath, "card[0-9]"))

	for i, c := range drmCards {
		d, _ := filepath.EvalSymlinks(path.Join(c, "device/driver"))
		d = path.Base(d)

		cards = append(cards, Card{
			Path:     c,
			Driver:   d,
			Index:    i,
			Embedded: embedded(c),
		})
	}

	return
}

func embedded(cardPath string) (embed bool) {
	filepath.Walk(drmPath, func(p string, f os.FileInfo, err error) error {
		if !strings.HasPrefix(p, cardPath) {
			return nil
		}

		for _, hwd := range embeddedDisplays {
			if strings.Contains(p, hwd) {
				embed = true
			}
		}

		return nil
	})

	return
}
