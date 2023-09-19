package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/util"
)

var Path = filepath.Join(dirs.Config, "config.toml")

type Environment map[string]string

type Application struct {
	Channel        string        `toml:"channel"`
	Renderer       string        `toml:"renderer"`
	ForcedVersion  string        `toml:"forced_version"`
	AutoKillPrefix bool          `toml:"auto_kill_prefix"`
	Dxvk           bool          `toml:"dxvk"`
	FFlags         roblox.FFlags `toml:"fflags"`
	Env            Environment   `toml:"env"`
}

type Config struct {
	Launcher           string      `toml:"launcher"`
	WineRoot           string      `toml:"wineroot"`
	DxvkVersion        string      `toml:"dxvk_version"`
	MultipleInstances  bool        `toml:"multiple_instances"`
	SanitizeEnv        bool        `toml:"sanitize_env"`
	Player             Application `toml:"player"`
	Studio             Application `toml:"studio"`
	Env                Environment `toml:"env"`
	WineHQReportMode   bool        `toml:"winehq_report_mode"`
	WineRootReportMode string      `toml:"wineroot_report_mode"`
}

func Load() (Config, error) {
	cfg := Default()

	if _, err := os.Stat(Path); errors.Is(err, os.ErrNotExist) {
		log.Println("Using default configuration")

		return cfg, nil
	}

	if _, err := toml.DecodeFile(Path, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode configuration file: %w", err)
	}

	if err := cfg.Setup(); err != nil {
		return cfg, fmt.Errorf("failed to setup configuration: %w", err)
	}

	return cfg, nil
}

func Default() Config {
	return Config{
		DxvkVersion:      "2.3",
		WineHQReportMode: false,

		Env: Environment{
			"WINEARCH":         "win64",
			"WINEDEBUG":        "err-kerberos,err-ntlm",
			"WINEESYNC":        "1",
			"WINEDLLOVERRIDES": "dxdiagn=d;winemenubuilder.exe=d",

			"DXVK_LOG_LEVEL": "warn",
			"DXVK_LOG_PATH":  "none",

			"MESA_GL_VERSION_OVERRIDE":    "4.4",
			"__GL_THREADED_OPTIMIZATIONS": "1",
		},

		Player: Application{
			Dxvk:           true,
			AutoKillPrefix: true,
			FFlags: roblox.FFlags{
				"DFIntTaskSchedulerTargetFps": 640,
			},
		},
		Studio: Application{
			Dxvk: true,
		},
	}
}

func (e *Environment) Setenv() {
	for name, value := range *e {
		os.Setenv(name, value)
	}
}

func (c *Config) Setup() error {
	if c.SanitizeEnv {
		util.SanitizeEnv()
	}

	if c.WineHQReportMode {
		log.Printf("WARNING: WineHQReportMode is enabled. This is a development option; do not continue unless you know what you're doing.")

		//Override the WineRoot with the one defined by WineRootReportMode. This root *must* point to a WineHQ (unpatched) wine version, otherwise reports will be invalid.
		//Note for packagers: Feedback is needed here for those which are bundling their own wine build.
		c.WineRoot = c.WineRootReportMode

		if c.WineRoot == "" {
			log.Printf("WineHQReportMode: Wine Root has been overriden; using system's wine.")
		} else {
			log.Printf("WineHQReportMode: Wine Root has been overriden to \"%s\".", c.WineRoot)
		}
	}

	if c.WineRoot != "" {
		bin := filepath.Join(c.WineRoot, "bin")

		if !filepath.IsAbs(c.WineRoot) {
			return errors.New("ensure that the wine root given is an absolute path")
		}

		_, err := os.Stat(filepath.Join(bin, "wine"))
		if err != nil {
			return fmt.Errorf("invalid wine root given: %s", err)
		}

		c.Env["PATH"] = bin + ":" + os.Getenv("PATH")
		os.Unsetenv("WINEDLLPATH")
		log.Printf("Using Wine Root: %s", c.WineRoot)
	}

	if !roblox.ValidRenderer(c.Player.Renderer) || !roblox.ValidRenderer(c.Studio.Renderer) {
		return fmt.Errorf("invalid renderer given to either player or studio")
	}

	c.Env.Setenv()

	return nil
}
