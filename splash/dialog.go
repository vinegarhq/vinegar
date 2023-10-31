//go:build !nogui && !nosplash

package splash

import (
	"log"

	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Make a new application window using vinegar's existing properties to
// simulate a dialog.
func (ui *Splash) Dialog(title, msg string) {
	var ops op.Ops
	var okButton widget.Clickable
	w := window(unit.Dp(480), unit.Dp(144))

	if !ui.Config.Enabled || ui.Theme == nil {
		log.Printf("Dialog: %s %s", title, msg)
		return
	}

	for e := range w.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			// no real care for errors, this is a dialog
			return
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			paint.Fill(gtx.Ops, ui.Theme.Palette.Bg)

			if okButton.Clicked() {
				w.Perform(system.ActionClose)
			}

			layout.Center.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(material.H6(ui.Theme, title).Layout),
					layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
					layout.Rigid(material.Body2(ui.Theme, msg).Layout),
					layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
					layout.Rigid(button(ui.Theme, &okButton, "Ok").Layout),
				)
			})

			e.Frame(gtx.Ops)
		}
	}
}
