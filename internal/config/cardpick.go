package config

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/vinegarhq/vinegar/sysinfo"
)

// Check if the system actually has PRIME offload and there's no ambiguity with the GPUs.
func prime(vk bool) (bool, error) {
	//There's no ambiguity when there's only one card.
	if len(sysinfo.Cards) <= 1 {
		log.Printf("Number of cards is equal or below 1. Skipping prime logic.")
		return false, nil
	}
	//Check if the main card has an embedded display.
	if !sysinfo.Cards[0].Embedded {
		log.Printf("card0 is not embedded. This machine is not a laptop. Skipping prime logic.")
		return false, nil
	}
	//Don't mess with devices that have more than two cards.
	if len(sysinfo.Cards) > 2 {
		//OpenGL cannot choose the right card properly. Prompt user to define it themselves
		if !isVulkan {
			return false, fmt.Errorf("opengl cannot select the right gpu. gpus detected: %d", len(sysinfo.Cards))
		}
		log.Printf("System has %d cards. Skipping prime logic and leaving card selection up to Vulkan.", len(sysinfo.Cards))
		return false, nil
	}
	return true, nil
}

func pickCard(opt string, env Environment, isVulkan bool) error {
	if opt == "" { //Default value for opt
		opt = "none"
	}

	var cIndex int
	var indexStr string

	prime := false

	switch opt {
	//Handle PRIME options
	case "integrated":
		cIndex = 0
		prime = true
	case "prime-discrete":
		cIndex = 1
		prime = true
	//Skip pickCard logic option
	case "none":
		return nil
	//Otherwise, interpret opt as a card index
	default:
		var err error
		cIndex, err = strconv.Atoi(opt)
		if err != nil {
			return err
		}
	}

	if cIndex < 0 {
		return errors.New("card index cannot be negative")
	}

	indexStr = strconv.Itoa(cIndex)

	//PRIME Validation
	if prime {
		allowed, err := primeIsAllowed(isVulkan)
		if err != nil {
			return err
		}
		if !allowed {
			return nil
		}
	}

	if len(sysinfo.Cards) < cIndex+1 {
		return errors.New("gpu not found")
	}

	c := sysinfo.Cards[cIndex]

	env.SetIfUndefined("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	env.SetIfUndefined("DRI_PRIME", indexStr)

	if strings.HasSuffix(c.Driver, "nvidia") { //Workaround for OpenGL in nvidia GPUs
		env.SetIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		env.SetIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	log.Printf("Chose card %s (%s).", c.Path, indexStr)

	env.Setenv()

	return nil
}
