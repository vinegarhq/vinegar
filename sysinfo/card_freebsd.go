//go:build freebsd
package sysinfo

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"fmt"
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
			Embedded: isEmbedded(c),
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

func isEmbedded(cardPath string) bool {
	var isEmbed bool
	filepath.Walk(drmPath, func(p string, f os.FileInfo, err error) error {
		if !strings.HasPrefix(p, cardPath) {
			return nil
		}
		
		for _, hwd := range embeddedDisplays {
			if strings.Contains(p, hwd) {
				isEmbed = true
			}
		}
		return nil
	})
	return isEmbed
}
