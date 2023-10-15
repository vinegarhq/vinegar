package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/vinegarhq/vinegar/internal/config"
)

var knownVendors = map[string]string{
	//Intel
	"8086": "mesa",
	//AMD
	"1002": "mesa",
	//NVIDIA
	"10de": "nvidia-or-mesa",
	//Fallback in case vendor is unknown
	"default": "mesa",
}

type Card struct {
	path string
	eDP  bool
	id   string
}

//Note: sysfs is located entirely in memory, and as a result does not have IO errors.
//As a result, no error handling when calling os IO operations is done.

func ChooseCard(bcfg config.Binary, c Card) config.Binary {
	vid := strings.Split(c.id, ":")[0]

	vendor := knownVendors[vid]
	if vendor == "" {
		vendor = knownVendors["default"]
	}

	bcfg.Env["MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE"] = "1"
	bcfg.Env["DRI_PRIME"] = c.id

	switch vendor {
	case "mesa":
		bcfg.Env["__GLX_VENDOR_LIBRARY_NAME"] = "mesa"
	case "nvidia-or-mesa":
		driverPath, _ := filepath.EvalSymlinks(filepath.Join(c.path, "device/driver"))

		//Nvidia proprietary driver is being used
		if strings.HasSuffix(driverPath, "nvidia") {
			bcfg.Env["__GLX_VENDOR_LIBRARY_NAME"] = "nvidia"
		} else { //Nouveau is being used
			bcfg.Env["__GLX_VENDOR_LIBRARY_NAME"] = "mesa"
		}
		bcfg.Env.Setenv()
	}

	return bcfg
}

func SetupPrimeOffload(bcfg config.Binary) config.Binary {
	//Sanitize gpu ID
	bcfg.ForcedGpuId = strings.ReplaceAll(strings.ToLower(bcfg.ForcedGpuId), "0x", "")

	cardPattern := regexp.MustCompile("card([0-9]+)$")
	eDP := regexp.MustCompile("card([0-9]+)-eDP-[0-9]+$")
	drmPath := "/sys/class/drm"

	var cards = make([]*Card, 0)
	idDict := make(map[string]*Card, 100)

	dirEntries, _ := os.ReadDir(drmPath)

	for _, v := range dirEntries {
		name := v.Name()
		submatch := cardPattern.FindStringSubmatch(name)
		eDPSubmatch := eDP.FindStringSubmatch(name)

		if submatch != nil {
			i, _ := strconv.Atoi(submatch[1])

			cardPath := path.Join(drmPath, name)

			card := new(Card)
			cards = append(cards, card)

			cards[i].path = cardPath
			vid, _ := os.ReadFile(path.Join(cardPath, "device/vendor"))
			did, _ := os.ReadFile(path.Join(cardPath, "device/device"))

			vidCut, _ := strings.CutPrefix(string(vid), "0x")
			didCut, _ := strings.CutPrefix(string(did), "0x")

			id := strings.ReplaceAll(strings.ToLower(vidCut+":"+didCut), "\n", "")
			cards[i].id = id
			idDict[id] = cards[i]

		} else if eDPSubmatch != nil {
			i, _ := strconv.Atoi(eDPSubmatch[0])
			cards[i].eDP = true
		}
	}

	//Handle cases where the user explictly chooses a gpu to use
	if bcfg.ForcedGpuId != "" {
		if strings.Contains(bcfg.ForcedGpuId, ":") { //This is a gpu ID
			card := idDict[bcfg.ForcedGpuId]

			if card == nil {
				log.Printf("ForcedGpuId is not a valid index or ID. Aborting.")
				os.Exit(1)
			}

			return ChooseCard(bcfg, *card)
		} else { // This is an index
			id, err := strconv.Atoi(bcfg.ForcedGpuId)
			if err != nil {
				log.Printf("ForcedGpuId is not a valid index or ID. Aborting.")
				os.Exit(1)
			}

			card := cards[id]

			if card.path != "" {
				log.Printf("index %d of ForcedGpuId does not exist. Aborting.", id)
				os.Exit(1)
			}

			return ChooseCard(bcfg, *card)
		}

	}

	//There's no need to touch single-gpu systems.
	if len(cards) <= 1 {
		log.Printf("Number of cards is equal or below 1. Skipping prime logic.")
		return bcfg
	}
	//card0 is always an igpu if it exists. If it has no eDP, then Vinegar isn't running on a laptop.
	//As a result, prime doesn't exist and should be skipped.
	if !cards[0].eDP {
		log.Printf("card0 has no eDP. This machine is not a laptop. Skipping prime logic.")
		return bcfg
	}
	if len(cards) > 2 {
		log.Printf("System has %d cards. Skipping prime logic.", len(cards))
		return bcfg
	}

	if bcfg.PrimeOffload {
		return ChooseCard(bcfg, *cards[1])
	} else {
		return ChooseCard(bcfg, *cards[0])
	}
}
