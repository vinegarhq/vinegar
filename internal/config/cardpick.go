package config

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/vinegarhq/vinegar/internal/sysinfo"
)

func (s *Studio) card() (*sysinfo.Card, error) {
	if s.ForcedGPU == "" {
		return nil, nil
	}

	n := len(sysinfo.Cards)
	idx := -1
	if i, ok := map[string]int{
		"integrated":     0,
		"prime-discrete": 1,
	}[s.ForcedGPU]; ok {
		idx = i

		vk := s.DXVK != "" || s.Renderer == "Vulkan"

		if n <= 1 {
			return nil, nil
		}

		if n > 2 && !vk {
			return nil, errors.New("GPU must be explicitly defined for OpenGL")
		}

		if !sysinfo.Cards[0].Embedded {
			return nil, nil
		}
	} else {
		i, err := strconv.Atoi(s.ForcedGPU)
		if err != nil {
			return nil, err
		}

		idx = i
	}

	if idx < 0 {
		return nil, errors.New("GPU index is negative")
	}

	if n < idx+1 {
		return nil, fmt.Errorf("gpu %s not found", s.ForcedGPU)
	}

	return &sysinfo.Cards[idx], nil
}
