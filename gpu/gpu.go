package gpu

import (
	"log"
	"strings"

	"github.com/vinegarhq/vinegar/gpu/target"
)

type GPU struct {
	path   string
	eDP    bool
	id     string
	driver string
}

func (gpu *GPU) ApplyEnv(env map[string]string) map[string]string {
	setIfUndefined := func(k string, v string) {
		if _, ok := env[k]; ok {
			log.Printf("Warning: env var %s is already defined. Will not redefine it.", k)
		} else {
			env[k] = v
		}
	}

	setIfUndefined("MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE", "1")
	setIfUndefined("DRI_PRIME", gpu.id)

	if gpu.IsUsingNvidiaDriver() {
		setIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		setIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	log.Printf("Chose card %s (%s).", gpu.path, gpu.id)
	return env
}

func (g *GPU) SetId(vid []byte, nid []byte) {
	g.id = target.SanitizeGpuId(strings.ReplaceAll(string(vid)+":"+string(nid), "\n", ""))
}

func (g *GPU) IsUsingNvidiaDriver() bool {
	return strings.HasSuffix(g.driver, "nvidia")
}
