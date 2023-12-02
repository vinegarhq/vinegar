package roblox

import (
	"testing"
	"errors"
	"maps"
)

func TestFFlagRenderer(t *testing.T) {
	f := make(FFlags)

	if err := f.SetRenderer(""); err != nil {
		t.Error("expected no failure with no renderer")
	}

	if !maps.Equal(f, FFlags{}) {
		t.Error("expected no change with no renderer")
	}

	if err := f.SetRenderer("meow"); !errors.Is(err, ErrInvalidRenderer) {
		t.Error("expected invalid renderer check")
	}

	if err := f.SetRenderer("Vulkan"); err != nil {
		t.Error("expected no failure with correct renderer")
	}

	expected := FFlags{
		"FFlagDebugGraphicsPreferOpenGL": false,
		"FFlagDebugGraphicsPreferD3D11FL10": false,
		"FFlagDebugGraphicsPreferD3D11": false,
		"FFlagDebugGraphicsPreferVulkan": true,
		"FFlagDebugGraphicsDisableOpenGL": true,
		"FFlagDebugGraphicsDisableD3D11FL10": true,
		"FFlagDebugGraphicsDisableD3D11": true,
		"FFlagDebugGraphicsDisableVulkan": false,
	}

	if !maps.Equal(f, expected) {
		t.Error("expected fflag set renderer to match expected set")
	}
}
