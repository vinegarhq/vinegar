package config

import (
	"fmt"

	"github.com/vinegarhq/vinegar/internal/adwaux"
)

// Backwards compatibility to allow:
// 'dxvk = true' and move to 'dxvk = [version]'
type DxvkVersion string

func (v *DxvkVersion) UnmarshalTOML(data interface{}) error {
	switch d := data.(type) {
	case bool:
		*v = ""
		if d {
			*v = "2.7.1"
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

type Renderer string

func (r Renderer) Values() []string {
	return []string{"D3D11", "D3D11FL10", "Vulkan", "OpenGL"}
}

func (r *Renderer) Select(s string) {
	// No enum parsing necessary
	*r = Renderer(s)
}

var _ adwaux.Selector = (*Renderer)(nil)
