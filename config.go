package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Channels struct {
	Force  bool   `toml:"force"`
	Player string `toml:"player"`
	Studio string `toml:"studio"`
}

type Configuration struct {
	// Tagging required for the pretty print to make sense
	RCO      bool                   `toml:"rco"`
	AutoKill bool                   `toml:"autokill"`
	Dxvk     bool                   `toml:"dxvk"`
	Log      bool                   `toml:"log"`
	Prime    bool                   `toml:"prime"`
	Launcher string                 `toml:"launcher"`
	Channels Channels               `toml:"channels"`
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
		log.Printf("Using Wine Root: %s", config.WineRoot)

		// I'm not sure if this is how it should be done, but it works *shrug*
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
		// Dont use 'LIVE' as the default channel, as an empty
		// channel, as roblox sets empty channels for live users.
		Channels: Channels{
			Force:  false,
			Player: "",
			Studio: "",
		},
		Env: map[string]string{
			"WINEPREFIX":       Dirs.Prefix,
			"WINEARCH":         "win64",
			"WINEDEBUG":        "-all", // Peformance gain by removing most Wine logging
			"WINEESYNC":        "1",
			"WINEDLLOVERRIDES": "dxdiagn=d;winemenubuilder.exe=d;",

			"DXVK_LOG_LEVEL":        "info",
			"DXVK_LOG_PATH":         "none", // DXVK will leave log files in CWD
			"DXVK_STATE_CACHE_PATH": filepath.Join(Dirs.Cache, "dxvk"),

			"MESA_GL_VERSION_OVERRIDE":    "4.4", // Fixes many 'white screen' issues
			"__GL_THREADED_OPTIMIZATIONS": "1",   // NVIDIA
		},
	}
}

func (c *Configuration) Print() {
	if err := toml.NewEncoder(os.Stdout).Encode(c); err != nil {
		log.Fatal(err)
	}
}
