package config

import (
	"fmt"
	"log"

	"github.com/vinegarhq/vinegar/sysinfo"
)

type primeType int

const (
	primeIntegrated primeType = iota
	primeDiscrete
	primeNone
	primeUnknown
)

func (pt primeType) String() string {
	switch pt {
	case primeIntegrated:
		return "integrated"
	case primeDiscrete:
		return "prime-discrete"
	case primeNone:
		return "none"
	default:
		return "unknown"
	}
}

func getPrimeType(s string) primeType {
	if s == "" {
		s = primeNone.String()
	}

	switch s {
	case primeIntegrated.String():
		return primeIntegrated
	case primeDiscrete.String():
		return primeDiscrete
	case primeNone.String():
		return primeNone
	default:
		return primeUnknown
	}
}

// Check if the system actually has PRIME offload and there's no ambiguity with the GPUs.
func primeIsAllowed(isVulkan bool) (bool, error) {
	//There's no ambiguity when there's only one card.
	if len(sysinfo.Cards) <= 1 {
		log.Printf("Number of cards is equal or below 1. Skipping prime logic.")
		return false, nil
	}
	//Handle exotic systems with three or more GPUs (Usually laptops with an epu connnected or workstation desktops)
	if len(sysinfo.Cards) > 2 {
		//OpenGL cannot choose the right card properly. Prompt user the define it themselves
		if !isVulkan {
			return false, fmt.Errorf("opengl cannot select the right gpu. gpus detected: %d", len(sysinfo.Cards))
		} else { //Vulkan knows better than us. Let it do its thing.
			log.Printf("System has %d cards. Skipping prime logic and leaving card selection up to Vulkan.", len(sysinfo.Cards))
		}
		return false, nil
	}
	//card0 is always an igpu if it exists. If it has no eDP, then Vinegar isn't running on a laptop.
	//As a result, prime doesn't exist and should be skipped.
	if !sysinfo.Cards[0].Embedded {
		log.Printf("card0 has no eDP. This machine is not a laptop. Skipping prime logic.")
		return false, nil
	}
	return true, nil
}
