package config

import (
	"errors"
	"path"
	"strconv"
	"strings"

	"github.com/vinegarhq/vinegar/sysinfo"
)

var (
	ErrOpenGLBlind = errors.New("opengl is not capable of choosing the right gpu, it must be explicitly defined")
	ErrNoCardFound = errors.New("gpu not found")
	ErrBadGpuIndex = errors.New("gpu index cannot be negative")
)

func (s *Studio) pickCard() error {
	if s.ForcedGpu == "" {
		return nil
	}

	n := len(sysinfo.Cards)
	idx := -1
	prime := false
	aliases := map[string]int{
		"integrated":     0,
		"prime-discrete": 1,
	}

	if i, ok := aliases[s.ForcedGpu]; ok {
		idx = i
		prime = true
	} else {
		i, err := strconv.Atoi(s.ForcedGpu)
		if err != nil {
			return err
		}

		idx = i
	}

	if prime {
		vk := s.DXVK || s.Renderer == "Vulkan"

		if n <= 1 {
			return nil
		}

		if n > 2 && !vk {
			return ErrOpenGLBlind
		}

		if !sysinfo.Cards[0].Embedded {
			return nil
		}
	}

	if idx < 0 {
		return ErrBadGpuIndex
	}

	if n < idx+1 {
		return ErrNoCardFound
	}

	c := sysinfo.Cards[idx]

	s.Env.Set("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	s.Env.Set("DRI_PRIME",
		"pci-"+strings.NewReplacer(":", "_", ".", "_").Replace(path.Base(c.Device)),
	)

	if strings.HasPrefix(c.Driver, "nvidia") { // Workaround for OpenGL in nvidia GPUs
		s.Env.Set("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		s.Env.Set("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	return nil
}
