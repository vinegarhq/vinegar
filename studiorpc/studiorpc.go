// Package studiorpc implements basic Roblox Studio rich presence.
package studiorpc

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/altfoxie/drpc"
	"github.com/apprehensions/rbxweb"
)

const appID = "1159891020956323923"

const (
	gameOpenEntry  = "[FLog::LifecycleManager] Entered PlaceSessionScope:"
	gameCloseEntry = "[FLog::LifecycleManager] Exited PlaceSessionScope:"
)

var gameOpenEntryPattern = regexp.MustCompile(`Entered PlaceSessionScope:'([0-9]+)'`)

type StudioRPC struct {
	presence drpc.Activity
	client   *drpc.Client

	placeID rbxweb.PlaceID
}

func New() *StudioRPC {
	c, _ := drpc.New("1159891020956323923")
	return &StudioRPC{
		client: c,
	}
}

// Handle implements the BinaryRichPresence interface
func (s *StudioRPC) Handle(line string) error {
	if strings.Contains(line, gameOpenEntry) {
		return s.handleGameOpen(line)
	} else if strings.Contains(line, gameCloseEntry) {
		return s.handleGameClose()
	}

	return nil
}

func (s *StudioRPC) handleGameOpen(line string) error {
	m := gameOpenEntryPattern.FindStringSubmatch(line)
	if len(m) != 2 {
		return fmt.Errorf("log game join report entry is invalid")
	}

	pid, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return err
	}

	s.placeID = rbxweb.PlaceID(pid)

	slog.Info("Handled GameOpen", "placeid", s.placeID)

	return s.UpdateGamePresence()
}

func (s *StudioRPC) handleGameClose() error {
	s.presence = drpc.Activity{}
	s.placeID = rbxweb.PlaceID(0)

	slog.Info("Handled GameClose")

	return s.client.SetActivity(s.presence)
}
