// Package splash contains routines to run Vinegar's splash window.
package splash

import (
	"bytes"
	_ "embed"
	"errors"
	"image"
	_ "image/png"
	"io"
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
)

//go:embed vinegar.png
var vinegarLogo []byte

var ErrClosed = errors.New("window closed")

type Config struct {
	Enabled     bool   `toml:"enabled"`     // Determines if splash is shown or not
	LogoPath    string `toml:"logo_path"`   // Logo file path used to load and render the logo
	Style       string `toml:"style"`       // Style to use for the splash layout
	BgColor     uint32 `toml:"background"`  // Foreground color
	FgColor     uint32 `toml:"foreground"`  // Background color
	CancelColor uint32 `toml:"cancel,red"`  // Background color for the Cancel button
	AccentColor uint32 `toml:"accent"`      // Color for progress bar's track and ShowLog button
	TrackColor  uint32 `toml:"track,gray1"` // Color for the progress bar's background
	InfoColor   uint32 `toml:"info,gray2"`  // Foreground color for the text containing binary information
}

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
	if ui.Window == nil {
		return
	}

	ui.message = msg
	ui.Invalidate()
}

func (ui *Splash) SetDesc(desc string) {
	if ui.Window == nil {
		return
	}

	ui.desc = desc
	ui.Invalidate()
}

func (ui *Splash) SetProgress(progress float32) {
	if ui.Window == nil {
		return
	}

	ui.progress = progress
	ui.Invalidate()
}

func (ui *Splash) Close() {
	if ui.Window == nil {
		return
	}

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
	if !cfg.Enabled {
		return &Splash{
			closed: true,
			Config: cfg,
		}
	}

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
		closed:        true,
	}
}

func (ui *Splash) loadLogo() error {
	var r io.Reader

	if ui.Config.LogoPath != "" {
		lf, err := os.Open(ui.Config.LogoPath)
		if err != nil {
			return err
		}
		defer lf.Close()
		r = lf
	} else {
		r = bytes.NewReader(vinegarLogo)
	}

	logo, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	ui.logo = &logo
	return nil
}

func (ui *Splash) Run() error {
	if ui.closed {
		return nil
	}

	drawfn := ui.drawCompact

	if err := ui.loadLogo(); err != nil {
		log.Println("Failed to load logo:", err)
	}

	defer func() {
		ui.closed = true
	}()

	if ui.Style == Familiar {
		drawfn = ui.drawFamiliar
	}

	ui.closed = false
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
				err := XDGOpen(ui.LogPath).Start()
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
