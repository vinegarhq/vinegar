//go:build nogui || nosplash

package splash

import (
	"errors"
	"log"

	"github.com/vinegarhq/vinegar/config"
)

var ErrClosed = errors.New("window closed")

type Splash struct {
	LogPath string
}

func (ui *Splash) SetMessage(msg string) {
}

func (ui *Splash) SetDesc(desc string) {
}

func (ui *Splash) SetProgress(progress float32) {
}

func (ui *Splash) Close() {
}

func (ui *Splash) Invalidate() {
}

func (ui *Splash) IsClosed() bool {
	return true
}

func (ui *Splash) Dialog(title, msg string) {
	log.Printf("splash: dialog: %s %s", title, msg)
}

func New(cfg *config.Splash) *Splash {
	return &Splash{}
}

func (ui *Splash) Run() error {
	return nil
}