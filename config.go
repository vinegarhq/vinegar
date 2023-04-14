package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
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

var (
	ConfigFilePath = filepath.Join(Dirs.Config, "config.toml")
	Config         = loadConfig()
)

func defConfig() Configuration {
	return Configuration{
		RCO:      true,
		AutoKill: true,
		Dxvk:     true,
		Log:      true,
		Prime:    false,
		Launcher: "",
		Renderer: "D3D11",
		Version:  "win10",
		WineRoot: "",
		Env: map[string]string{
			"WINEPREFIX": Dirs.Pfx,
			"WINEARCH":   "win64",
			// "WINEDEBUG":        "fixme-all,-wininet,-ntlm,-winediag,-kerberos",
			"WINEDEBUG":        "-all",
			"WINEESYNC":        "1",
			"WINEDLLOVERRIDES": "dxdiagn=d;winemenubuilder.exe=d;",

			"DXVK_LOG_LEVEL":        "warn",
			"DXVK_LOG_PATH":         "none",
			"DXVK_STATE_CACHE_PATH": filepath.Join(Dirs.Cache, "dxvk"),

			"MESA_GL_VERSION_OVERRIDE":    "4.4",
			"__GL_THREADED_OPTIMIZATIONS": "1",
		},
		FFlags: make(map[string]interface{}),
	}
}

func writeConfigTemplate() {
	CreateDirs(Dirs.Config)

	log.Println("Creating configuration template")

	file, err := os.Create(ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}

	template := "# See how to configure Vinegar on the documentation website:\n" +
		"# https://vinegarhq.github.io/Configuration\n\n"

	if _, err = file.WriteString(template); err != nil {
		log.Fatal(err)
	}
}

func (c *Configuration) Post() {
	if c.Prime {
		c.Env["DRI_PRIME"] = "1"
		c.Env["__NV_PRIME_RENDER_OFFLOAD"] = "1"
		c.Env["__VK_LAYER_NV_optimus"] = "NVIDIA_only"
		c.Env["__GLX_VENDOR_LIBRARY_NAME"] = "nvidia"
	}

	if c.WineRoot != "" {
		log.Println("Using Wine Root")
		os.Setenv("PATH", filepath.Join(c.WineRoot, "bin")+":"+os.Getenv("PATH"))
	}

	for name, value := range c.Env {
		os.Setenv(name, value)
	}
}

func loadConfig() Configuration {
	config := defConfig()

	if _, err := os.Stat(ConfigFilePath); errors.Is(err, os.ErrNotExist) {
		writeConfigTemplate()
	} else if err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile(ConfigFilePath, &config); err != nil {
		log.Fatal("could not parse configuration file:", err)
	}

	config.Post()

	return config
}

func (c *Configuration) Print() {
	if err := toml.NewEncoder(os.Stdout).Encode(c); err != nil {
		log.Fatal(err)
	}
}
