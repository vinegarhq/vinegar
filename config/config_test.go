package config

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/wine"
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
	c := Default()
	c.WineRoot = wr

	// Required to not conflict with system environment
	os.Setenv("PATH", "")

	if err := c.setup(); !errors.Is(err, ErrWineRootInvalid) {
		t.Error("expected wine root wine check")
	}

	if err := os.Mkdir(filepath.Join(wr, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := os.OpenFile(filepath.Join(wr, "bin", wine.Wine), os.O_CREATE, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.setup(); err != nil {
		t.Error("valid wine root is invalid")
	}

	c.WineRoot = filepath.Join(".", wr)
	if err := c.setup(); !errors.Is(err, ErrWineRootAbs) {
		t.Error("expected wine root absolute path")
	}
}
