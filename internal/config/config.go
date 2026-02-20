// Package config implements types and routines to configure Vinegar.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sewnie/rbxbin"
	"github.com/sewnie/wine"
	"github.com/sewnie/wine/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/internal/sysinfo"
)

const (
	DXVKVersion      = "2.7.1"
	DXVKSarekVersion = "Sarek-1.11.0-async"
	WebViewVersion   = "144.0.3719.92"

	DesktopsResolution = "1814x1024"
)

// Order must be the same as the renderer model in the configurator.
var RendererValues = []string{
	"D3D11",
	"D3D11FL10",
	"DXVK",
	"DXVK-Sarek",
	"Vulkan",
}

type Studio struct {
	WebView  string `toml:"webview"`
	WineRoot string `toml:"wineroot"`

	Renderer  string `toml:"renderer"`
	Desktop   string `toml:"virtual_desktop"`
	ForcedGpu string `toml:"gpu"`

	Launcher   string `toml:"launcher"`
	DiscordRPC bool   `toml:"discord_rpc"`
	GameMode   bool   `toml:"gamemode"`

	Env    map[string]string `toml:"env"`
	FFlags rbxbin.FFlags     `toml:"fflags"`

	ForcedVersion string `toml:"forced_version"`
	Channel       string `toml:"channel"`
}

type Config struct {
	Studio Studio `toml:"studio"`
	// Only adds to Studio.Env, reserved for backwards compatibility
	Env   map[string]string `toml:"env"`
	Debug bool              `toml:"debug"`
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
func Default() (cfg *Config) {

	cfg = &Config{
		Debug: false,

		Env: make(map[string]string),

		Studio: Studio{
			WebView:    WebViewVersion,
			GameMode:   true,
			Renderer:   "DXVK",
			Channel:    "",
			DiscordRPC: true,
			FFlags:     make(rbxbin.FFlags),
			Env:        make(map[string]string),
		},
	}
	// No need to select if there is only a single GPU, and to
	// prefer PRIME discrete behavior by default, incase the first
	// GPU is not integrated.
	if len(sysinfo.Cards) >= 2 && !sysinfo.Cards[0].Embedded {
		cfg.Studio.ForcedGpu = sysinfo.Cards[1].String()
	}

	// Prefer to use the VinegarHQ Kombucha builds to be
	// downloaded at runtime, on non musl systems.
	// Note: Author of this code uses a musl system. (me)
	if !strings.Contains(sysinfo.LibC, "musl") {
		cfg.Studio.WineRoot = dirs.WinePath
	}

	return
}

func (s *Studio) UnmarshalTOML(data any) error {
	// prevent recursion by typing
	type Alias Studio
	proxy := struct {
		*Alias
		DXVK string `toml:"dxvk"`
	}{Alias: (*Alias)(s)}

	// encode to and back to retrieve all options from raw TOML
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}
	if _, err := toml.Decode(buf.String(), &proxy); err != nil {
		return err
	}

	s = (*Studio)(proxy.Alias)

	if proxy.DXVK != "" {
		slog.Warn("The DXVK option alongside it's versioning has been deprecated, setting Renderer")
		s.Renderer = "DXVK"
	}
	if !slices.Contains(RendererValues, s.Renderer) {
		return fmt.Errorf("renderer must be one of %s", RendererValues)
	}
	return nil
}

func (s *Studio) LauncherPath() (string, error) {
	return exec.LookPath(strings.Fields(s.Launcher)[0])
}

func (s *Studio) DXVKVersion() string {
	switch s.Renderer {
	case "DXVK":
		return DXVKVersion
	case "DXVK-Sarek":
		return DXVKSarekVersion
	}
	return ""
}

func (c *Config) Prefix() *wine.Prefix {
	pfx := wine.New(
		path.Join(dirs.Prefixes, "studio"),
		string(c.Studio.WineRoot),
	)

	env := maps.Clone(c.Studio.Env)

	if len(sysinfo.Cards) > 1 {
		env["MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE"] = "1"
	}

	for _, card := range sysinfo.Cards {
		if string(c.Studio.ForcedGpu) != card.String() {
			continue
		}

		slog.Debug("Using GPU", "index", card.Index, "card", card.Product)
		env["MESA_VK_DEVICE_SELECT_FORCE_DEFAULT_DEVICE"] = "1"
		env["DRI_PRIME"] = "pci-" + strings.NewReplacer(":", "_", ".", "_").
			Replace(path.Base(card.Device))

		env["__GLX_VENDOR_LIBRARY_NAME"] = "mesa"
		if strings.HasPrefix(card.Driver, "nvidia") {
			env["__GLX_VENDOR_LIBRARY_NAME"] = "nvidia"
		}
	}

	env["WINEDEBUG"] += ",warn+debugstr" // required to read Roblox logs
	env["XR_LOADER_DEBUG"] = "none"      // already shown in Roblox log
	env["WINEDLLOVERRIDES"] += ";" + "dxdiagn,winemenubuilder.exe,mscoree,mshtml="
	if !c.Debug {
		env["WINEDEBUG"] += ",fixme-all,err-kerberos,err-ntlm,err-combase"
	}

	env["WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"] = "--disable-gpu-compositing "

	switch c.Studio.Renderer {
	case "D3D11", "D3D11FL10", "OpenGL":
		env["WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"] += "--use-angle=gl"
	case "DXVK", "DXVK-Sarek", "Vulkan":
		env["WINE_D3D_CONFIG"] = "renderer=vulkan"
		env["WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"] += "--use-angle=vulkan"
	}

	useDXVK := c.Studio.DXVKVersion() != ""

	if useDXVK {
		if !c.Debug {
			env["DXVK_LOG_LEVEL"] = "warn"
		}
		env["DXVK_LOG_PATH"] = "none"
		env["DXVK_STATE_CACHE_PATH"] = dirs.Cache
	}

	for k, v := range env {
		pfx.Env = append(pfx.Env, k+"="+v)
	}

	dxvk.EnvOverride(pfx, useDXVK)

	slog.Debug("Using Prefix environment", "env", pfx.Env)

	return pfx
}
