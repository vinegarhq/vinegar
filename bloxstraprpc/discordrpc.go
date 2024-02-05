package bloxstraprpc

import (
	"log/slog"

	"github.com/altfoxie/drpc"
	"github.com/vinegarhq/vinegar/roblox/api"
)

func (a *Activity) Connect() error {
	slog.Info("Connecting to Discord RPC")

	return a.client.Connect()
}

func (a *Activity) Close() error {
	slog.Info("Closing Discord RPC")

	return a.client.Close()
}

// UpdateGamePresence sets the activity based on the current
// game information present in Activity. 'initial' is used
// to fetch game information required for rich presence.
func (a *Activity) UpdateGamePresence(initial bool) error {
	a.presence.Buttons = []drpc.Button{{
		Label: "See game page",
		URL:   "https://www.roblox.com/games/" + a.placeID,
	}}

	if a.server == Public {
		joinurl := "roblox://experiences/start?placeId=" + a.placeID + "&gameInstanceId=" + a.jobID
		a.presence.Buttons = append(a.presence.Buttons, drpc.Button{
			Label: "Join server",
			URL:   joinurl,
		})
	}

	if a.presence.Assets == nil {
		a.presence.Assets = new(drpc.Assets)
	}

	if initial || (a.presence.Details == Reset ||
		a.presence.State == Reset ||
		a.presence.Assets.LargeText == Reset) {
		gd, err := api.GetGameDetails(a.universeID)
		if err != nil {
			return err
		}

		if initial || a.presence.Details == Reset {
			a.presence.Details = "Playing " + gd.Name
		}

		if initial || a.presence.State == Reset {
			a.presence.State = "by " + gd.Creator.Name

			switch a.server {
			case Private:
				a.presence.State = "In a private server"
			case Reserved:
				a.presence.State = "In a reserved server"
			}
		}

		if initial || a.presence.Assets.LargeText == Reset {
			a.presence.Assets.LargeText = gd.Name
		}
	}

	if initial || a.presence.Assets.LargeImage == Reset {
		tn, err := api.GetGameIcon(a.universeID, "PlaceHolder", "512x512", "Png", false)
		if err != nil {
			return err
		}

		a.presence.Assets.LargeImage = tn.ImageURL
	}

	if initial || a.presence.Assets.SmallImage == Reset {
		a.presence.Assets.SmallImage = "roblox"
	}

	if initial || a.presence.Assets.SmallText == Reset {
		a.presence.Assets.SmallText = "Roblox"
	}

	if a.presence.Timestamps == nil {
		a.presence.Timestamps = &drpc.Timestamps{
			Start: a.gameTime,
		}
	}

	slog.Info("Updating Discord Rich Presence", "presence", a.presence)

	return a.client.SetActivity(a.presence)
}
