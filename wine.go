// Copyright vinegar-development 2023

package vinegar

import (
	"log"
)

func PfxKill() {
	log.Println("Killing wineprefix")
	Exec("wineserver", "-k")
}
