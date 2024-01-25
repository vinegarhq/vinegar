package bloxstraprpc

import (
	"log"
	"strconv"
	"time"

	"github.com/altfoxie/drpc"
	"github.com/vinegarhq/vinegar/roblox/api"
)

// This is Bloxstrap's Discord RPC application ID.
const RPCAppID = "1005469189907173486"

func (a *Activity) Connect() error {
	log.Println("Connecting to Discord RPC")

	return a.client.Connect()
}

func (a *Activity) Close() error {
	log.Println("Closing Discord RPC")

	return a.client.Close()
}

func (a *Activity) SetCurrentGame() error {
	if err := a.SetPresence(); err != nil {
		return err
	}

	return a.UpdatePresence()
}

func (a *Activity) SetPresence() error {
	var status string

	gd, err := api.GetGameDetails(a.universeID)
	if err != nil {
		return err
	}
	log.Println("Got Game details")

	tn, err := api.GetGameIcon(a.universeID, "PlaceHolder", "512x512", "Png", false)
	if err != nil {
		return err
	}
	log.Printf("Got Universe thumbnail as %s", tn.ImageURL)

	buttons := []drpc.Button{{
		Label: "See game page",
		URL:   "https://www.roblox.com/games/" + a.placeID,
	}}

	joinurl := "roblox://experiences/start?placeId=" + a.placeID + "&gameInstanceId=" + a.jobID
	switch a.server {
	case Public:
		status = "by " + gd.Creator.Name
		buttons = append(buttons, drpc.Button{
			Label: "Join server",
			URL:   joinurl,
		})
	case Private:
		status = "In a private server"
	case Reserved:
		status = "In a reserved server"
	}

	a.presence = drpc.Activity{
		State:   status,
		Details: "Playing " + gd.Name,
		Assets: &drpc.Assets{
			LargeImage: tn.ImageURL,
			LargeText:  gd.Name,
			SmallImage: "roblox",
			SmallText:  "Roblox",
		},
		Timestamps: &drpc.Timestamps{
			Start: a.gameTime,
		},
		Buttons: buttons,
	}

	return nil
}

func (a *Activity) ProcessMessage(m *Message) {
	if m.Command != "SetRichPresence" {
		return
	}

	if m.Data.Details != "" {
		a.presence.Details = m.Data.Details
	}

	if m.Data.State != "" {
		a.presence.State = m.Data.State
	}

	if a.presence.Timestamps != nil {
		if m.TimestampStart == 0 {
			a.presence.Timestamps = nil
		} else {
			a.presence.Timestamps.Start = time.UnixMilli(m.TimestampStart)
		}
		if m.TimestampEnd == 0 {
			a.presence.Timestamps = nil
		} else {
			a.presence.Timestamps.End = time.UnixMilli(m.TimestampEnd)
		}
	}

	if m.SmallImage.Clear {
		a.presence.Assets.SmallImage = ""
	}

	if m.SmallImage.AssetID != 0 {
		a.presence.Assets.SmallImage = "https://assetdelivery.roblox.com/v1/asset/?id" +
			strconv.FormatInt(m.SmallImage.AssetID, 10)
	}

	if m.LargeImage.Clear {
		a.presence.Assets.LargeImage = ""
	}

	if m.LargeImage.AssetID != 0 {
		a.presence.Assets.LargeImage = "https://assetdelivery.roblox.com/v1/asset/?id" +
			strconv.FormatInt(m.LargeImage.AssetID, 10)
	}
}

func (a *Activity) UpdatePresence() error {
	log.Printf("Updating presence: %+v", a.presence)
	return a.client.SetActivity(a.presence)
}
