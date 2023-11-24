package bus

import (
	"github.com/godbus/dbus/v5"
)

type SessionBus struct{
	conn   *dbus.Conn
	portal dbus.BusObject
}

func NewSession() *SessionBus {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return &SessionBus{}
	}

	session := &SessionBus{
		conn: conn,
		portal: conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop"),
	}
}
