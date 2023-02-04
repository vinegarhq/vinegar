// Copyright vinegar-development 2023

package vinegar

import (
	"log"
)

func PfxKill(dirs *Dirs) {
	log.Println("Killing wineprefix")
	Exec(dirs, "wineserver", "-k")
}
