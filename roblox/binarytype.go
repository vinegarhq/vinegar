package roblox

// BinaryType is a representation of available
// Roblox applications aka. Binaries
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

// BinaryName returns Roblox's internal API name for the
// named BinaryType
//
// Does not support platforms other than Windows.
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

// Executable returns the executable file name for the
// named BinaryType
//
// Does not support platforms other than Windows.
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
