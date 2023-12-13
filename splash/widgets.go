//go:build !nogui && !nosplash

package splash

import (
	_ "image/png"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func (ui *Splash) drawLogo() *widget.Image {
	if ui.logo == nil {
		return &widget.Image{}
	}

	return &widget.Image{Src: paint.NewImageOp(*ui.logo)}
}

func (ui *Splash) drawButtons(gtx C, s layout.Spacing) D {
	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: s,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if ui.LogPath == "" {
				return D{}
			}

			btn := button(ui.Theme, ui.openLogButton, "Show logs")
			return layout.Inset{Right: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
				return btn.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			btn := button(ui.Theme, ui.exitButton, "Cancel")
			btn.Color = ui.Theme.Palette.Fg
			btn.Background = rgb(ui.Config.CancelColor)
			return btn.Layout(gtx)
		}),
	)
}

func (ui *Splash) drawDesc(gtx C) D {
	d := material.Caption(ui.Theme, ui.desc)
	d.Font.Typeface = "go mono, monospace"
	d.Color = rgb(ui.Config.InfoColor)
	return d.Layout(gtx)
}

func button(th *material.Theme, b *widget.Clickable, txt string) (bs material.ButtonStyle) {
	bs = material.Button(th, b, txt)
	bs.Inset = layout.Inset{
		Top: unit.Dp(10), Bottom: unit.Dp(10),
		Left: unit.Dp(16), Right: unit.Dp(16),
	}
	bs.Color = th.Palette.Fg
	bs.CornerRadius = 6
	return
}
