package adwaux

import (
	"log/slog"
	"reflect"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
)

func newPathRow(v reflect.Value) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetSubtitle(v.String())
	row.AddCssClass("property")
	changed := func() {
		v.SetString(row.GetSubtitle())
		row.ActivateActionVariant("win.save", nil)
	}
	row.ConnectSignal("notify::subtitle", &changed)

	reset := gtk.NewButtonFromIconName("edit-undo-symbolic")
	reset.SetValign(gtk.AlignCenterValue)
	reset.AddCssClass("flat")
	reset.SetTooltipText("Reset to Default")
	resetClicked := func(_ gtk.Button) {
		row.SetSubtitle("")
	}
	reset.ConnectClicked(&resetClicked)
	row.AddSuffix(&reset.Widget)

	open := gtk.NewButtonFromIconName("document-open-symbolic")
	open.SetValign(gtk.AlignCenterValue)
	openClicked := func(_ gtk.Button) {
		dialog := gtk.NewFileDialog()
		var ready gio.AsyncReadyCallback = func(_, resPtr, _ uintptr) {
			res := gtkutil.AsyncResultFromInternalPtr(resPtr)
			f, err := dialog.SelectFolderFinish(res)
			if err != nil {
				slog.Error("FileDialog error", "err", err)
				return
			}
			row.SetSubtitle(f.GetPath())
		}
		win := gtk.WindowNewFromInternalPtr(row.GetRoot().Ptr)
		dialog.SelectFolder(win, nil, &ready, 0)
	}
	open.ConnectClicked(&openClicked)

	row.AddSuffix(&open.Widget)

	return row
}
