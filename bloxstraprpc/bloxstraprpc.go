// Package bloxstraprpc implements the BloxstrapRPC protocol.
//
// This package remains undocumented as it is modeled after Bloxstrap's
// implementation protocol.
package bloxstraprpc

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/altfoxie/drpc"
)

const Reset = "<reset>"

const (
	GameJoinRequestEntry = "[FLog::GameJoinUtil] GameJoinUtil::makePlaceLauncherRequest"
	GameJoiningEntry     = "[FLog::Output] ! Joining game"
	GameJoinReportEntry  = "[FLog::GameJoinLoadTime] Report game_join_loadtime:"
	GameJoinedEntry      = "[FLog::Output] Connection accepted from"
	BloxstrapRPCEntry    = "[FLog::Output] [BloxstrapRPC]"
	GameLeaveEntry       = "[FLog::SingleSurfaceApp] leaveUGCGameInternal"
)

var (
	GameJoinRequestEntryPattern = regexp.MustCompile(`makePlaceLauncherRequest(ForTeleport)?: requestCount: [0-9], url: https:\/\/gamejoin\.roblox\.com\/v1\/([^\s\/]+)`)
	GameJoiningEntryPattern     = regexp.MustCompile(`! Joining game '([0-9a-f\-]{36})'`)
	GameJoinReportEntryPattern  = regexp.MustCompile(`Report game_join_loadtime: placeid:([0-9]+).*universeid:([0-9]+)`)
)

type ServerType int

const (
	Public ServerType = iota
	Private
	Reserved
)

type Activity struct {
	presence drpc.Activity
	client   *drpc.Client

	gameTime    time.Time
	teleporting bool
	server      ServerType

	universeID string
	placeID    string
	jobID      string
}

func New() Activity {
	c, _ := drpc.New("1159891020956323923")
	return Activity{
		client: c,
	}
}

func (a *Activity) HandleRobloxLog(line string) error {
	entries := map[string]func(string) error{
		// In order of which it should appear in log file
		GameJoinRequestEntry: a.handleGameJoinRequest,                              // For game join type is private, reserved
		GameJoiningEntry:     a.handleGameJoining,                                  // For JobID (server ID, to join from Discord)
		GameJoinReportEntry:  a.handleGameJoinReport,                               // For PlaceID and UniverseID
		GameJoinedEntry:      func(_ string) error { return a.handleGameJoined() }, // Sets presence and time
		BloxstrapRPCEntry:    a.handleGameMessage,                                  // BloxstrapRPC
		GameLeaveEntry:       func(_ string) error { return a.handleGameLeave() },  // Clears presence and time
	}

	for e, h := range entries {
		if strings.Contains(line, e) {
			return h(line)
		}
	}

	return nil
}

func (a *Activity) handleGameJoinRequest(line string) error {
	m := GameJoinRequestEntryPattern.FindStringSubmatch(line)
	// There are multiple outputs for makePlaceLauncherRequest
	if len(m) != 3 {
		return nil
	}

	if m[1] == "ForTeleport" {
		a.teleporting = true
	}

	// Keep up to date from upstream Roblox GameJoin API
	a.server = map[string]ServerType{
		"join-private-game":       Private,
		"join-reserved-game":      Reserved,
		"join-game":               Public,
		"join-game-instance":      Public,
		"join-play-together-game": Public,
	}[m[2]]

	log.Printf("Got Game type %d teleporting %t!", a.server, a.teleporting)
	return nil
}

func (a *Activity) handleGameJoining(line string) error {
	m := GameJoiningEntryPattern.FindStringSubmatch(line)
	if len(m) != 2 {
		return fmt.Errorf("log game joining entry is invalid!")
	}

	a.jobID = m[1]

	log.Printf("Got Job %s!", a.jobID)
	return nil
}

func (a *Activity) handleGameJoinReport(line string) error {
	m := GameJoinReportEntryPattern.FindStringSubmatch(line)
	if len(m) != 3 {
		return fmt.Errorf("log game join report entry is invalid!")
	}

	a.placeID = m[1]
	a.universeID = m[2]

	log.Printf("Got Universe %s Place %s!", a.universeID, a.placeID)
	return nil
}

func (a *Activity) handleGameJoined() error {
	if !a.teleporting {
		log.Println("Updating time!")
		a.gameTime = time.Now()
	}

	a.teleporting = false

	log.Println("Game Joined!")
	return a.UpdateGamePresence(true)
}

func (a *Activity) handleGameMessage(line string) error {
	m, err := NewMessage(line)
	if err != nil {
		return fmt.Errorf("parse bloxstraprpc message: %w", err)
	}
	m.ApplyRichPresence(&a.presence)

	return a.UpdateGamePresence(false)
}

func (a *Activity) handleGameLeave() error {
	log.Println("Left game, clearing presence!")

	a.presence = drpc.Activity{}
	a.gameTime = time.Time{}
	a.teleporting = false
	a.server = Public
	a.universeID = ""
	a.placeID = ""
	a.jobID = ""

	return a.client.SetActivity(a.presence)
}
