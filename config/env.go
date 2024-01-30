package config

import (
	"os"
	"strings"
)

// Environment is a map representation of a operating environment
// with it's variables.
type Environment map[string]string

// Set will only set the given environment key and value
// if it isn't already set within Environment.
func (e Environment) Set(key, value string) {
	if _, ok := e[key]; ok {
		return
	}
	e[key] = value
}

// Setenv will apply the environment's variables onto the
// global environment using os.Setenv.
func (e Environment) Setenv() {
	for name, value := range e {
		os.Setenv(name, value)
	}
}

var AllowedEnv = []string{
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

// SanitizeEnv modifies the global environment by removing
// all environment variables that are not present in [AllowedEnv].
func SanitizeEnv() {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)

		if len(parts) != 2 {
			continue
		}

		allowed := false

		for _, aenv := range AllowedEnv {
			if aenv == parts[0] {
				allowed = true
			}
		}

		if !allowed {
			os.Unsetenv(parts[0])
		}
	}
}
