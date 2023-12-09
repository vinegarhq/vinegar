//go:build !nogui && !nosplash

package splash

import (
	"log"

	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Make a new application window using vinegar's existing properties to
// simulate a dialog.
func (ui *Splash) Dialog(title, msg string) {
	var ops op.Ops
	var okButton widget.Clickable
	w := window(384, 120)

	if !ui.Config.Enabled || ui.Theme == nil {
		log.Printf("Dialog: %s %s", title, msg)
		return
	}

	for {
		switch e := ui.NextEvent().(type) {
		case system.DestroyEvent:
			// no real care for errors, this is a dialog
			return
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			paint.Fill(gtx.Ops, ui.Theme.Palette.Bg)

			if okButton.Clicked(gtx) {
				w.Perform(system.ActionClose)
			}

			layout.UniformInset(18).Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis:    layout.Vertical,
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Rigid(material.Body2(ui.Theme, title).Layout),
					layout.Rigid(material.Body2(ui.Theme, msg).Layout),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
							layout.Rigid(button(ui.Theme, &okButton, "Ok").Layout),
						)
					}),
				)
			})

			e.Frame(gtx.Ops)
		}
	}
}
