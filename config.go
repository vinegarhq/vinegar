// Copyright vinegar-development 2023

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
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
	Renderer        string                 `yaml:"renderer"`
	UseRCOFFlags    bool                   `yaml:"rco"`
	AutoLaunchRFPSU bool                   `yaml:"rfpsu"`
	Env             map[string]string      `yaml:"env"`
	FFlags          map[string]interface{} `yaml:"fflags"`
}

var Dirs       = defDirs()
var ConfigFile = filepath.Join(Dirs.Config, "config.yaml")
var Config     = loadConfig()

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
	for varName, _ := range xdgDirs {
		value := os.Getenv(varName)

		if value != "" {
			xdgDirs[varName] = value
		}
	}

	dirs := Directories{
		Cache:  filepath.Join(xdgDirs["XDG_CACHE_HOME"],  "vinegar"),
		Config: filepath.Join(xdgDirs["XDG_CONFIG_HOME"], "vinegar"),
		Data:   filepath.Join(xdgDirs["XDG_DATA_HOME"],   "vinegar"),
	}

	dirs.Pfx = filepath.Join(dirs.Data, "pfx")
	dirs.Log = filepath.Join(dirs.Cache, "logs")

	return dirs
}

// Initialize the configuration, and load the configuration file (if available)
func defConfig() Configuration {
	config := Configuration{
		Renderer:        "Vulkan",
		Env:             make(map[string]string),
		FFlags:          make(map[string]interface{}),
		UseRCOFFlags:    true,
		AutoLaunchRFPSU: false,
	}

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

	if runtime.GOOS == "freebsd" {
		config.Env["WINEARCH"] = "win32"
		config.Env["WINE_NO_WOW64"] = "1"
	}

	return config
}

func loadConfig() Configuration {
	var config Configuration

	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		log.Println("Configuration not found, appending defaults")

		CheckDirs(Dirs.Config)
		configFile, err := os.Create(ConfigFile)
		Errc(err)
		defer configFile.Close()

		config := defConfig()
		err = yaml.NewEncoder(configFile).Encode(config)
		Errc(err)
	}

	configFile, err := ioutil.ReadFile(ConfigFile)

	if err == nil {
		log.Println("Loading", ConfigFile)
		err = yaml.Unmarshal(configFile, &config)
		Errc(err)
	} else {
		log.Fatal("Failed to load configuration")
	}

	possibleRenderers := []string{
		"OpenGL", 
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	for _, rend := range possibleRenderers {
		if rend == config.Renderer {
			config.FFlags["FFlagDebugGraphicsPrefer"  + rend] = true
			config.FFlags["FFlagDebugGraphicsDisable" + rend] = false
		} else {
			config.FFlags["FFlagDebugGraphicsPrefer" + rend]  = false
			config.FFlags["FFlagDebugGraphicsDisable" + rend] = true
		}
	}

	for name, value := range config.Env {
		os.Setenv(name, value)
	}

	return config
}
