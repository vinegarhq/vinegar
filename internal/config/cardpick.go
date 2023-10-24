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
func (b *Binary) pickCard() error {
	aliases := map[string]string{
		"integrated":     "0",
		"prime-discrete": "1",
		"none":           "",
	}

	var (
		aIdx  string
		prime bool
	)

	aIdx = b.ForcedGpu
	if a, ok := aliases[b.ForcedGpu]; ok {
		aIdx = a
		prime = true
	}

	if aIdx == "" {
		return nil
	}

	n := len(sysinfo.Cards)

	// Check if the system actually has PRIME offload and there's no ambiguity with the GPUs.
	if prime {
		vk := (b.Dxvk && b.Renderer == "D3D11") || b.Renderer == "Vulkan"

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

	idx, err := strconv.Atoi(aIdx)
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

	b.Env.Set("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	b.Env.Set("DRI_PRIME", aIdx)

	if strings.HasSuffix(c.Driver, "nvidia") { //Workaround for OpenGL in nvidia GPUs
		b.Env.Set("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		b.Env.Set("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}
	return nil
}
