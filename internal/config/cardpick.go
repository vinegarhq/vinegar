package config

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/vinegarhq/vinegar/sysinfo"
)

func pickCard(opt string, env Environment, isVulkan bool) error {
	var i int
	var iStr string

	var err error
	isP := false

	switch getPrimeType(opt) {
	case primeIntegrated:
		i = 0
		isP = true
	case primeDiscrete:
		i = 1
		isP = true
	case primeNone:
		i = -1
	case primeUnknown:
		i, err = strconv.Atoi(opt)
		iStr = opt
	}
	if err != nil {
		return err
	}
	if isP {
		iStr = strconv.Itoa(i)
	}

	//Treat negative values as an "ignore" condition.
	if i < 0 {
		return nil
	}

	//PRIME Validation
	if isP {
		allowed, err := primeIsAllowed(isVulkan)
		if err != nil {
			return err
		}
		if !allowed {
			return nil
		}
	}

	if len(sysinfo.Cards) < i+1 {
		return errors.New("gpu not found")
	}

	c := sysinfo.Cards[i]

	setIfUndefined := func(k string, v string) {
		if _, ok := env[k]; ok {
			log.Printf("Warning: env var %s is already defined. Will not redefine it.", k)
			return
		}
		env[k] = v
	}

	setIfUndefined("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	setIfUndefined("DRI_PRIME", iStr)

	if strings.HasSuffix(c.Driver, "nvidia") { //Workaround for OpenGL in nvidia GPUs
		setIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		setIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	log.Printf("Chose card %s (%s).", c.Path, iStr)

	env.Setenv()

	return nil
}
