package config

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/vinegarhq/vinegar/sysinfo"
)

func (b *Binary) pickCard() error {
	if b.ForcedGpu == "" {
		return nil
	}

	n := len(sysinfo.Cards)
	idx := -1
	prime := false
	aliases := map[string]int{
		"integrated":     0,
		"prime-discrete": 1,
	}

	if i, ok := aliases[b.ForcedGpu]; ok {
		idx = i
		prime = true
	} else {
		i, err := strconv.Atoi(b.ForcedGpu)
		if err != nil {
			return err
		}

		idx = i
	}

	// Check if the system actually has PRIME offload and there's no ambiguity with the GPUs.
	if prime {
		vk := b.Dxvk || b.Renderer == "Vulkan"

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

	if idx < 0 {
		return errors.New("gpu index cannot be negative")
	}

	if n < idx+1 {
		return errors.New("gpu not found")
	}

	b.Env.Set("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	b.Env.Set("DRI_PRIME", strconv.Itoa(idx))

	if sysinfo.Cards[idx].Driver == "nvidia" { // Workaround for OpenGL in nvidia GPUs
		b.Env.Set("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		b.Env.Set("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	return nil
}
