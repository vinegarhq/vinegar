//go:build !nogui
// +build !nogui

package main

import (
	"fmt"
	"image"
	"image/color"
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
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func logoImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func (b *Binary) SendLog(msg string) {
	if b.cfg.UI.Enabled {
		b.log <- msg
	}
}

func (b *Binary) SendProgress(progress float32) {
	if b.cfg.UI.Enabled {
		b.progress <- progress
	}
}

func (b *Binary) Glass(exit <-chan bool) {
	var ops op.Ops
	var logMsg string
	var progress float32

	width := unit.Dp(448)
	height := unit.Dp(192)
	w := app.NewWindow(
		app.Decorated(false),
		app.Size(width, height),
		app.MinSize(width, height),
		app.MaxSize(width, height),
		app.Title("Vinegar"),
	)

	logo, err := logoImage(b.cfg.Logo)
	if err != nil {
		log.Printf("Failed to load logo: %s", err)

		emptyImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
		emptyImg.Set(0, 0, color.NRGBA{0, 0, 0, 0})
		logo = emptyImg
	}

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = material.Palette{
		Fg:         rgb(b.cfg.UI.Foreground),
		Bg:         rgb(b.cfg.UI.Background),
		ContrastBg: rgb(b.cfg.UI.Accent),
		ContrastFg: rgb(b.cfg.UI.Foreground),
	}

	for {
		select {
		case logMsg = <-b.log:
			w.Invalidate()
		case progress = <-b.progress:
			w.Invalidate()
		case <-exit:
			w.Perform(system.ActionClose)
			return
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return
			case system.FrameEvent:
				log.Println("FrameEvent")
				gtx := layout.NewContext(&ops, e)

				paint.Fill(gtx.Ops, th.Palette.Bg)
				layout.Center.Layout(gtx, func(gtx C) D {
					return layout.Flex{
						Axis:      layout.Vertical,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Rigid(widget.Image{Src: paint.NewImageOp(logo)}.Layout),
						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
						layout.Rigid(material.Label(th, unit.Sp(16), logMsg).Layout),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{
								Top:    unit.Dp(16),
								Bottom: unit.Dp(20),
								Right:  unit.Dp(25),
								Left:   unit.Dp(25),
							}.Layout(gtx, func(gtx C) D {
								pb := material.ProgressBar(th, progress)
								pb.TrackColor = rgb(b.cfg.Gray1)
								return pb.Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx C) D {
							info := material.Body2(th,
								fmt.Sprintf("%s %s — %s", b.name, b.ver.Channel, b.ver.GUID),
							)
							info.Color = rgb(b.cfg.Gray2)
							return info.Layout(gtx)
						}),
					)
				})

				e.Frame(gtx.Ops)
			}
		}
	}
}

func rgb(c uint32) color.NRGBA {
	return argb(0xff000000 | c)
}

func argb(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}
