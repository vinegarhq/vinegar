// Package sysinfo provides global top-level variables about
// the running host machine.
package sysinfo

import (
	"os"
)

var (
	Cards     []Card
	InFlatpak bool
)

func init() {
	Cards = getCards()

	_, err := os.Stat("/.flatpak-info")
	InFlatpak = err == nil
}
