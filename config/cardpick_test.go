package config

import (
	"errors"
	"strconv"
	"testing"
	"github.com/vinegarhq/vinegar/sysinfo"
)

func TestCard(t *testing.T) {
	sysinfo.Cards = []sysinfo.Card{}
	b := Binary{
		ForcedGpu: "meow",
		Env: Environment{},
	}

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

func TestPRIMECard(t *testing.T) {
	b := Binary{
		ForcedGpu: "integrated",
		Env: Environment{},
	}
	sysinfo.Cards = []sysinfo.Card{
		{
			Driver: "i915",
			Embedded: true,
		},
	}

	if err := b.pickCard(); err != nil {
		t.Fatal(err)
	}

	if _, ok := b.Env["DRI_PRIME"]; ok {
		t.Fatal("expected no change")
	}

	sysinfo.Cards = append(sysinfo.Cards, sysinfo.Card{
		Driver: "nvidia",
		Embedded: false,
	})

	if err := b.pickCard(); err != nil {
		t.Fatal(err)
	}

	if v := b.Env["DRI_PRIME"]; v != "0" {
		t.Fatal("expected change in prime index")
	}

	if v := b.Env["__GLX_VENDOR_LIBRARY_NAME"]; v != "mesa" {
		t.Fatal("expected glx vendor to be mesa")
	}

	b.ForcedGpu = "prime-discrete"
	delete(b.Env, "DRI_PRIME")
	delete(b.Env, "__GLX_VENDOR_LIBRARY_NAME")
	if err := b.pickCard(); err != nil {
		t.Fatal(err)
	}

	if v := b.Env["DRI_PRIME"]; v != "1" {
		t.Fatal("expected change in descrete index")
	}

	if v := b.Env["__GLX_VENDOR_LIBRARY_NAME"]; v != "nvidia" {
		t.Fatal("expected glx vendor to be nvidia")
	}
}