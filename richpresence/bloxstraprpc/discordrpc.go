package bloxstraprpc

import (
	"log/slog"

	"github.com/altfoxie/drpc"
	"github.com/apprehensions/rbxweb/games"
	"github.com/apprehensions/rbxweb/thumbnails"
)

// UpdateGamePresence sets the activity based on the current
// game information present in BloxstrapRPC. 'initial' is used
// to fetch game information required for rich presence, regardless
// if the BloxstrapRPC properties have been set for resetting.
//
// UpdateGamePresence is called by Handle whenever needed.
func (b *BloxstrapRPC) UpdateGamePresence(initial bool) error {
	b.presence.Buttons = []drpc.Button{{
		Label: "See game page",
		URL:   "https://www.roblox.com/games/" + b.placeID,
	}}

	if b.server == Public {
		joinurl := "roblox://experiences/start?placeId=" + b.placeID + "&gameInstanceId=" + b.jobID
		b.presence.Buttons = append(b.presence.Buttons, drpc.Button{
			Label: "Join server",
			URL:   joinurl,
		})
	}

	if b.presence.Assets == nil {
		b.presence.Assets = new(drpc.Assets)
	}

	if initial || (b.presence.Details == reset ||
		b.presence.State == reset ||
		b.presence.Assets.LargeText == reset) {
		gd, err := games.GetGameDetail(b.universeID)
		if err != nil {
			return err
		}

		if initial || b.presence.Details == reset {
			b.presence.Details = "Playing " + gd.Name
		}

		if initial || b.presence.State == reset {
			b.presence.State = "by " + gd.Creator.Name

			switch b.server {
			case Private:
				b.presence.State = "In a private server"
			case Reserved:
				b.presence.State = "In a reserved server"
			}
		}

		if initial || b.presence.Assets.LargeText == reset {
			b.presence.Assets.LargeText = gd.Name
		}
	}

	if initial || b.presence.Assets.LargeImage == reset {
		tn, err := thumbnails.GetGameIcon(b.universeID,
			thumbnails.PlaceHolder, "512x512", thumbnails.Png, false)
		if err != nil {
			return err
		}

		b.presence.Assets.LargeImage = tn.ImageURL
	}

	if initial || b.presence.Assets.SmallImage == reset {
		b.presence.Assets.SmallImage = "roblox"
	}

	if initial || b.presence.Assets.SmallText == reset {
		b.presence.Assets.SmallText = "Roblox"
	}

	if b.presence.Timestamps == nil {
		b.presence.Timestamps = &drpc.Timestamps{
			Start: b.gameTime,
		}
	}

	slog.Info("Updating Discord Rich Presence", "presence", b.presence)

	return b.client.SetActivity(b.presence)
}
