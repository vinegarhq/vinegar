package roblox

type BinaryType int

const (
	Player BinaryType = iota
	Studio
)

func (bt BinaryType) String() string {
	switch bt {
	case Player:
		return "Player"
	case Studio:
		return "Studio"
	default:
		return "unknown"
	}
}

func (bt BinaryType) BinaryName() string {
	switch bt {
	case Player:
		return "WindowsPlayer"
	case Studio:
		return "WindowsStudio64"
	default:
		return "unknown"
	}
}

func (bt BinaryType) Executable() string {
	switch bt {
	case Player:
		return "RobloxPlayerBeta.exe"
	case Studio:
		return "RobloxStudioBeta.exe"
	default:
		return "unknown"
	}
}
