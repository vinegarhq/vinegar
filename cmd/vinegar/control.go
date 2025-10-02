package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/sysinfo"
)

type control struct {
	*app

	builder *gtk.Builder
	win     adw.ApplicationWindow
}

func (s *app) newControl() *control {
	ctl := control{
		app:     s,
		builder: gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/control.ui"),
	}

	ctl.builder.GetObject("window").Cast(&ctl.win)
	ctl.win.SetApplication(&s.Application.Application)

	abt := gio.NewSimpleAction("about", nil)
	abtcb := func(_ gio.SimpleAction, p uintptr) {
		w := adw.NewAboutDialogFromAppdata("/org/vinegarhq/Vinegar/metainfo.xml", version[1:])
		w.Present(&ctl.win.Widget)
		w.SetDebugInfo(ctl.debugInfo())
		w.Unref()
	}
	abt.ConnectActivate(&abtcb)
	ctl.AddAction(abt)
	abt.Unref()

	logsA := gio.NewSimpleAction("open-log-dir", nil)
	logsCb := func(_ gio.SimpleAction, p uintptr) {
		gtk.ShowUri(&ctl.win.Window, "file://"+dirs.Logs, 0)
	}
	logsA.ConnectActivate(&logsCb)
	ctl.AddAction(logsA)
	logsA.Unref()

	ctl.configPut()
	ctl.updateButtons()

	toastA := gio.NewSimpleAction("control-toast", glib.NewVariantType("s"))
	toastCb := func(a gio.SimpleAction, p uintptr) {
		msg := (*glib.Variant)(unsafe.Pointer(p)).GetString(0)
		var overlay adw.ToastOverlay
		ctl.builder.GetObject("overlay").Cast(&overlay)
		toast := adw.NewToast(msg)
		overlay.AddToast(toast)
	}
	toastA.ConnectActivate(&toastCb)
	ctl.AddAction(toastA)
	toastA.Unref()

	// For the time being, use in-house editing.
	// ctl.setupConfigurationActions()
	ctl.setupControlActions()

	var wineRunner adw.EntryRow
	ctl.builder.GetObject("prefix-custom-run").Cast(&wineRunner)
	applyCb := func(_ adw.EntryRow) {
		ctl.ActivateAction("wine-run", nil)
	}
	wineRunner.ConnectApply(&applyCb)

	return &ctl
}

func (ctl *control) configPut() {
	var view gtk.TextView
	ctl.builder.GetObject("config-view").Cast(&view)

	b, err := os.ReadFile(dirs.ConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		b = []byte(err.Error())
	}

	view.GetBuffer().SetText(string(b), -1)
}

func (ctl *control) setupLogDialog() *adw.Dialog {
	var dialog adw.Dialog
	ctl.builder.GetObject("dialog-working").Cast(&dialog)

	var view gtk.TextView
	ctl.builder.GetObject("textview-log").Cast(&view)
	buf := view.GetBuffer()

	h, _ := slog.Default().Handler().(*logging.Handler)

	lvlTags := make(map[slog.Level]*gtk.TextTag, 6)
	for lvl, color := range map[slog.Level]string{
		slog.LevelDebug:             "white",
		slog.LevelInfo:              "lime",
		logging.LevelWine.Level():   "darkred",
		logging.LevelRoblox.Level(): "cyan",
		slog.LevelWarn:              "yellow",
		slog.LevelError:             "red",
	} {
		lvlTags[lvl] = buf.CreateTag(lvl.String(), "foreground", color)
	}

	dialogShowCb := func(_ gtk.Widget) {
		h.ReadRecord = func(r *slog.Record) {
			tag := lvlTags[r.Level]
			name := logging.FromLevel(r.Level).String() + " "
			uiThread(func() {
				var iter gtk.TextIter
				buf.GetEndIter(&iter)
				buf.InsertWithTags(&iter, name, -1, tag)
				buf.Insert(&iter, r.Message+"\n", -1)
				view.ScrollToIter(&iter, 0, false, 0, 0)
			})
		}
	}
	dialog.ConnectMap(&dialogShowCb)

	dialogClosedCb := func(_ adw.Dialog) {
		h.ReadRecord = nil
		var start, end gtk.TextIter
		buf.GetStartIter(&start)
		buf.GetEndIter(&end)
		buf.Delete(&start, &end)
	}
	dialog.ConnectClosed(&dialogClosedCb)

	return &dialog
}

