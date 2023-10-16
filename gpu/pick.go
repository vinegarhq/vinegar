package gpu

import (
	"strconv"

	"github.com/vinegarhq/vinegar/gpu/target"
)

func HandleGpu(target target.TargetGpu, env map[string]string, isVulkan bool) (map[string]string, error) {
	if target.Id == "" {
		return env, nil
	}

	gpus, keyId := GetSystemGPUs()

	if target.Prime {
		allowed, err := PrimeIsAllowed(gpus, isVulkan)
		if err != nil {
			return env, err
		}
		if !allowed {
			return env, nil
		}
	}

	var gpu *GPU

	if target.IsIndex {
		i, err := strconv.Atoi(target.Id)
		if err != nil {
			return env, err
		}

		gpu = gpus[i]
	} else {
		gpu = keyId[target.Id]
	}

	env = gpu.ApplyEnv(env)

	return env, nil
}
