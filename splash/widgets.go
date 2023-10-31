//go:build !nogui && !nosplash

package splash

import (
	_ "image/png"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func (ui *Splash) buttons(gtx C, s layout.Spacing) D {
	inset := layout.Inset{
		Top:   unit.Dp(10),
		Left:  unit.Dp(10),
	}

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: s,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if ui.LogPath == "" {
				return D{}
			}
			btn := button(ui.Theme, &ui.openLogButton, "Show logs")
			return inset.Layout(gtx, func(gtx C) D {
				return btn.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			btn := button(ui.Theme, &ui.exitButton, "Cancel")
			btn.Background = rgb(ui.Config.Red)
			return inset.Layout(gtx, func(gtx C) D {
				return btn.Layout(gtx)
			})
		}),
	)
}

func (ui *Splash) drawDesc(gtx C) D {
	d := material.Caption(ui.Theme, ui.desc)
	d.Font.Typeface = "go mono, monospace"
	d.Color = ui.Theme.Palette.ContrastFg
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
