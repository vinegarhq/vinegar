package roblox

import (
	"testing"
)

func TestProperBinaryExecutableNames(t *testing.T) {
	if Player.Executable() != "RobloxPlayerBeta.exe" {
		t.Fatal("Player's program name is not as specified")
	}
	if Studio.Executable() != "RobloxStudioBeta.exe" {
		t.Fatal("Studio's program name is not as specified")
	}
	if BinaryType(-1).Executable() != "unknown" {
		t.FailNow()
	}
}
