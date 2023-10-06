//go:build nogui

package gui

import (
	"errors"

	"github.com/vinegarhq/vinegar/internal/config"
)

type UI struct {
	Config *config.UI
}

func (ui *UI) Message(msg string) {
}

func (ui *UI) Desc(desc string) {
}

func (ui *UI) ShowLog(name string) {
}

func (ui *UI) Progress(progress float32) {
}

func (ui *UI) Close() {
}

func New(cfg *config.UI) *UI {
	return &UI{
		Config: cfg,
	}
}

func (ui *UI) Run() error {
	if ui.Config.Enabled {
		return errors.New("GUI is enabled, despite being built with nogui")
	}

	return nil
}
