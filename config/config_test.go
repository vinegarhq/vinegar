package config

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/vinegarhq/vinegar/roblox"
)

func TestBinarySetup(t *testing.T) {
	b := Binary{
		FFlags: make(roblox.FFlags),
		Env: Environment{
			"MEOW": "MEOW",
		},
	}

	if err := b.setup(); err != nil {
		t.Fatal(err)
	}

	b.Renderer = "Meow"
	if err := b.setup(); !errors.Is(err, roblox.ErrInvalidRenderer) {
		t.Error("expected renderer check")
	}

	b.Dxvk = true
	b.Renderer = "Vulkan"
	if err := b.setup(); !errors.Is(err, ErrNeedDXVKRenderer) {
		t.Error("expected dxvk appropiate renderer check")
	}

	b.Renderer = "D3D11FL10"
	if err := b.setup(); errors.Is(err, ErrNeedDXVKRenderer) {
		t.Error("expected not dxvk appropiate renderer check")
	}

	if os.Getenv("MEOW") == "MEOW" {
		t.Error("expected no change in environment")
	}

	b.Launcher = "_"
	if err := b.setup(); !errors.Is(err, exec.ErrNotFound) {
		t.Error("expected exec not found")
	}
}

func TestSetup(t *testing.T) {
	wr := t.TempDir()
	c := Config{
		WineRoot: wr,
	}

	if err := c.setup(); !errors.Is(err, ErrWineRootInvalid) {
		t.Error("expected wine root wine check")
	}

	c.WineRoot = filepath.Join(".", wr)
	if err := c.setup(); !errors.Is(err, ErrWineRootAbs) {
		t.Error("expected wine root absolute path")
	}
}
