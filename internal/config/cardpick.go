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
	n := len(sysinfo.Cards)

	if n != 2 {
		return false, nil
	}

	if n != 2 && (!vk && n != 1) {
		return false, fmt.Errorf("opengl is not capable of choosing the right gpu, it must be explicitly defined")
	}

	return sysinfo.Cards[0].Embedded, nil
}

func pickCard(opt string, env Environment, isVulkan bool) error {
	if opt == "" {
		return nil
	}

	var cIndex int
	var indexStr string

	usePrime := false

	switch opt {
	//Handle PRIME options
	case "integrated":
		cIndex = 0
		usePrime = true
	case "prime-discrete":
		cIndex = 1
		usePrime = true
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
	if usePrime {
		allowed, err := prime(isVulkan)
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

	env.Set("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	env.Set("DRI_PRIME", indexStr)

	if strings.HasSuffix(c.Driver, "nvidia") { //Workaround for OpenGL in nvidia GPUs
		env.Set("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		env.Set("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	log.Printf("Chose card %s (%s).", c.Path, indexStr)

	env.Setenv()

	return nil
}
