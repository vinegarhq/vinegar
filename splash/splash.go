//go:build !nogui && !nosplash

package splash

import (
	"errors"
	"image"
	_ "image/png"
	"log"
	"os"

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
	"github.com/vinegarhq/vinegar/util"
)

var ErrClosed = errors.New("window closed")

type Splash struct {
	*app.Window

	Theme  *material.Theme
	Config *Config
	Style
	LogPath string

	logo    *image.Image
	message string
	desc    string

	progress float32
	closed   bool

	exitButton    *widget.Clickable
	openLogButton *widget.Clickable
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

func New(cfg *Config) *Splash {
	s := Compact

	if cfg.Style == "familiar" {
		s = Familiar
	}

	w := window(s.Size())
	w.Perform(system.ActionCenter)

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = material.Palette{
		Bg:         rgb(cfg.BgColor),
		Fg:         rgb(cfg.FgColor),
		ContrastBg: rgb(cfg.AccentColor),
		ContrastFg: rgb(cfg.InfoColor),
	}

	eb := new(widget.Clickable)
	olb := new(widget.Clickable)

	return &Splash{
		Theme:         th,
		Style:         s,
		Config:        cfg,
		Window:        w,
		exitButton:    eb,
		openLogButton: olb,
	}
}

func (ui *Splash) loadLogo() error {
	if ui.Config.LogoPath == "" {
		return errors.New("logo file path unset")
	}

	logoFile, err := os.Open(ui.Config.LogoPath)
	if err != nil {
		return err
	}
	defer logoFile.Close()

	logo, _, err := image.Decode(logoFile)
	if err != nil {
		return err
	}

	ui.logo = &logo
	return nil
}

func (ui *Splash) Run() error {
	drawfn := ui.drawCompact

	if err := ui.loadLogo(); err != nil {
		log.Println("Failed to load logo:", err)
	}

	defer func() {
		ui.closed = true
	}()

	if !ui.Config.Enabled {
		return nil
	}

	if ui.Style == Familiar {
		drawfn = ui.drawFamiliar
	}

	var ops op.Ops
	for {
		switch e := ui.NextEvent().(type) {
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

			if ui.openLogButton.Clicked(gtx) {
				log.Printf("Opening log file: %s", ui.LogPath)
				err := util.XDGOpen(ui.LogPath).Start()
				if err != nil {
					return err
				}
			}

			if ui.exitButton.Clicked(gtx) {
				ui.Perform(system.ActionClose)
			}

			drawfn(gtx)

			e.Frame(gtx.Ops)
		}
	}

	return nil
}
