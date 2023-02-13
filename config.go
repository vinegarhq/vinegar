// Copyright vinegar-development 2023

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml/v2"
)

// Thank you ayn2op. Thank you so much.

// Primary struct keeping track of vinegar's directories.
type Directories struct {
	Cache  string
	Config string
	Data   string
	Pfx    string
	Log    string
}

type Configuration struct {
	Renderer  string
	ApplyRCO  bool
	AutoRFPSU bool
	Env       map[string]string
	FFlags    map[string]any
}

var Dirs = defDirs()
var ConfigFilePath = filepath.Join(Dirs.Config, "config.toml")
var Config = loadConfig()

// Define the default values for the Directories struct globally
// for other functions to use it.
func defDirs() Directories {
	homeDir, err := os.UserHomeDir()
	Errc(err)

	xdgDirs := map[string]string{
		"XDG_CACHE_HOME":  filepath.Join(homeDir, ".cache"),
		"XDG_CONFIG_HOME": filepath.Join(homeDir, ".config"),
		"XDG_DATA_HOME":   filepath.Join(homeDir, ".local", "share"),
	}

	// If the variable has already been set, we
	// should use it instead of our own.
	for varName := range xdgDirs {
		value := os.Getenv(varName)

		if value != "" {
			xdgDirs[varName] = value
		}
	}

	dirs := Directories{
		Cache:  filepath.Join(xdgDirs["XDG_CACHE_HOME"], "vinegar"),
		Config: filepath.Join(xdgDirs["XDG_CONFIG_HOME"], "vinegar"),
		Data:   filepath.Join(xdgDirs["XDG_DATA_HOME"], "vinegar"),
	}

	dirs.Pfx = filepath.Join(dirs.Data, "pfx")
	dirs.Log = filepath.Join(dirs.Cache, "logs")

	return dirs
}

// Initialize the configuration, and load the configuration file (if available)
func defConfig() Configuration {
	var config Configuration

	// Main environment variables initialization
	// Note: these can be overrided by the user.
	config.Env = map[string]string{
		"WINEPREFIX":       Dirs.Pfx,
		"WINEARCH":         "win64", // required for rbxfpsunlocker
		"WINEDEBUG":        "fixme-all,-wininet,-ntlm,-winediag,-kerberos",
		"WINEDLLOVERRIDES": "dxdiagn=d;winemenubuilder.exe=d",

		"DXVK_LOG_LEVEL":        "warn",
		"DXVK_LOG_PATH":         "none",
		"DXVK_STATE_CACHE_PATH": filepath.Join(Dirs.Cache, "dxvk"),

		// should be set by the user.
		"MESA_GL_VERSION_OVERRIDE": "4.4",

		// PRIME, should be automatic.
		"DRI_PRIME":                 "1",
		"__NV_PRIME_RENDER_OFFLOAD": "1",
		"__VK_LAYER_NV_optimus":     "NVIDIA_only",
		"__GLX_VENDOR_LIBRARY_NAME": "nvidia",
	}
	config.FFlags = map[string]any{}

	if runtime.GOOS == "freebsd" {
		config.Env["WINEARCH"] = "win32"
		config.Env["WINE_NO_WOW64"] = "1"
	}

	return config
}

func loadConfig() Configuration {
	defaultConfiguration := defConfig()
	var config Configuration

	configText, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Is ok, generate the file

			err = os.MkdirAll(Dirs.Config, os.ModePerm)
			Errc(err)

			file, err := os.Create(ConfigFilePath)
			fmt.Println(err)
			Errc(err)
			defer file.Close()

			// TODO: Change to point to actual chapter that describes this
			_, err = file.WriteString("# See how to configure Vinegar on the documentation website:\n# https://vinegarhq.github.io")
			Errc(err)
		} else {
			log.Println("Unable to read the configuration file. Will use default configuration presets. File path:", ConfigFilePath)
			log.Println(err.Error())

			config = defaultConfiguration
		}
	} else {
		err = toml.Unmarshal([]byte(configText), &config)
		Errc(err, "Could not parse configuration file.")
	}

	if config.Env == nil {
		config.Env = defaultConfiguration.Env
	} else {
		AddDefaultsToMap(config.Env, defaultConfiguration.Env)
	}

	if config.FFlags == nil {
		config.FFlags = defaultConfiguration.FFlags
	} else {
		AddDefaultsToMap(config.FFlags, defaultConfiguration.FFlags)
	}

	possibleRenderers := []string{
		"OpenGL",
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	for _, rend := range possibleRenderers {
		isRenderer := rend == config.Renderer
		config.FFlags["FFlagDebugGraphicsPrefer"+rend] = isRenderer
		config.FFlags["FFlagDebugGraphicsDisable"+rend] = !isRenderer
	}

	for name, value := range config.Env {
		os.Setenv(name, value)
	}

	return config
}
