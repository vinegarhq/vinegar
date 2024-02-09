//go:build windows
// +build windows

package main

import (
	"log"
	"unsafe"
	"time"

	"golang.org/x/sys/windows"
)

func main() {
	log.SetPrefix("robloxmutexer: ")
	log.SetFlags(log.Lmsgprefix | log.LstdFlags)

	if err := lock(); err != nil {
		log.Fatal(err)
	}
}

func lock() error {
	name, err := windows.UTF16PtrFromString("ROBLOX_singletonMutex")
	if err != nil {
		return err
	}

	mutex, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(mutex)

	_, err = windows.WaitForSingleObject(mutex, 0)
	if err != nil {
		return err
	}

	log.Println("Singleton mutex locked, waiting for no roblox processes")

	for {
		time.Sleep(5 * time.Second)
		r, err := robloxRunning()
		if err != nil {
			return err
		}

		if !r {
			log.Println("No roblox processes found, freeing mutex")
			break
		}
	}

	return nil
}

func robloxRunning() (bool, error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(snap)

	pe := windows.ProcessEntry32{
		Size: (uint32)(unsafe.Sizeof(windows.ProcessEntry32{})),
	}

	for {
		if err := windows.Process32Next(snap, &pe); err != nil {
			// Reached end
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return false, err
		}
		if windows.UTF16ToString(pe.ExeFile[:]) == "RobloxPlayerBeta.exe" {
			return true, nil
		}
	}

	return false, nil
}
