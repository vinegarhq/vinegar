package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	// Tagging required for the pretty print to make sense
	RCO      bool                   `toml:"rco"`
	AutoKill bool                   `toml:"autokill"`
	Dxvk     bool                   `toml:"dxvk"`
	Log      bool                   `toml:"log"`
	Prime    bool                   `toml:"prime"`
	Launcher string                 `toml:"launcher"`
	Renderer string                 `toml:"renderer"`
	Version  string                 `toml:"version"`
	WineRoot string                 `toml:"wineroot"`
	Env      map[string]string      `toml:"env"`
	FFlags   map[string]interface{} `toml:"fflags"`
}

// Global Configuration for all of Vinegar's functions to be able to access.
// The configuration will ALWAYS be loaded upon vinegar's launch.
var (
	ConfigFilePath = filepath.Join(Dirs.Config, "config.toml")
	Config         = LoadConfigFile()
)

func LoadConfigFile() Configuration {
	config := DefaultConfig()

	if _, err := os.Stat(ConfigFilePath); errors.Is(err, os.ErrNotExist) {
		log.Println("Using default configuration")

		return config
	}

	if _, err := toml.DecodeFile(ConfigFilePath, &config); err != nil {
		log.Fatal(err)
	}

	if config.Prime {
		config.Env["DRI_PRIME"] = "1" // nouveau
		config.Env["__NV_PRIME_RENDER_OFFLOAD"] = "1"
		config.Env["__VK_LAYER_NV_optimus"] = "NVIDIA_only"
		config.Env["__GLX_VENDOR_LIBRARY_NAME"] = "nvidia"
	}

	if config.WineRoot != "" {
		config.Env["PATH"] = filepath.Join(config.WineRoot, "bin") + ":" + os.Getenv("PATH")
	}

	if config.Dxvk {
		// Tells wine to use the DXVK DLLs
		config.Env["WINEDLLOVERRIDES"] += "d3d10core=n;d3d11=n;d3d9=n;dxgi=n"
	}

	for name, value := range config.Env {
		os.Setenv(name, value)
	}

	return config
}

// Default or 'sane' configuration of Vinegar.
func DefaultConfig() Configuration {
	return Configuration{
		RCO:      true,
		AutoKill: true,
		Dxvk:     true,
		Log:      true,
		Prime:    false,
		Renderer: "D3D11",
		Version:  "win10",
		Env: map[string]string{
			"WINEPREFIX":       Dirs.Prefix,
			"WINEARCH":         "win64",
			"WINEDEBUG":        "-all",
			"WINEESYNC":        "1",
			"WINEDLLOVERRIDES": "dxdiagn=d;winemenubuilder.exe=d;",

			"DXVK_LOG_LEVEL":        "warn",
			"DXVK_LOG_PATH":         "none",
			"DXVK_STATE_CACHE_PATH": filepath.Join(Dirs.Cache, "dxvk"),

			"MESA_GL_VERSION_OVERRIDE":    "4.4",
			"__GL_THREADED_OPTIMIZATIONS": "1",
		},
	}
}

func (c *Configuration) Print() {
	if err := toml.NewEncoder(os.Stdout).Encode(c); err != nil {
		log.Fatal(err)
	}
}
