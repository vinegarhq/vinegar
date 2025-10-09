// Package config implements types and routines to configure Vinegar.
package config

import (
	"errors"
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

type Studio struct {
	GameMode bool `toml:"gamemode" group:"Behavior" row:"Apply system optimizations. May improve performance."`

	ForcedGpu   string `toml:"gpu" group:"Rendering" row:"Named or Indexed GPU (ex. integrated or 0)"`
	DXVK        bool   `toml:"dxvk" group:"Rendering" row:"Improve D3D11 compatibility by translating it to Vulkan"`
	DXVKVersion string `toml:"dxvk_version" group:"Rendering" row:"DXVK Version"`
	Renderer    string `toml:"renderer" group:"Rendering" row:"Studio's Graphics Mode,vals,D3D11,D3D11FL10,Vulkan,OpenGL"` // Enum reflection is impossible

	WineRoot string `toml:"wineroot" group:"Custom Wine" row:"Installation Directory,path"`
	WebView  string `toml:"webview" group:"Custom Wine" row:"WebView2 Runtime Version"`
	Launcher string `toml:"launcher" group:"Custom Wine" row:"Launcher Command"`

	ForcedVersion string `toml:"forced_version" group:"Deployment Overrides" row:"Studio Deployment Version"`
	Channel       string `toml:"channel" group:"Deployment Overrides" row:"Studio Update Channel"`

	DiscordRPC bool              `toml:"discord_rpc" group:"Behavior" row:"Display your development status on your Discord profile"`
	Env        map[string]string `toml:"env" group:"hidden"`
	FFlags     rbxbin.FFlags     `toml:"fflags" group:"Studio Fast Flags"`
}

type Config struct {
	Debug  bool              `toml:"debug" group:"Behavior" row:"Enable full Wine logging and Log Roblox API requests"`
	Studio Studio            `toml:"studio"`
	Env    map[string]string `toml:"env" group:"Environment"`
}

var (
	ErrWineRootAbs     = errors.New("wine root path is not an absolute path")
	ErrWineRootInvalid = errors.New("no wine binary present in wine root")
)

// Load will load the configuration file; if it doesn't exist, it
// will fallback to the default configuration.
func Load() (*Config, error) {
	d := Default()

	if _, err := os.Stat(dirs.ConfigPath); errors.Is(err, os.ErrNotExist) {
		return d, nil
	}

	if _, err := toml.DecodeFile(dirs.ConfigPath, &d); err != nil {
		return d, err
	}

	logging.LoggerLevel = slog.LevelInfo
	if d.Debug {
		logging.LoggerLevel = slog.LevelDebug
	}

	return d, nil
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Debug: false,

		Env: map[string]string{
			"WINEESYNC": "1",
		},

		Studio: Studio{
			DXVK:        false,
			DXVKVersion: "2.7",
			WebView:     "109.0.1518.140", // Last known win7
			GameMode:    true,
			ForcedGpu:   "prime-discrete",
			Renderer:    "Vulkan",
			Channel:     "LIVE",
			DiscordRPC:  true,
			FFlags:      make(rbxbin.FFlags),
			Env:         make(map[string]string),
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

	if pfx.IsProton() {
		// https://github.com/bottlesdevs/Bottles/issues/3485
		c.Studio.DXVK = true
	}

	env := maps.Clone(c.Env)
	maps.Copy(env, c.Studio.Env)

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
	env["WINEDLLOVERRIDES"] += ";" + dxvk.EnvOverride(c.Studio.DXVK) +
		";" + "dxdiagn,winemenubuilder.exe,mscoree,mshtml="
	if !c.Debug {
		env["WINEDEBUG"] += ",fixme-all,err-kerberos,err-ntlm"
	}

	if c.Studio.DXVK {
		env["DXVK_LOG_LEVEL"] = "warn"
		env["DXVK_LOG_PATH"] = "none"
	}
	env["VK_LOADER_LAYERS_ENABLE"] = "VK_LAYER_VINEGAR_VinegarLayer"
	env["MESA_GL_VERSION_OVERRIDE"] = "4.4"
	env["__GL_THREADED_OPTIMIZATIONS"] = "1"

	for k, v := range env {
		pfx.Env = append(pfx.Env, k+"="+v)
	}

	if c.Studio.WineRoot != "" {
		w := pfx.Wine("")
		if w.Err != nil {
			return nil, errors.New("invalid wineroot")
		}
	}

	return pfx, nil
}
