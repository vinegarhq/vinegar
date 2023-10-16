package gpu

import (
	"fmt"
	"log"
)

// Check if the system actually has PRIME offload and there's no ambiguity with the GPUs.
func PrimeIsAllowed(gpus []*GPU, isVulkan bool) (bool, error) {
	//There's no ambiguity when there's only one card.
	if len(gpus) <= 1 {
		log.Printf("Number of cards is equal or below 1. Skipping prime logic.")
		return false, nil
	}
	//Handle exotic systems with three or more GPUs (Usually laptops with an epu connnected or workstation desktops)
	if len(gpus) > 2 {
		//OpenGL cannot choose the right card properly. Prompt user the define it themselves
		if !isVulkan {
			return false, fmt.Errorf("opengl cannot select the right gpu. gpus detected: %d", len(gpus))
		} else { //Vulkan knows better than us. Let it do its thing.
			log.Printf("System has %d cards. Skipping prime logic and leaving card selection up to Vulkan.", len(gpus))
		}
		return false, nil
	}
	//card0 is always an igpu if it exists. If it has no eDP, then Vinegar isn't running on a laptop.
	//As a result, prime doesn't exist and should be skipped.
	if !gpus[0].eDP {
		log.Printf("card0 has no eDP. This machine is not a laptop. Skipping prime logic.")
		return false, nil
	}
	return true, nil
}
