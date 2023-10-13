//go:build !nogui

package splash

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type ProgressBarStyle struct {
	Color      color.NRGBA
	TrackColor color.NRGBA
	Progress   float32
}

func ProgressBar(th *material.Theme, progress float32) ProgressBarStyle {
	return ProgressBarStyle{
		Progress:   progress,
		Color:      th.Palette.ContrastBg,
		TrackColor: mulAlpha(th.Palette.Fg, 0x05),
	}
}

func (p ProgressBarStyle) Layout(gtx layout.Context) layout.Dimensions {
	shader := func(width int, color color.NRGBA) layout.Dimensions {
		d := image.Point{X: width, Y: gtx.Dp(unit.Dp(6))}
		rr := gtx.Dp(4)

		defer clip.UniformRRect(image.Rectangle{Max: image.Pt(width, d.Y)}, rr).Push(gtx.Ops).Pop()
		paint.ColorOp{Color: color}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		return layout.Dimensions{Size: d}
	}

	progressBarWidth := gtx.Constraints.Max.X
	return layout.Stack{Alignment: layout.W}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return shader(progressBarWidth, p.TrackColor)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if p.Progress == 0.0 {
				return layout.Dimensions{}
			}

			fillWidth := int(float32(progressBarWidth) * clamp1(p.Progress))
			fillColor := p.Color
			return shader(fillWidth, fillColor)
		}),
	)
}

// clamp1 limits v to range [0..1].
func clamp1(v float32) float32 {
	if v >= 1 {
		return 1
	} else if v <= 0 {
		return 0
	} else {
		return v
	}
}
