// Package bloxstraprpc implements the BloxstrapRPC protocol.
//
// This package remains undocumented as it is modeled after Bloxstrap's
// implementation protocol.
package bloxstraprpc

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hugolgst/rich-go/client"
)

const (
	GameJoiningEntry               = "[FLog::Output] ! Joining game"
	GameJoiningPrivateServerEntry  = "[FLog::GameJoinUtil] GameJoinUtil::joinGamePostPrivateServer"
	GameJoiningReservedServerEntry = "[FLog::GameJoinUtil] GameJoinUtil::initiateTeleportToReservedServer"
	GameJoiningUDMUXEntry          = "[FLog::Network] UDMUX Address = "
	GameJoinedEntry                = "[FLog::Network] serverId:"
	GameDisconnectedEntry          = "[FLog::Network] Time to disconnect replication data:"
	GameTeleportingEntry           = "[FLog::SingleSurfaceApp] initiateTeleport"
	GameMessageEntry               = "[FLog::Output] [BloxstrapRPC]"
)

var (
	GameJoiningEntryPattern = regexp.MustCompile(`! Joining game '([0-9a-f\-]{36})' place ([0-9]+) at ([0-9\.]+)`)
	GameJoiningUDMUXPattern = regexp.MustCompile(`UDMUX Address = ([0-9\.]+), Port = [0-9]+ \| RCC Server Address = ([0-9\.]+), Port = [0-9]+`)
	GameJoinedEntryPattern  = regexp.MustCompile(`serverId: ([0-9\.]+)\|[0-9]+`)
)

type ServerType int

const (
	Public ServerType = iota
	Private
	Reserved
)

type Activity struct {
	presence            client.Activity
	timeStartedUniverse time.Time
	currentUniverseID   string

	ingame     bool
	teleported bool
	server     ServerType
	placeID    string
	jobID      string
	mac        string

	teleport         bool
	reservedteleport bool
}

func (a *Activity) HandleRobloxLog(line string) error {
	if !a.ingame && a.placeID == "" {
		if strings.Contains(line, GameJoiningPrivateServerEntry) {
			a.server = Private
			return nil
		}

		if strings.Contains(line, GameJoiningEntry) {
			a.handleGameJoining(line)
			return nil
		}
	}

	if !a.ingame && a.placeID != "" {
		if strings.Contains(line, GameJoiningUDMUXEntry) {
			a.handleUDMUX(line)
			return nil
		}

		if strings.Contains(line, GameJoinedEntry) {
			a.handleGameJoined(line)
			return a.SetCurrentGame()
		}
	}

	if a.ingame && a.placeID != "" {
		if strings.Contains(line, GameDisconnectedEntry) {
			log.Printf("Disconnected From Game (%s/%s/%s)", a.placeID, a.jobID, a.mac)
			a.Clear()
			return a.SetCurrentGame()
		}

		if strings.Contains(line, GameTeleportingEntry) {
			log.Printf("Teleporting to server (%s/%s/%s)", a.placeID, a.jobID, a.mac)
			a.teleport = true
			return nil
		}

		if a.teleport && strings.Contains(line, GameJoiningReservedServerEntry) {
			log.Printf("Teleporting to reserved server")
			a.reservedteleport = true
			return nil
		}

		if strings.Contains(line, GameMessageEntry) {
			m, err := ParseMessage(line)
			if err != nil {
				return err
			}

			a.ProcessMessage(&m)
			return a.UpdatePresence()
		}
	}

	return nil
}

func (a *Activity) handleUDMUX(line string) {
	m := GameJoiningUDMUXPattern.FindStringSubmatch(line)
	if len(m) != 3 || m[2] != a.mac {
		return
	}

	a.mac = m[1]
	log.Printf("Got game join UDMUX: %s", a.mac)
}

func (a *Activity) handleGameJoining(line string) {
	m := GameJoiningEntryPattern.FindStringSubmatch(line)
	if len(m) != 4 {
		return
	}

	a.ingame = false
	a.jobID = m[1]
	a.placeID = m[2]
	a.mac = m[3]

	if a.teleport {
		a.teleported = true
		a.teleport = false
	}

	if a.reservedteleport {
		a.server = Reserved
		a.reservedteleport = false
	}

	log.Printf("Joining Game (%s/%s/%s)", a.jobID, a.placeID, a.mac)
}

func (a *Activity) handleGameJoined(line string) {
	m := GameJoinedEntryPattern.FindStringSubmatch(line)
	if len(m) != 2 || m[1] != a.mac {
		return
	}

	a.ingame = true
	log.Printf("Joined Game (%s/%s/%s)", a.placeID, a.jobID, a.mac)
	// handle rpc
}

func (a *Activity) Clear() {
	a.teleported = false
	a.ingame = false
	a.placeID = ""
	a.jobID = ""
	a.mac = ""
	a.server = Public
	a.presence = client.Activity{}
}
