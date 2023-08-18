package roblox

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
