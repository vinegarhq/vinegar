package splash

import (
	"image"
	"log"

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

// Make a new application window using vinegar's existing properties to
// simulate a dialog. user parameter dictates if Dialog returns a boolean
// based on if the user clicked 'Yes' or 'No' on the dialog, otherwise it will
// only make an 'Okay' button.
//
// The dialog window size will automatically resize itself vertically
// according to how many lines the text takes.
func (ui *Splash) Dialog(txt string, user bool, xdg string) (r bool) {
	var ops op.Ops

	if !ui.Config.Enabled {
		log.Printf("Dialog: %s", txt)
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

	width := 384
	// DOES NOT ACCOUNT FOR SCALING FACTORS
	// PLZ FIX GIO
	height := func() int {
		l := material.Body2(th, txt)
		gtx := layout.Context{
			Ops: &ops,
			Constraints: layout.Constraints{
				Min: image.Point{X: width, Y: 0},
				Max: image.Point{X: width, Y: width * 4},
			},
		}
		defer ops.Reset()

		return l.Layout(gtx).Size.Y + (18 * 2) + 57 + 4
	}()

	w := window(unit.Dp(width), unit.Dp(height))

	msgState := new(widget.Selectable)

	var yesButton widget.Clickable // Okay if !user
	var noButton widget.Clickable
	var xdgButton widget.Clickable

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return r
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			paint.Fill(gtx.Ops, th.Palette.Bg)

			if yesButton.Clicked(gtx) {
				r = true
				w.Perform(system.ActionClose)
			}
			if noButton.Clicked(gtx) {
				w.Perform(system.ActionClose)
			}
			if xdgButton.Clicked(gtx) {
				if err := XDGOpen(xdg).Start(); err != nil {
					log.Printf("Dialog: xdg-open: %s", err)
				}
			}

			layout.UniformInset(18).Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						m := material.Body2(th, txt)
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
								if !user {
									r := unit.Dp(16)
									if xdg == "" {
										r = 0
									}
									return layout.Inset{Right: r}.Layout(gtx, func(gtx C) D {
										return button(th, &yesButton, "Okay").Layout(gtx)
									})
								}

								return layout.Inset{Right: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
									return button(ui.Theme, &yesButton, "Yes").Layout(gtx)
								})
							}),
							layout.Rigid(func(gtx C) D {
								if !user {
									return D{}
								}
								btn := button(ui.Theme, &noButton, "No")
								btn.Color = ui.Theme.Palette.Fg
								btn.Background = rgb(ui.Config.CancelColor)
								return btn.Layout(gtx)
							}),
							layout.Rigid(func(gtx C) D {
								if user || xdg == "" {
									return D{}
								}

								return button(th, &xdgButton, "More Info").Layout(gtx)
							}),
						)
					}),
				)
			})

			e.Frame(gtx.Ops)
		}
	}
}
