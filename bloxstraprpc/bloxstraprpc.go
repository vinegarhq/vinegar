// Package bloxstraprpc implements the BloxstrapRPC protocol.
//
// For more information regarding the protocol, view [Bloxstrap's BloxstrapRPC wiki page]
//
// [Bloxstrap's BloxstrapRPC wiki page]: https://github.com/pizzaboxer/bloxstrap/wiki/Integrating-Bloxstrap-functionality-into-your-game
package bloxstraprpc

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/altfoxie/drpc"
	"github.com/apprehensions/rbxweb/games"
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

	universeID games.UniverseID
	placeID    string
	jobID      string
}

func New() Activity {
	c, _ := drpc.New("1159891020956323923")
	return Activity{
		client: c,
	}
}

// HandleRobloxLog handles the given Roblox log entry, to set data
// and call functions based on the log entry, declared as *Entry(Pattern) constants.
func (a *Activity) HandleRobloxLog(line string) error {
	entries := map[string]func(string) error{
		// In order of which it should appear in log file
		GameJoinRequestEntry: a.handleGameJoinRequest,                              // For game join type is private, reserved
		GameJoiningEntry:     a.handleGameJoining,                                  // For JobID (server ID, to join from Discord)
		GameJoinReportEntry:  a.handleGameJoinReport,                               // For PlaceID and UniverseID
		GameJoinedEntry:      func(_ string) error { return a.handleGameJoined() }, // Sets presence and time
		BloxstrapRPCEntry:    a.handleBloxstrapRPC,                                 // BloxstrapRPC
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
		return fmt.Errorf("log game join request entry is invalid")
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

	slog.Info("Handled GameJoinRequest", "server_type", a.server, "teleporting", a.teleporting)

	return nil
}

func (a *Activity) handleGameJoining(line string) error {
	m := GameJoiningEntryPattern.FindStringSubmatch(line)
	if len(m) != 2 {
		return fmt.Errorf("log game joining entry is invalid")
	}

	a.jobID = m[1]

	slog.Info("Handled GameJoining", "jobid", a.jobID)

	return nil
}

func (a *Activity) handleGameJoinReport(line string) error {
	m := GameJoinReportEntryPattern.FindStringSubmatch(line)
	if len(m) != 3 {
		return fmt.Errorf("log game join report entry is invalid")
	}

	uid, err := strconv.ParseInt(m[2], 10, 64)
	if err != nil {
		return err
	}

	a.placeID = m[1]
	a.universeID = games.UniverseID(uid)

	slog.Info("Handled GameJoinReport", "universeid", a.universeID, "placeid", a.placeID)

	return nil
}

func (a *Activity) handleGameJoined() error {
	if !a.teleporting {
		a.gameTime = time.Now()
	}

	a.teleporting = false

	slog.Info("Handled GameJoined", "time", a.gameTime)

	return a.UpdateGamePresence(true)
}

func (a *Activity) handleBloxstrapRPC(line string) error {
	m, err := NewMessage(line)
	if err != nil {
		return fmt.Errorf("parse bloxstraprpc message: %w", err)
	}
	m.ApplyRichPresence(&a.presence)

	slog.Info("Handled BloxstrapRPC", "message", m)

	return a.UpdateGamePresence(false)
}

func (a *Activity) handleGameLeave() error {
	a.presence = drpc.Activity{}
	a.gameTime = time.Time{}
	a.teleporting = false
	a.server = Public
	a.universeID = games.UniverseID(0)
	a.placeID = ""
	a.jobID = ""

	slog.Info("Handled GameLeave")

	return a.client.SetActivity(a.presence)
}
