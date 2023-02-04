// Copyright vinegar-development 2023

package vinegar

import (
	"log"
)

// Kill the wineprefix.
func PfxKill() {
	log.Println("Killing wineprefix")
	Exec("wineserver", "-k")
}
