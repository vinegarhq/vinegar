//go:build !nogui && !nosplash

package splash

import (
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
// simulate a dialog.
func (ui *Splash) Dialog(title, msg string) {
	var ops op.Ops
	var okButton widget.Clickable
	w := window(384, 152)

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

	for {
		switch e := w.NextEvent().(type) {
		case system.DestroyEvent:
			// no real care for errors, this is a dialog
			return
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			paint.Fill(gtx.Ops, th.Palette.Bg)

			if okButton.Clicked(gtx) {
				w.Perform(system.ActionClose)
			}

			layout.UniformInset(18).Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis:    layout.Vertical,
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Rigid(material.Body1(th, title).Layout),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx C) D {
						m := material.Body2(th, msg)
						m.State = msgState
						return m.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
							layout.Rigid(button(th, &okButton, "Okay").Layout),
						)
					}),
				)
			})

			e.Frame(gtx.Ops)
		}
	}
}
