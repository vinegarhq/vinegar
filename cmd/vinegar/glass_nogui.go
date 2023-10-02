//go:build nogui
// +build nogui

package main

func (b *Binary) Glass(exit <-chan bool) {
	for {
		select {
		case <-b.log:
		case <-progress:
		case <-exit:
			return
		}
	}
}
