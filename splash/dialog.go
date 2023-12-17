//go:build !nogui && !nosplash

package splash

import (
	"image"
	"log"

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

// Make a new application window using vinegar's existing properties to
// simulate a dialog. YesNo parameter dictates if Dialog returns a boolean
// based on if the user clicked 'Yes' or 'No' on the dialog, otherwise it will
// only make an 'Okay' button.
func (ui *Splash) Dialog(title, msg string, YesNo bool) (r bool) {
	var ops op.Ops
	width, height := 384, 152

	if YesNo {
		width, height = 400, 176
	}

	w := window(unit.Dp(width), unit.Dp(height))

	if !ui.Config.Enabled {
		log.Printf("Dialog: %s %s", title, msg)
		return
	}

	// This is required for time when Dialog is called before the main
	// window is ready for retrieving events.
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = material.Palette{
		Bg:         rgb(ui.Config.BgColor),
		Fg:         rgb(ui.Config.FgColor),
		ContrastBg: rgb(ui.Config.AccentColor),
		ContrastFg: rgb(ui.Config.InfoColor),
	}

	msgState := new(widget.Selectable)

	var yesButton widget.Clickable // Okay if !YesNo
	var noButton widget.Clickable

	for {
		switch e := w.NextEvent().(type) {
		case system.DestroyEvent:
			return r
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			paint.Fill(gtx.Ops, th.Palette.Bg)

			if yesButton.Clicked(gtx) {
				r = true
				w.Perform(system.ActionClose)
			}
			if noButton.Clicked(gtx) {
				w.Perform(system.ActionClose)
			}

			layout.UniformInset(18).Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(material.Body1(th, title).Layout),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx C) D {
						m := material.Body2(th, msg)
						m.State = msgState
						return m.Layout(gtx)
					}),
					// ugly hack to take up all remaining space
					layout.Flexed(1, func(gtx C) D {
						return layout.Dimensions{
							Size: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y},
						}
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{
							Axis:    layout.Horizontal,
							Spacing: layout.SpaceStart,
						}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								if !YesNo {
									return button(th, &yesButton, "Okay").Layout(gtx)
								}

								return layout.Inset{Right: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
									return button(ui.Theme, &yesButton, "Yes").Layout(gtx)
								})
							}),
							layout.Rigid(func(gtx C) D {
								if !YesNo {
									return D{}
								}
								btn := button(ui.Theme, &noButton, "No")
								btn.Color = ui.Theme.Palette.Fg
								btn.Background = rgb(ui.Config.CancelColor)
								return btn.Layout(gtx)
							}),
						)
					}),
				)
			})

			e.Frame(gtx.Ops)
		}
	}
}
