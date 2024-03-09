// Package richpresence provides interfaces against Roblox Binaries for Discord Rich Presence
package richpresence

const AppID = "1159891020956323923"

type BinaryRichPresence interface {
	Handle(string) error // Log entry handler
}
