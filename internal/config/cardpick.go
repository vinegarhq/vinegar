package config

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/vinegarhq/vinegar/sysinfo"
)

func (s *Studio) card() (*sysinfo.Card, error) {
	if s.ForcedGpu == "" {
		return nil, nil
	}

	n := len(sysinfo.Cards)
	idx := -1
	if i, ok := map[string]int{
		"integrated":     0,
		"prime-discrete": 1,
	}[s.ForcedGpu]; ok {
		idx = i

		vk := s.DXVK || s.Renderer == "Vulkan"

		if n <= 1 {
			return nil, nil
		}

		if n > 2 && !vk {
			return nil, errors.New("gpu must be explicitly defined for opengl")
		}

		if !sysinfo.Cards[0].Embedded {
			return nil, nil
		}
	} else {
		i, err := strconv.Atoi(s.ForcedGpu)
		if err != nil {
			return nil, err
		}

		idx = i
	}

	if idx < 0 {
		return nil, errors.New("gpu index is negative")
	}

	if n < idx+1 {
		return nil, fmt.Errorf("gpu %s not found", s.ForcedGpu)
	}

	return &sysinfo.Cards[idx], nil
}
