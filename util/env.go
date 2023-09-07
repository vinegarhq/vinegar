package util

import (
	"log"
	"os"
	"strings"
)

var allowedEnv = []string{
	"PATH",
	"HOME", "USER", "LOGNAME",
	"TZ",
	"LANG", "LC_ALL",
	"EDITOR",
	"XDG_CACHE_HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_DATA_DIRS",
	"XDG_RUNTIME_DIR", // Required by Wayland and Pipewire
	"PULSE_SERVER", "PULSE_CLIENTCONFIG",
	"DISPLAY", "WAYLAND_DISPLAY", "XAUTHORITY",
	"WINEDLLPATH",
	"SDL_GAMECONTROLLERCONFIG",
	"__EGL_EXTERNAL_PLATFORM_CONFIG_DIRS", // Flatpak
}

func SanitizeEnv() {
	log.Println("Sanitizing environment")

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)

		if len(parts) != 2 {
			continue
		}

		allowed := false

		for _, aenv := range allowedEnv {
			if aenv == parts[0] {
				allowed = true
			}
		}

		if !allowed {
			os.Unsetenv(parts[0])
		}
	}
}
