//go:build nogui
// +build nogui

package main

func (b *Binary) Glass(exit <-chan bool) {
	EmptyGlass(exit)
}
