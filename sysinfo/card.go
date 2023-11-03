//go:build linux

package sysinfo

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Card struct {
	Path     string
	Driver   string
	Embedded bool
}

// Determines if a Card is 'embedded' or not, by checking
// if one of these displays belong to the card.
var embeddedDisplays = []string{"eDP", "LVDS"}

func getCards() (cs []card) {
	const drmPath = "/sys/class/drm"

	drmCards, _ := filepath.Glob(path.Join(drmPath, "card[0-9]"))

	for _, c := range drmCards {
		d, _ := filepath.EvalSymlinks(path.Join(c, "device/driver"))
		d = path.Base(d)

		cs = append(cs, card{
			Path:     c,
			Driver:   d,
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
