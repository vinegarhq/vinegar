// Package config implements types and routines to configure Vinegar.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sewnie/rbxbin"
	"github.com/sewnie/wine"
	"github.com/sewnie/wine/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
)

// Backwards compatibility to allow:
// 'dxvk = true' and move to 'dxvk = [version]'
type DxvkVersion string

func (v *DxvkVersion) UnmarshalTOML(data interface{}) error {
	switch d := data.(type) {
	case bool:
		*v = ""
		if d {
			*v = "2.7"
		}
	case string:
		*v = DxvkVersion(d)
	default:
		return fmt.Errorf("unsupported type: %T", d)
	}
	return nil
}

func (v DxvkVersion) String() string {
	return string(v)
}

type Studio struct {
	DiscordRPC bool `toml:"discord_rpc" group:"Behavior" row:"Display your development status on your Discord profile"`
	GameMode   bool `toml:"gamemode" group:"Behavior" row:"Apply system optimizations. May improve performance."`

	DXVK      DxvkVersion `toml:"dxvk" group:"Rendering" row:"Improve D3D11 compatibility by translating it to Vulkan,entry,Version,2.7"`
	Renderer  string      `toml:"renderer" group:"Rendering" row:"Studio's Graphics Mode,vals,D3D11,D3D11FL10,Vulkan,OpenGL"` // Enum reflection is impossible
	ForcedGpu string      `toml:"gpu" group:"Rendering" row:"Named or Indexed GPU (ex. integrated or 0)"`

	WineRoot string `toml:"wineroot" group:"Custom Wine" row:"Installation Directory,path"`
	Webview  string `toml:"webview" group:"Custom Wine" row:"For web pages — disable if nonfunctional,entry,Version,109.0.1518.140"`
	Launcher string `toml:"launcher" group:"Custom Wine" row:"Launcher Command (ex. gamescope)"`

	Env    map[string]string `toml:"env" group:"Environment"`
	FFlags rbxbin.FFlags     `toml:"fflags" group:"Fast Flags"`

	ForcedVersion string `toml:"forced_version" group:"Deployment Overrides" row:"Studio Deployment Version"`
	Channel       string `toml:"channel" group:"Deployment Overrides" row:"Studio Update Channel"`
}

type Config struct {
	Debug  bool   `toml:"debug" group:"hidden"`
	Studio Studio `toml:"studio"`
	// Only adds to Studio.Env, reserved for backwards compatibility
	Env map[string]string `toml:"env" group:"hidden"`
}

var (
	ErrWineRootAbs     = errors.New("wine root path is not an absolute path")
	ErrWineRootInvalid = errors.New("no wine binary present in wine root")
)

// Load will load the configuration file; if it doesn't exist, it
// will fallback to the default configuration.
func Load() (*Config, error) {
	cfg := Default()

	if _, err := os.Stat(dirs.ConfigPath); errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(dirs.ConfigPath, &cfg); err != nil {
		return cfg, err
	}

	maps.Copy(cfg.Studio.Env, cfg.Env)
	cfg.Env = nil

	logging.LoggerLevel = slog.LevelInfo
	if cfg.Debug {
		logging.LoggerLevel = slog.LevelDebug
	}

	return cfg, nil
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Debug: false,

		Env: make(map[string]string),

		Studio: Studio{
			DXVK:       "",
			Webview:    "109.0.1518.140", // Last known win7
			GameMode:   true,
			ForcedGpu:  "prime-discrete",
			Renderer:   "Vulkan",
			Channel:    "LIVE",
			DiscordRPC: true,
			FFlags:     make(rbxbin.FFlags),
			Env: map[string]string{
				"WINEESYNC": "1",
			},
		},
	}
}

func (s *Studio) LauncherPath() (string, error) {
	return exec.LookPath(strings.Fields(s.Launcher)[0])
}

func (c *Config) Prefix() (*wine.Prefix, error) {
	pfx := wine.New(
		path.Join(dirs.Prefixes, "studio"),
		c.Studio.WineRoot,
	)

	env := maps.Clone(c.Studio.Env)

	card, err := c.Studio.card()
	if err != nil {
		return nil, err
	}
	if card != nil {
		env["MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE"] = "1"
		env["DRI_PRIME"] = "pci-" + strings.NewReplacer(":", "_", ".", "_").
			Replace(path.Base(card.Device))

		env["__GLX_VENDOR_LIBRARY_NAME"] = "mesa"
		if strings.HasPrefix(card.Driver, "nvidia") {
			env["__GLX_VENDOR_LIBRARY_NAME"] = "nvidia"
		}
	}

	env["WINEDEBUG"] += ",warn+debugstr" // required to read Roblox logs
	env["WINEDLLOVERRIDES"] += ";" + "dxdiagn,winemenubuilder.exe,mscoree,mshtml="
	if !c.Debug {
		env["WINEDEBUG"] += ",fixme-all,err-kerberos,err-ntlm"
	}

	if c.Studio.DXVK != "" {
		if !c.Debug {
			env["DXVK_LOG_LEVEL"] = "warn"
		}
		env["DXVK_LOG_PATH"] = "none"
	}
	env["VK_LOADER_LAYERS_ENABLE"] = "VK_LAYER_VINEGAR_VinegarLayer"
	env["MESA_GL_VERSION_OVERRIDE"] = "4.4"
	env["__GL_THREADED_OPTIMIZATIONS"] = "1"

	for k, v := range env {
		pfx.Env = append(pfx.Env, k+"="+v)
	}

	dxvk.EnvOverride(pfx, c.Studio.DXVK != "")

	slog.Debug("Using Prefix environment", "env", pfx.Env)

	if c.Studio.WineRoot != "" {
		w := pfx.Wine("")
		if w.Err != nil {
			return nil, w.Err
		}
	}

	return pfx, nil
}
