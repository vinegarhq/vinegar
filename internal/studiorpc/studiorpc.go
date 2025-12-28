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

type StudioRPC struct {
	presence drpc.Activity
	client   *drpc.Client
	rbx      *rbxweb.Client

	place     *rbxweb.PlaceDetail
	thumbnail *rbxweb.Thumbnail
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
	const entry = "[FLog::CloseDataModel] Setting place ID "
	if !strings.HasPrefix(line, entry) {
		return nil
	}

	// Setting place ID $id
	i, err := strconv.ParseInt(strings.TrimPrefix(line, entry), 10, 64)
	if err != nil {
		return err
	}
	place := rbxweb.PlaceID(i)

	slog.Info("studiorpc: Opened Place", "placeid", i)

	s.place, err = s.rbx.GamesV1.GetPlaceDetail(place)
	if err != nil {
		slog.Error("studiorpc: Failed to fetch place detail", "err", err)
	}

	if s.place != nil {
		s.thumbnail, err = s.rbx.ThumbnailsV1.GetGameIcon(s.place.UniverseID, &rbxweb.GameIconOptions{
			Size:        "512x512",
			Rectangular: true,
			Format:      rbxweb.ThumbnailFormatPng,
		})
		if err != nil {
			slog.Error("studiorpc: Failed to fetch game icon", "err", err)
		}
	} else {
		s.thumbnail = nil
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
		s.presence.Assets.LargeImage = ""
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

		if s.thumbnail != nil {
			s.presence.Assets.LargeImage = s.thumbnail.ImageURL
		} else {
			s.presence.Assets.LargeImage = ""
		}
	case "2":
		s.presence.State = "Editing Script"
		s.presence.Assets.SmallImage = "script"
	}

	return s.client.SetActivity(s.presence)
}
