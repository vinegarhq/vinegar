package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/internal/adwaux"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/sysinfo"
)

type WineOption string

func (WineOption) Default() string {
	// Kombucha only has glibc builds currently.
	//
	// Ironic because I use a musl system.
	if strings.Contains(sysinfo.LibC, "musl") {
		return ""
	}
	return dirs.WinePath
}

func (o WineOption) IsDefault() bool {
	return string(o) == o.Default()
}

func (o *WineOption) SetDefault() {
	*o = WineOption(o.Default())
}

func (WineOption) SelectPath() {}

type Renderer string

func (r Renderer) Values() []string {
	return []string{"D3D11", "DXVK", "DXVK-Sarek", "Vulkan", "OpenGL"}
}

func (r Renderer) IsDXVK() bool {
	return strings.HasPrefix(string(r), "DXVK")
}

func (r Renderer) DXVKVersion() string {
	if r == "DXVK-Sarek" {
		return dxvkSarekVersion
	}
	return dxvkVersion
}

type DesktopOption string

func (DesktopOption) Default() string {
	return "1814x1024"
}

func (r *Renderer) Select(s string) {
	// No enum parsing necessary
	*r = Renderer(s)
}

// Allow 'webview = true' when the internal value is set to the default.
type WebViewOption string

func (o *WebViewOption) UnmarshalTOML(data interface{}) error {
	switch d := data.(type) {
	case bool:
		*o = ""
		if d {
			*o = WebViewOption(o.Default())
		}
	case string:
		*o = WebViewOption(d)
	default:
		return fmt.Errorf("unsupported type: %T", d)
	}
	return nil
}

func (o WebViewOption) MarshalTOML() ([]byte, error) {
	if string(o) == o.Default() {
		return []byte("true"), nil
	}
	return []byte(`"` + o + `"`), nil
}

func (o WebViewOption) Default() string {
	return webViewVersion
}

func (o WebViewOption) String() string {
	return string(o)
}

func (o WebViewOption) Enabled() bool {
	return string(o) == ""
}

func (o *WebViewOption) SetDefault() {
	*o = WebViewOption(o.Default())
}

var _ adwaux.Selector = (*Renderer)(nil)
var _ adwaux.Defaulter = (*DesktopOption)(nil)
var _ adwaux.PathSelector = (*WineOption)(nil)
var _ toml.Marshaler = (*WebViewOption)(nil)
var _ toml.Unmarshaler = (*WebViewOption)(nil)
