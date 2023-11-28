package config

import (
	"errors"
	"strconv"
	"testing"

	"github.com/vinegarhq/vinegar/sysinfo"
)

func TestCard(t *testing.T) {
	b := Binary{
		ForcedGpu: "meow",
		Env:       Environment{},
	}
	sysinfo.Cards = []sysinfo.Card{}

	if err := b.pickCard(); !errors.Is(err, strconv.ErrSyntax) {
		t.Fatal("expected to not handle string gpu")
	}

	b.ForcedGpu = "1"
	if err := b.pickCard(); !errors.Is(err, ErrNoCardFound) {
		t.Fatal("expected to handle no gpu found")
	}

	b.ForcedGpu = "-1"
	if err := b.pickCard(); !errors.Is(err, ErrBadGpuIndex) {
		t.Fatal("expected to not handle negative gpu index")
	}
}

func TestIntegratedAndDiscreteCard(t *testing.T) {
	b := Binary{
		ForcedGpu: "integrated",
		Env:       Environment{},
	}
	sysinfo.Cards = []sysinfo.Card{
		{
			Driver:   "i915",
			Device:   "0000:01:00.0",
			Embedded: true,
		},
		{
			Driver:   "nvidia",
			Device:   "0000:02:00.0",
			Embedded: false,
		},
	}

	if err := b.pickCard(); err != nil {
		t.Fatal(err)
	}

	if v := b.Env["DRI_PRIME"]; v != "pci-0000_01_00_0" {
		t.Fatal("expected change in integrated prime index")
	}

	if v := b.Env["__GLX_VENDOR_LIBRARY_NAME"]; v != "mesa" {
		t.Fatal("expected glx vendor to be mesa")
	}

	b.Env = Environment{}
	b.ForcedGpu = "prime-discrete"

	if err := b.pickCard(); err != nil {
		t.Fatal(err)
	}

	if v := b.Env["DRI_PRIME"]; v != "pci-0000_02_00_0" {
		t.Fatal("expected change in discrete prime index")
	}

	if v := b.Env["__GLX_VENDOR_LIBRARY_NAME"]; v != "nvidia" {
		t.Fatal("expected glx vendor to be nvidia")
	}
}

func TestVulkanCard(t *testing.T) {
	b := Binary{
		ForcedGpu: "prime-discrete",
		Renderer:  "OpenGL",
		Env:       Environment{},
	}
	sysinfo.Cards = []sysinfo.Card{
		{
			Driver:   "i915",
			Embedded: true,
		},
		{
			Driver:   "nvidia",
			Embedded: false,
		},
		{
			Driver:   "radeon",
			Embedded: false,
		},
	}

	if err := b.pickCard(); !errors.Is(err, ErrOpenGLBlind) {
		t.Fatal("expected handle opengl skill issue")
	}
}
