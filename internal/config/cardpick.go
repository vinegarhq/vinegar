package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/vinegarhq/vinegar/sysinfo"
)

// opt accepts the following values:
// aliases - "integrated", "prime-discrete" and "none": Equivalent to "0", "1" or empty. Enables an extra "prime" check.
// integer - GPU index
// empty   - Skips logic and does nothing.
func pickCard(opt string, env Environment, vk bool) error {
	aliases := map[string]string{
		"integrated":     "0",
		"prime-discrete": "1",
		"none":           "",
	}

	var aIdx string

	prime := false

	aIdx = opt
	if a, ok := aliases[opt]; ok {
		aIdx = a
		prime = true
	}

	if aIdx == "" {
		return nil
	}

	n := len(sysinfo.Cards)

	// Check if the system actually has PRIME offload and there's no ambiguity with the GPUs.
	if prime {
		if n != 2 && (!vk && n != 1) {
			return fmt.Errorf("opengl is not capable of choosing the right gpu, it must be explicitly defined")
		}

		if n != 2 {
			return nil
		}

		if !sysinfo.Cards[0].Embedded {
			return nil
		}
	}

	idx, err := strconv.Atoi(opt)
	if err != nil {
		return err
	}

	if idx < 0 {
		return errors.New("gpu index cannot be negative")
	}
	if n < idx+1 {
		return errors.New("gpu not found")
	}
	c := sysinfo.Cards[idx]

	env.Set("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	env.Set("DRI_PRIME", aIdx)

	if strings.HasSuffix(c.Driver, "nvidia") { //Workaround for OpenGL in nvidia GPUs
		env.Set("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		env.Set("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	env.Setenv()

	return nil
}
