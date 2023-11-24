package bus

import (
	"errors"
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
)

func (s *SessionBus) GamemodeRegister(pid int) (bool, error) {
	if s.conn == nil {
		return false, nil
	}

	fmt.Println(os.Getpid())

	call := s.portal.Call("org.freedesktop.portal.GameMode.RegisterGame", 0, int32(pid))
	if call.Err != nil {
		//Transparently handle missing portal
		if !errors.Is(call.Err, dbus.ErrMsgNoObject) {
			return false, nil
		}
		return false, call.Err
	}

	response := call.Body[0].(int32)

	return response > 0, nil
}