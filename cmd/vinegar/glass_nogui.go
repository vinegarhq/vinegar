//go:build nogui
// +build nogui

package main

func (b *Binary) Glass(exit <-chan bool) {
	for {
		select {
		case <-exit:
			return
		}
	}
}

func (b *Binary) SendLog(msg string) {
	return
}

func (b *Binary) SendProgress(progress float32) {
	return
}
