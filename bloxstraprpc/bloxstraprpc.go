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

const (
	GameJoinReportEntry = "[FLog::GameJoinLoadTime] Report game_join_loadtime:"
	GameJoiningEntry    = "[FLog::Output] ! Joining game"
	GameJoinedEntry     = "[FLog::Output] Connection accepted from"

	GameJoinPrivateServerEntry   = "[FLog::GameJoinUtil] GameJoinUtil::joinGamePostPrivateServer"
	GameTeleportingReservedEntry = "[FLog::GameJoinUtil] GameJoinUtil::initiateTeleportToReservedServer"

	GameDisconnectedEntry = "[FLog::Network] Time to disconnect replication data:"
	GameTeleportingEntry  = "[FLog::SingleSurfaceApp] initiateTeleport"
	GameMessageEntry      = "[FLog::Output] [BloxstrapRPC]"
)

var (
	GameJoinReportEntryPattern = regexp.MustCompile(`Report game_join_loadtime: placeid:([0-9]+).*universeid:([0-9]+)`)
	GameJoiningEntryPattern    = regexp.MustCompile(`! Joining game '([0-9a-f\-]{36})'`) // for JobID
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
	c, _ := drpc.New(RPCAppID)
	return Activity{
		client: c,
	}
}

func (a *Activity) HandleRobloxLog(line string) error {
	entries := map[string]func(string) error{
		GameJoiningEntry:      a.handleGameJoining,
		GameJoinReportEntry:   a.handleGameJoinReport,
		GameJoinedEntry:       func(_ string) error { return a.handleGameJoined() },
		GameMessageEntry:      a.handleGameMessage,
		GameDisconnectedEntry: func(_ string) error { return a.handleGameDisconnect() },

		GameTeleportingEntry: func(_ string) error {
			log.Println("Got Teleporting Game!")
			a.teleporting = true
			return nil
		},
		GameJoinPrivateServerEntry: func(_ string) error {
			log.Println("Got Private Game!")
			a.server = Private
			return nil
		},
		GameTeleportingReservedEntry: func(_ string) error {
			log.Println("Got Teleporting Reserved Game!")
			a.server = Reserved
			a.teleporting = true
			return nil
		},
	}

	for e, h := range entries {
		if strings.Contains(line, e) {
			return h(line)
		}
	}

	return nil
}

func (a *Activity) handleGameDisconnect() error {
	log.Println("Disconnected from game!")
	a.teleporting = false
	a.placeID = ""
	a.jobID = ""
	a.server = Public
	a.presence = drpc.Activity{}

	return a.SetCurrentGame()
}

func (a *Activity) handleGameMessage(line string) error {
	m, err := ParseMessage(line)
	if err != nil {
		return fmt.Errorf("parse bloxstraprpc message: %w", err)
	}
	a.ProcessMessage(&m)

	return a.UpdatePresence()
}

func (a *Activity) handleGameJoinReport(line string) error {
	m := GameJoinReportEntryPattern.FindStringSubmatch(line)
	if len(m) != 3 {
		return fmt.Errorf("game join report entry is invalid!")
	}

	a.placeID = m[1]
	a.universeID = m[2]

	log.Printf("Got Universe %s Place %s!", a.universeID, a.placeID)
	return nil
}

func (a *Activity) handleGameJoining(line string) error {
	m := GameJoiningEntryPattern.FindStringSubmatch(line)
	if len(m) != 2 {
		return fmt.Errorf("game joining entry is invalid!")
	}

	a.jobID = m[1]

	log.Printf("Got Job %s!", a.jobID)
	return nil
}

func (a *Activity) handleGameJoined() error {
	if !a.teleporting {
		log.Println("Updating time!")
		a.gameTime = time.Now()
	}

	a.teleporting = false

	log.Println("Game Joined!")
	return a.SetCurrentGame()
}
