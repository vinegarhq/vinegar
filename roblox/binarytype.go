package roblox

type BinaryType int

const (
	Player BinaryType = iota
	Studio
)

func (bt BinaryType) String() string {
	switch bt {
	case Player:
		return "WindowsPlayer"
	case Studio:
		return "WindowsStudio64"
	default:
		return "unknown"
	}
}
