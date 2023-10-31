//go:build !nogui && !nosplash

package splash

import (
	"bytes"
	"errors"
	"image"
	_ "image/png"
	"log"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/util"
)

var ErrClosed = errors.New("window closed")

type Splash struct {
	*app.Window

	Theme  *material.Theme
	Config *config.Splash
	Style
	LogPath string

	logo    image.Image
	message string
	desc    string

	progress float32
	closed   bool

	exitButton    widget.Clickable
	openLogButton widget.Clickable
}

func (ui *Splash) SetMessage(msg string) {
	ui.message = msg
	ui.Invalidate()
}

func (ui *Splash) SetDesc(desc string) {
	ui.desc = desc
	ui.Invalidate()
}

func (ui *Splash) SetProgress(progress float32) {
	ui.progress = progress
	ui.Invalidate()
}

func (ui *Splash) Close() {
	ui.closed = true
	ui.Perform(system.ActionClose)
}

func (ui *Splash) IsClosed() bool {
	return ui.closed
}

func window(width, height unit.Dp) *app.Window {
	return app.NewWindow(
		app.Decorated(false),
		app.Size(width, height),
		app.MinSize(width, height),
		app.MaxSize(width, height),
		app.Title("Vinegar"),
	)
}

func New(cfg *config.Splash) *Splash {
	s := Compact

	if cfg.Style == "familiar" {
		s = Familiar
	}

	logo, _, _ := image.Decode(bytes.NewReader(vinegarlogo))
	w := window(s.Size())
	w.Perform(system.ActionCenter)

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = material.Palette{
		Bg:         rgb(cfg.Bg),
		Fg:         rgb(cfg.Fg),
		ContrastBg: rgb(cfg.Accent),
		ContrastFg: rgb(cfg.Gray2),
	}

	return &Splash{
		logo:   logo,
		Theme:  th,
		Style:  s,
		Config: cfg,
		Window: w,
	}
}

func (ui *Splash) Run() error {
	var ops op.Ops
	drawfn := ui.drawCompact

	if !ui.Config.Enabled {
		ui.closed = true
		return nil
	}

	if ui.Style == Familiar {
		drawfn = ui.drawFamiliar
	}

	for e := range ui.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			if ui.closed && e.Err == nil {
				return nil
			} else if e.Err == nil {
				return ErrClosed
			} else {
				return e.Err
			}
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			paint.Fill(gtx.Ops, ui.Theme.Palette.Bg)

			if ui.openLogButton.Clicked() {
				log.Printf("Opening log file: %s", ui.LogPath)
				err := util.XDGOpen(ui.LogPath).Start()
				if err != nil {
					return err
				}
			}

			if ui.exitButton.Clicked() {
				ui.Perform(system.ActionClose)
			}

			drawfn(gtx)

			e.Frame(gtx.Ops)
		}
	}

	ui.closed = true
	return nil
}
