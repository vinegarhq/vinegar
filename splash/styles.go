package splash

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type Style int

const (
	Compact Style = iota
	Familiar
)

func (s Style) Size() (w, h unit.Dp) {
	switch s {
	case Compact:
		w = unit.Dp(448)
		h = unit.Dp(150) // 118, 0
	case Familiar:
		w = unit.Dp(480)
		h = unit.Dp(246) // 198
	}

	return
}

func (ui *Splash) drawCompact(gtx C) D {
	return layout.Inset{
		Top:    unit.Dp(16),
		Bottom: unit.Dp(16),
		Left:   unit.Dp(16),
		Right:  unit.Dp(16),
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Start,
				}.Layout(gtx,
					layout.Rigid(ui.drawLogo().Layout),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{
							Axis:      layout.Vertical,
							Alignment: layout.Start,
						}.Layout(gtx,
							layout.Rigid(material.Label(ui.Theme, unit.Sp(16), ui.message).Layout),
							layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
							layout.Rigid(func(gtx C) D {
								return ui.drawDesc(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
							layout.Rigid(func(gtx C) D {
								pb := ProgressBar(ui.Theme, ui.progress)
								pb.TrackColor = rgb(ui.Config.TrackColor)
								return pb.Layout(gtx)
							}),
						)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
			layout.Rigid(func(gtx C) D {
				return ui.drawButtons(gtx, layout.SpaceStart)
			}),
		)
	})
}

func (ui *Splash) drawFamiliar(gtx C) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(ui.drawLogo().Layout),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(material.Label(ui.Theme, unit.Sp(16), ui.message).Layout),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{
							Top:    unit.Dp(16),
							Bottom: unit.Dp(16),
							Left:   unit.Dp(32),
							Right:  unit.Dp(32),
						}.Layout(gtx, func(gtx C) D {
							pb := ProgressBar(ui.Theme, ui.progress)
							pb.TrackColor = rgb(ui.Config.TrackColor)
							return pb.Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D {
						return ui.drawDesc(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
				)
			}),
			layout.Rigid(func(gtx C) D {
				return ui.drawButtons(gtx, layout.SpaceAround)
			}),
		)
	})
}
