// Package studiorpc implements basic Roblox Studio rich presence.
package studiorpc

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/altfoxie/drpc"
	"github.com/sewnie/rbxweb"
)

const appID = "1159891020956323923"

type StudioRPC struct {
	presence drpc.Activity
	client   *drpc.Client
	rbx      *rbxweb.Client

	place *rbxweb.PlaceDetail
}

func New(c *rbxweb.Client) *StudioRPC {
	return &StudioRPC{
		client: drpc.New("1159891020956323923"),
		rbx:    c,
		presence: drpc.Activity{
			Assets: &drpc.Assets{
				LargeImage: "studio",
				LargeText:  "Vinegar",
			},
		},
	}
}

func (s *StudioRPC) Handle(line string) error {
	for _, handler := range []func(string) error{
		s.handleOpen,
		s.handleEdit,
	} {
		if err := handler(line); err != nil {
			return err
		}
	}

	return nil
}

func (s *StudioRPC) handleOpen(line string) error {
	const entry = "[FLog::StudioKeyEvents] open place (identifier = "
	if !strings.HasPrefix(line, entry) {
		return nil
	}

	// open place (identifier = $id) [start]
	i, err := strconv.ParseInt(line[len(entry):len(line)-len(") [start]")], 10, 64)
	if err != nil {
		return err
	}
	place := rbxweb.PlaceID(i)

	slog.Info("studiorpc: Opened Place", "placeid", i)

	s.place, err = s.rbx.GamesV1.GetPlaceDetail(place)
	if err != nil {
		slog.Error("studiorpc: Failed to fetch place detail", "err", err)
	}

	s.presence.Timestamps = &drpc.Timestamps{
		Start: time.Now(),
	}

	return nil
}

func (s *StudioRPC) handleEdit(line string) error {
	const entry = "[FLog::RobloxDocManager] Setting current doc to type "
	if !strings.HasPrefix(line, entry) {
		return nil
	}
	slog.Info("studiorpc: Changed main view")
	switch line[len(entry):] {
	case "-1":
		s.place = nil
		s.presence.Details = "Idling"
		s.presence.State = ""
		s.presence.Assets.SmallImage = ""
		s.presence.Timestamps = &drpc.Timestamps{
			Start: time.Now(),
		}
	case "0":
		s.presence.Details = "Developing a place"
		s.presence.State = ""
		s.presence.Assets.SmallImage = "place"
		if s.place != nil {
			s.presence.Details = "Developing " + s.place.Name
		}
	case "2":
		s.presence.State = "Editing Script"
		s.presence.Assets.SmallImage = "script"
	}

	return s.client.SetActivity(s.presence)
}
