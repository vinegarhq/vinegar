// Package xstray hides stray Roblox Studio windows under X11.
//
// Studio creates small top-level windows titled "RobloxStudio" that are
// meant to stay hidden off-screen. Many X11 window managers move them back
// onto the screen, where they show up as black squares that cannot be
// closed, and steal focus whenever another window closes. Vinegar's virtual
// desktop is unaffected, as Wine's explorer manages the window positions.
package xstray

import (
	"bytes"
	"context"
	"log/slog"
	"math"
	"strings"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const (
	strayName  = "RobloxStudio"
	strayClass = "robloxstudiobeta.exe"
	straySize  = 100 // largest width and height of a stray window
)

type watcher struct {
	conn       *xgb.Conn
	root       xproto.Window
	clientList xproto.Atom

	// Studio windows selected for StructureNotify, as stray
	// windows may be resized after being mapped.
	watched map[xproto.Window]bool
}

// Watch hides stray Studio windows as they appear, until ctx is canceled.
func Watch(ctx context.Context) error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	w := watcher{
		conn:    conn,
		root:    xproto.Setup(conn).DefaultScreen(conn).Root,
		watched: make(map[xproto.Window]bool),
	}

	name := "_NET_CLIENT_LIST"
	atom, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		return err
	}
	w.clientList = atom.Atom

	if err := xproto.ChangeWindowAttributesChecked(conn, w.root,
		xproto.CwEventMask, []uint32{xproto.EventMaskPropertyChange},
	).Check(); err != nil {
		return err
	}

	w.scan()
	for {
		ev, err := conn.WaitForEvent()
		if ev == nil && err == nil { // connection closed
			return nil
		}
		if err != nil {
			// Usually from a window that was destroyed while being checked.
			continue
		}

		switch e := ev.(type) {
		case xproto.PropertyNotifyEvent:
			if e.Window == w.root && e.Atom == w.clientList {
				w.scan()
			}
		case xproto.ConfigureNotifyEvent:
			w.check(e.Window)
		case xproto.DestroyNotifyEvent:
			delete(w.watched, e.Window)
		}
	}
}

// scan checks all windows managed by the window manager.
func (w *watcher) scan() {
	reply, err := xproto.GetProperty(w.conn, false, w.root, w.clientList,
		xproto.AtomWindow, 0, math.MaxUint32).Reply()
	if err != nil {
		slog.Debug("Failed to retrieve X11 client list", "err", err)
		return
	}

	for i := 0; i+4 <= len(reply.Value); i += 4 {
		w.check(xproto.Window(xgb.Get32(reply.Value[i:])))
	}
}

func (w *watcher) check(win xproto.Window) {
	if !w.studio(win) {
		return
	}

	if !w.watched[win] {
		w.watched[win] = true
		xproto.ChangeWindowAttributes(w.conn, win,
			xproto.CwEventMask, []uint32{xproto.EventMaskStructureNotify})
	}

	attrs, err := xproto.GetWindowAttributes(w.conn, win).Reply()
	if err != nil || attrs.MapState == xproto.MapStateUnmapped {
		return
	}

	geom, err := xproto.GetGeometry(w.conn, xproto.Drawable(win)).Reply()
	if err != nil || geom.Width > straySize || geom.Height > straySize {
		return
	}

	slog.Info("Hiding stray Studio window",
		"window", win, "width", geom.Width, "height", geom.Height)
	xproto.UnmapWindow(w.conn, win)
}

// studio reports whether the window is a Studio window titled strayName.
func (w *watcher) studio(win xproto.Window) bool {
	if string(w.property(win, xproto.AtomWmName)) != strayName {
		return false
	}

	for class := range bytes.SplitSeq(w.property(win, xproto.AtomWmClass), []byte{0}) {
		if strings.EqualFold(string(class), strayClass) {
			return true
		}
	}
	return false
}

func (w *watcher) property(win xproto.Window, atom xproto.Atom) []byte {
	reply, err := xproto.GetProperty(w.conn, false, win, atom,
		xproto.GetPropertyTypeAny, 0, 64).Reply()
	if err != nil {
		return nil
	}
	return reply.Value
}
