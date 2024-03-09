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
	"github.com/apprehensions/rbxweb"
	"github.com/vinegarhq/vinegar/richpresence"
)

const reset = "<reset>"

const (
	gameJoinRequestEntry = "[FLog::GameJoinUtil] GameJoinUtil::makePlaceLauncherRequest"
	gameJoiningEntry     = "[FLog::Output] ! Joining game"
	gameJoinReportEntry  = "[FLog::GameJoinLoadTime] Report game_join_loadtime:"
	gameJoinedEntry      = "[FLog::Output] Connection accepted from"
	bloxstrapRPCEntry    = "[FLog::Output] [BloxstrapRPC]"
	gameLeaveEntry       = "[FLog::SingleSurfaceApp] leaveUGCGameInternal"
)

var (
	gameJoinRequestEntryPattern = regexp.MustCompile(`makePlaceLauncherRequest(ForTeleport)?: requestCount: [0-9], url: https:\/\/gamejoin\.roblox\.com\/v1\/([^\s\/]+)`)
	gameJoiningEntryPattern     = regexp.MustCompile(`! Joining game '([0-9a-f\-]{36})'`)
	gameJoinReportEntryPattern  = regexp.MustCompile(`Report game_join_loadtime: placeid:([0-9]+).*universeid:([0-9]+)`)
)

type ServerType int

const (
	Public ServerType = iota
	Private
	Reserved
)

type BloxstrapRPC struct {
	presence drpc.Activity
	client   *drpc.Client

	gameTime    time.Time
	teleporting bool
	server      ServerType

	universeID rbxweb.UniverseID
	placeID    string
	jobID      string
}

func New() *BloxstrapRPC {
	c, _ := drpc.New(richpresence.AppID)
	return &BloxstrapRPC{
		client: c,
	}
}

// Handle implements the BinaryRichPresence interface
func (b *BloxstrapRPC) Handle(line string) error {
	entries := map[string]func(string) error{
		// In order of which it should appear in log file
		gameJoinRequestEntry: b.handleGameJoinRequest,                              // For game join type is private, reserved
		gameJoiningEntry:     b.handleGameJoining,                                  // For JobID (server ID, to join from Discord)
		gameJoinReportEntry:  b.handleGameJoinReport,                               // For PlaceID and UniverseID
		gameJoinedEntry:      func(_ string) error { return b.handleGameJoined() }, // Sets presence and time
		bloxstrapRPCEntry:    b.handleBloxstrapRPC,                                 // BloxstrapRPC
		gameLeaveEntry:       func(_ string) error { return b.handleGameLeave() },  // Clears presence and time
	}

	for e, h := range entries {
		if strings.Contains(line, e) {
			return h(line)
		}
	}

	return nil
}

func (b *BloxstrapRPC) handleGameJoinRequest(line string) error {
	m := gameJoinRequestEntryPattern.FindStringSubmatch(line)
	// There are multiple outputs for makePlaceLauncherRequest
	if len(m) != 3 {
		return fmt.Errorf("log game join request entry is invalid")
	}

	if m[1] == "ForTeleport" {
		b.teleporting = true
	}

	// Keep up to date from upstream Roblox GameJoin API
	b.server = map[string]ServerType{
		"join-private-game":       Private,
		"join-reserved-game":      Reserved,
		"join-game":               Public,
		"join-game-instance":      Public,
		"join-play-together-game": Public,
	}[m[2]]

	slog.Info("Handled GameJoinRequest", "server_type", b.server, "teleporting", b.teleporting)

	return nil
}

func (b *BloxstrapRPC) handleGameJoining(line string) error {
	m := gameJoiningEntryPattern.FindStringSubmatch(line)
	if len(m) != 2 {
		return fmt.Errorf("log game joining entry is invalid")
	}

	b.jobID = m[1]

	slog.Info("Handled GameJoining", "jobid", b.jobID)

	return nil
}

func (b *BloxstrapRPC) handleGameJoinReport(line string) error {
	m := gameJoinReportEntryPattern.FindStringSubmatch(line)
	if len(m) != 3 {
		return fmt.Errorf("log game join report entry is invalid")
	}

	uid, err := strconv.ParseInt(m[2], 10, 64)
	if err != nil {
		return err
	}

	b.placeID = m[1]
	b.universeID = rbxweb.UniverseID(uid)

	slog.Info("Handled GameJoinReport", "universeid", b.universeID, "placeid", b.placeID)

	return nil
}

func (b *BloxstrapRPC) handleGameJoined() error {
	if !b.teleporting {
		b.gameTime = time.Now()
	}

	b.teleporting = false

	slog.Info("Handled GameJoined", "time", b.gameTime)

	return b.UpdateGamePresence(true)
}

func (b *BloxstrapRPC) handleBloxstrapRPC(line string) error {
	m, err := ParseMessage(line)
	if err != nil {
		return fmt.Errorf("parse bloxstraprpc message: %w", err)
	}
	m.ApplyRichPresence(&b.presence)

	slog.Info("Handled BloxstrapRPC", "message", m)

	return b.UpdateGamePresence(false)
}

func (b *BloxstrapRPC) handleGameLeave() error {
	b.presence = drpc.Activity{}
	b.gameTime = time.Time{}
	b.teleporting = false
	b.server = Public
	b.universeID = rbxweb.UniverseID(0)
	b.placeID = ""
	b.jobID = ""

	slog.Info("Handled GameLeave")

	return b.client.SetActivity(b.presence)
}