func (ctl *control) setupControlActions() {
	actions := map[string]interface{}{
		"run-studio":       (*bootstrapper).start,
		"install-studio":   (*bootstrapper).setup,
		"uninstall-studio": ctl.deleteDeployments,
		"kill-prefix":      ctl.pfx.Kill,
		"init-prefix":      (*bootstrapper).setupPrefix,
		"delete-prefix":    ctl.deletePrefixes,
		"wine-run-winecfg": func() error {
			return ctl.pfx.Wine("winecfg").Run()
		},
		"clear-cache": ctl.clearCache,
		"save-config": ctl.configSave,

		"wine-run": func() error {
			var runner adw.EntryRow
			ctl.builder.GetObject("prefix-custom-run").Cast(&runner)
			cmd := runner.GetText()
			args := strings.Fields(cmd)
			return ctl.pfx.Wine(args[0], args[1:]...).Run()
		},
	}

	logDialog := ctl.setupLogDialog()

	for name, action := range actions {
		act := gio.NewSimpleAction(name, nil)
		actcb := func(_ gio.SimpleAction, p uintptr) {
			var run func() error

			switch f := action.(type) {
			case func() error:
				run = func() error {
					return f()
				}
			case func(*bootstrapper) error:
				if ctl.boot == nil {
					ctl.boot = ctl.newBootstrapper()
				}
				ctl.boot.win.SetTransientFor(&ctl.win.Window)
				run = func() error {
					defer uiThread(ctl.boot.win.Destroy)
					return f(ctl.boot)
				}
			default:
				panic("unreachable")
			}

			logDialog.Present(&ctl.win.Widget)
			ctl.app.errThread(func() error {
				defer uiThread(func() {
					ctl.updateButtons()
					logDialog.ForceClose()
				})
				return run()
			})
		}

		act.ConnectActivate(&actcb)
		ctl.AddAction(act)
		act.Unref()
	}
}

/*
func (ctl *control) setupConfigurationActions() {
	props := []struct {
		name   string
		widget string
		modify any
	}{
		{"gamemode", "switch", ctl.cfg.Studio.GameMode},
		{"launcher", "entry", ctl.cfg.Studio.Launcher},
	}

	save := func() {
		if err := ctl.LoadConfig(); err != nil {
			ctl.presentSimpleError(err)
		}
	}

	for _, p := range props {
		obj := ctl.builder.GetObject(p.name)

		switch p.widget {
		case "switch":
			// AdwSwitchRow does not implement Gtk.Switch, and neither
			// does puregotk have a reference for AdwSwitchRow.
			actcb := func() {
				var new bool
				obj.Get("active", &new)
				p.modify = new
				slog.Info("Configuration Switch", "name", p.widget, "value", p.modify)
				save()
			}
			var w gtk.Widget
			obj.Cast(&w)
			obj.ConnectSignal("notify::active", &actcb)
			slog.Info("Setup Widget", "name", p.name)
		case "entry":
			var entry adw.EntryRow
			obj.Cast(&entry)

			cb := func(_ adw.EntryRow) {
				p.modify = entry.GetText()
				slog.Info("Configuration Entry", "name", p.widget, "value", p.modify)
				save()
			}
			entry.ConnectApply(&cb)
		default:
			panic("unreachable")
		}
	}

	// this is a special button!
	var rootRow adw.ActionRow
	ctl.builder.GetObject("wineroot-row").Cast(&rootRow)
	rootRow.SetSubtitle(ctl.cfg.Studio.WineRoot)

	var rootSelect gtk.Button
	ctl.builder.GetObject("wineroot-select").Cast(&rootSelect)
	actcb := func(_ gtk.Button) {
		// gtk.FileChooser is deprecated for gtk.FileDialog, but puregotk does not have it.
		fc := gtk.NewFileChooserDialog("Select Wine Installation", &ctl.win.Window,
			gtk.FileChooserActionSelectFolderValue,
			"Cancel", gtk.ResponseCancelValue,
			"Select", gtk.ResponseAcceptValue,
			unsafe.Pointer(nil),
		)
		rcb := func(_ gtk.Dialog, _ int) {
			cp := ctl.cfg.Studio
			cp.WineRoot = fc.GetFile().GetPath()
			// if err := cp.Setup(); err != nil {
			rootRow.SetSubtitle(ctl.cfg.Studio.WineRoot)
			fc.Destroy()
		}
		fc.ConnectResponse(&rcb)
		fc.Present()
	}
	rootSelect.ConnectClicked(&actcb)
}
*/

