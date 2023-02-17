// Copyright vinegar-development 2023

package main

import (
	"log"
)

// Kill the wineprefix, required for autokill, and
// sometimes fixes Flatpak wine crashes.
func PfxKill() {
	log.Println("Killing wineprefix")
	Exec("wineserver", false, "-k")
}
