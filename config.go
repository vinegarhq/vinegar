package vinegar

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Thank you ayn2op. Thank you so much.

// Primary struct keeping track of vinegar's directories.
type Directories struct {
	Home   string
	Cache  string
	Config string
	Data   string
	Pfx    string
	Log    string
}

type Config struct {
	Env map[string]string `yaml:"env"`
}

var Dirs = DefDirs()

func DefDirs() Directories {
	homeDir, err := os.UserHomeDir()
	Errc(err)

	dirs := Directories{
		Home:   homeDir,
		Cache:  filepath.Join(homeDir, ".cache", "vinegar"),
		Config: filepath.Join(homeDir, ".config", "vinegar"),
		Data:   filepath.Join(homeDir, ".local", "share", "vinegar"),
	}

	dirs.Pfx = filepath.Join(dirs.Data, "pfx")
	dirs.Log = filepath.Join(dirs.Cache, "logs")

	return dirs
}

// Initialize the Dirs struct and Load the configuration file,
// mainly for environmental variables.
func LoadConfig() {
	data := Config{
		Env: make(map[string]string),
	}

	// Main environment variables initialization
	// Note: these can be overrided by the user.
	data.Env = map[string]string{
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

	configFile, err := ioutil.ReadFile(filepath.Join(Dirs.Config, "config.yaml"))

	// We don't particularly care about if the configuration exists or not,
	// as we are already setting default values.
	if err == nil {
		log.Println("Loading config.yaml")
		err = yaml.Unmarshal(configFile, &data)
		Errc(err)
	}

	for name, value := range data.Env {
		os.Setenv(name, value)
	}
}