func (ctl *control) configSave() error {
	var view gtk.TextView
	var start, end gtk.TextIter
	ctl.builder.GetObject("config-view").Cast(&view)

	buf := view.GetBuffer()
	buf.GetBounds(&start, &end)
	text := buf.GetText(&start, &end, false)

	if err := dirs.Mkdirs(dirs.Config); err != nil {
		return err
	}

	f, err := os.OpenFile(dirs.ConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(text)); err != nil {
		return err
	}

	ctl.ActivateAction("control-toast", glib.NewVariantString("Saved!"))

	return ctl.loadConfig()
}

func (ctl *control) deleteDeployments() error {
	if err := os.RemoveAll(dirs.Versions); err != nil {
		return err
	}

	ctl.state.Studio.Version = ""
	ctl.state.Studio.Packages = nil

	return ctl.state.Save()
}

func (ctl *control) deletePrefixes() error {
	slog.Info("Deleting Wineprefixes!")

	if err := ctl.pfx.Kill(); err != nil {
		return fmt.Errorf("kill prefix: %w", err)
	}

	if err := os.RemoveAll(dirs.Prefixes); err != nil {
		return err
	}

	ctl.state.Studio.DxvkVersion = ""

	if err := ctl.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func (ctl *control) clearCache() error {
	return filepath.WalkDir(dirs.Cache, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if path == dirs.Cache || path == dirs.Logs || path == logging.Path {
			return nil
		}

		slog.Info("Removing cache file", "path", path)
		return os.RemoveAll(path)
	})
}

func (ctl *control) updateButtons() {
	pfx := ctl.pfx.Exists()
	vers := dirs.Empty(dirs.Versions)

	var launch, inst, uninst gtk.Widget
	ctl.builder.GetObject("studio-run").Cast(&launch)
	ctl.builder.GetObject("studio-install").Cast(&inst)
	ctl.builder.GetObject("studio-uninstall").Cast(&uninst)
	launch.SetVisible(!vers)
	inst.SetVisible(vers)
	uninst.SetVisible(!vers)

	var init, kill, del, cfg, run gtk.Widget
	ctl.builder.GetObject("prefix-init").Cast(&init)
	ctl.builder.GetObject("prefix-kill").Cast(&kill)
	ctl.builder.GetObject("prefix-delete").Cast(&del)
	ctl.builder.GetObject("prefix-configure").Cast(&cfg)
	ctl.builder.GetObject("prefix-custom-run").Cast(&run)
	init.SetVisible(!pfx)
	kill.SetVisible(pfx)
	del.SetVisible(pfx)
	cfg.SetVisible(pfx)
	run.SetVisible(pfx)
}

func (ctl *control) debugInfo() string {
	var revision string
	bi, _ := debug.ReadBuildInfo()
	for _, bs := range bi.Settings {
		if bs.Key == "vcs.revision" {
			revision = fmt.Sprintf("(%s)", bs.Value)
		}
	}

	var b strings.Builder

	inst := "source"
	if sysinfo.InFlatpak {
		inst = "flatpak"
	}

	info := `* Vinegar: %s %s
* Distro: %s
* Processor: %s
* Kernel: %s
* Wine: %s
* Installation: %s
`

	fmt.Fprintf(&b, info,
		version, revision,
		sysinfo.Distro,
		sysinfo.CPU.Name,
		sysinfo.Kernel,
		ctl.pfx.Version(),
		inst,
	)

	fmt.Fprintln(&b, "* Cards:")
	for i, c := range sysinfo.Cards {
		fmt.Fprintf(&b, "  * Card %d: %s %s %s\n",
			i, c.Driver, filepath.Base(c.Device), c.Path)
	}

	return b.String()
}
