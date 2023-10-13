//go:build nogui || nosplash

package splash

import (
	"errors"

	"github.com/vinegarhq/vinegar/internal/config"
)

type Splash struct {
	Config *config.Splash
}

func (ui *Splash) Message(msg string) {
}

func (ui *Splash) Desc(desc string) {
}

func (ui *Splash) ShowLog(name string) {
}

func (ui *Splash) Progress(progress float32) {
}

func (ui *Splash) Close() {
}

func New(cfg *config.Splash) *Splash {
	return &Splash{
		Config: cfg,
	}
}

func (ui *Splash) Run() error {
	if ui.Config.Enabled {
		return errors.New("Splash is enabled, despite being built with nosplash")
	}

	return nil
}
