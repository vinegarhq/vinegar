package roblox

import (
	"errors"
	"maps"
	"testing"
)

func TestFFlagRenderer(t *testing.T) {
	f := make(FFlags)

	if err := f.SetRenderer(""); err != nil {
		t.Error("expected no failure with no renderer")
	}

	expectedUnset := FFlags{
		"FFlagDebugGraphicsPreferOpenGL":     false,
		"FFlagDebugGraphicsPreferD3D11FL10":  false,
		"FFlagDebugGraphicsPreferD3D11":      true,
		"FFlagDebugGraphicsPreferVulkan":     false,
		"FFlagDebugGraphicsDisableOpenGL":    true,
		"FFlagDebugGraphicsDisableD3D11FL10": true,
		"FFlagDebugGraphicsDisableD3D11":     false,
		"FFlagDebugGraphicsDisableVulkan":    true,
	}

	if !maps.Equal(f, expectedUnset) {
		t.Error("expected fflag set renderer no renderer to match expected d3d11 set")
	}

	if err := f.SetRenderer("meow"); !errors.Is(err, ErrInvalidRenderer) {
		t.Error("expected invalid renderer check")
	}

	if err := f.SetRenderer("Vulkan"); err != nil {
		t.Error("expected no failure with correct renderer")
	}

	expectedSet := FFlags{
		"FFlagDebugGraphicsPreferOpenGL":     false,
		"FFlagDebugGraphicsPreferD3D11FL10":  false,
		"FFlagDebugGraphicsPreferD3D11":      false,
		"FFlagDebugGraphicsPreferVulkan":     true,
		"FFlagDebugGraphicsDisableOpenGL":    true,
		"FFlagDebugGraphicsDisableD3D11FL10": true,
		"FFlagDebugGraphicsDisableD3D11":     true,
		"FFlagDebugGraphicsDisableVulkan":    false,
	}

	if !maps.Equal(f, expectedSet) {
		t.Error("expected fflag set renderer vulkan to match expected vulkan set")
	}
}
