//go:build linux

package sysinfo

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

type card struct {
	Path     string
	Driver   string
	Index    int
	Embedded bool
}

const drmPath = "/sys/class/drm"

var embeddedDisplays = []string{"eDP", "LVDS"}

func getCards() (cs []card) {
	drmCards, _ := filepath.Glob(path.Join(drmPath, "card[0-9]"))

	for i, c := range drmCards {
		d, _ := filepath.EvalSymlinks(path.Join(c, "device/driver"))
		d = path.Base(d)

		cs = append(cs, card{
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
