package gpu

import (
	"log"
	"strconv"
	"strings"
)

type GPU struct {
	path   string
	eDP    bool
	index  int
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
	setIfUndefined("DRI_PRIME", strconv.Itoa(gpu.index))

	if gpu.IsUsingNvidiaDriver() {
		setIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "nvidia")
	} else {
		setIfUndefined("__GLX_VENDOR_LIBRARY_NAME", "mesa")
	}

	log.Printf("Chose card %s (%s).", gpu.path, strconv.Itoa(gpu.index))
	return env
}

func (g *GPU) IsUsingNvidiaDriver() bool {
	return strings.HasSuffix(g.driver, "nvidia")
}
