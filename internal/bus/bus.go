package bus

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

type SessionBus struct {
	conn   *dbus.Conn
	portal dbus.BusObject
}

func NewSession() *SessionBus {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return &SessionBus{}
	}

	return &SessionBus{
		conn:   conn,
		portal: conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop"),
	}
}

func (s *SessionBus) GamemodeRegister(pid int32) error {
	if s.conn == nil {
		return nil
	}

	call := s.portal.Call("org.freedesktop.portal.GameMode.RegisterGame", 0, pid)
	if call.Err != nil && !errors.Is(call.Err, dbus.ErrMsgNoObject) {
		return call.Err
	}

	return nil
}
