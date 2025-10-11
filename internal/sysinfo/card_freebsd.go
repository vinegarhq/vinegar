//go:build freebsd

package sysinfo

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

const drmPath = "/dev/dri"

var embeddedDisplays = []string{"eDP", "LVDS"}

func getCards() (cs []Card) {
	drmCards, err := filepath.Glob(path.Join(drmPath, "card[0-9]*"))
	if err != nil {
		return
	}
	for i, c := range drmCards {
		devPath := path.Join(c, "device")
		driver, err := getDriverFromSysctl(i)
		if err != nil {
			driver = "unknown"
		}
		cs = append(cs, Card{
			Index:    i,
			Path:     c,
			Device:   devPath,
			Driver:   driver,
			Embedded: embedded(c),
		})
	}
	return
}

func getDriverFromSysctl(cardIndex int) (string, error) {
	driverSysctl := fmt.Sprintf("hw.dri.%d.name", cardIndex)
	val, err := syscall.Sysctl(driverSysctl)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
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
