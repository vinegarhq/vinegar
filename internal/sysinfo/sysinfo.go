// Package sysinfo provides basic information about the running host machine.
package sysinfo

import (
	"os"
)

var (
	Cards   []Card
	Flatpak bool
)

func init() {
	Cards = getCards()

	_, err := os.Stat("/.flatpak-info")
	Flatpak = err == nil
}
