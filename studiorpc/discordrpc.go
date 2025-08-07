package studiorpc

import (
	"log/slog"
	"time"

	"github.com/altfoxie/drpc"
)

// UpdateGamePresence sets the activity based on the current
// game information present in StudioRPC.
//
// UpdateGamePresence is called by Handle whenever needed.
func (s *StudioRPC) UpdateGamePresence() error {
	details := ""

	pd, err := s.rbx.GamesV1.GetPlaceDetail(s.placeID)
	// Sometimes the game itself is actually just a template, and is not owned by the
	// user, which is why details won't be fetched.
	if err != nil {
		slog.Error("Failed to fetch place details", "placeid", s.placeID, "error", err)
	} else {
		details = "Workspace " + pd.Name
	}

	s.presence = drpc.Activity{
		State:   "Developing",
		Details: details,
		Assets: &drpc.Assets{
			LargeImage: "studio",
			LargeText:  "studio",
		},
		Timestamps: &drpc.Timestamps{
			Start: time.Now(),
		},
	}

	slog.Info("Updating Discord Rich Presence", "presence", s.presence)

	return s.client.SetActivity(s.presence)
}
