//go:build windows
// +build windows

package main

import (
	"os"
	"log/slog"

	"golang.org/x/sys/windows"
)

const mutex = "ROBLOX_singletonMutex"

func main() {
	if err := lock(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("Locked", "mutex", mutex)

	_, _ = windows.WaitForSingleObject(windows.CurrentProcess(), windows.INFINITE)
}

func lock() error {
	name, err := windows.UTF16PtrFromString(mutex)
	if err != nil {
		return err
	}

	handle, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		return err
	}

	_, err = windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return err
	}

	return nil
}
