package main

import (
	"os"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/roblox"
)

type Binary struct {
	name string
	log <-chan string

	dir string
	pfx *wine.Prefix

	cfg  *config.Config
	bcfg *config.Application

	btype roblox.BinaryType
	ver   roblox.Version
}

func NewBinary(bt roblox.BinaryType, cfg *config.Config, pfx *wine.Prefix) Binary {
	var bcfg config.Application

	switch bt {
	case roblox.Player:
		bcfg = cfg.Player
	case roblox.Studio:
		bcfg = cfg.Studio
	}

	return Binary{
		name:  bt.String(),
		btype: bt,
		pfx:   pfx,
		cfg:   cfg,
		bcfg:  &bcfg,
	}
}

func (b *Binary) Run(args ...string) error {
	if err := b.Setup(); err != nil {
		return err
	}

	os.Exit(0)

	return b.Execute(args...)
}
