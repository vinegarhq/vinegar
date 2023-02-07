// Copyright vinegar-development 2023

package main

import (
	"log"
)

// Kill the wineprefix.
func PfxKill() {
	log.Println("Killing wineprefix")
	Exec("wineserver", true, "-k")
}
