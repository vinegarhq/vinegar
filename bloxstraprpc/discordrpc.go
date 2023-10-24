package bloxstraprpc

import (
	"log"
	"strconv"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/vinegarhq/vinegar/roblox/api"
)

// This is Bloxstrap's Discord RPC application ID.
const RPCAppID = "1005469189907173486"

func Login() error {
	log.Println("Authenticating Discord RPC")
	return client.Login(RPCAppID)
}

func Logout() {
	log.Println("Deauthenticating Discord RPC")
	client.Logout()
}

func (a *Activity) SetCurrentGame() error {
	if !a.ingame {
		log.Println("Not in game, clearing presence")
		a.presence = client.Activity{}
	} else {
		if err := a.SetPresence(); err != nil {
			return err
		}
	}

	return a.UpdatePresence()
}

func (a *Activity) SetPresence() error {
	var status string
	log.Printf("Setting presence for Place ID %s", a.placeID)

	uid, err := api.GetUniverseID(a.placeID)
	if err != nil {
		return err
	}
	log.Printf("Got Universe ID as %s", uid)

	if !a.teleported || uid != a.currentUniverseID {
		a.timeStartedUniverse = time.Now()
	}

	a.currentUniverseID = uid

	gd, err := api.GetGameDetails(uid)
	if err != nil {
		return err
	}
	log.Println("Got Game details")

	tn, err := api.GetGameIcon(uid, "PlaceHolder", "512x512", "Png", false)
	if err != nil {
		return err
	}
	log.Printf("Got Universe thumbnail as %s", tn.ImageURL)

	switch a.server {
	case Public:
		status = "by " + gd.Creator.Name
	case Private:
		status = "In a private server"
	case Reserved:
		status = "In a reserved server"
	}

	a.presence = client.Activity{
		State:      status,
		Details:    "Playing " + gd.Name,
		LargeImage: tn.ImageURL,
		LargeText:  gd.Name,
		SmallImage: "roblox",
		SmallText:  "Roblox",
		Timestamps: &client.Timestamps{
			Start: &a.timeStartedUniverse,
		},
		Buttons: []*client.Button{
			{
				Label: "See game page",
				Url:   "https://www.roblox.com/games/" + a.placeID,
			},
		},
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
			a.presence.Timestamps.Start = nil
		} else {
			ts := time.UnixMilli(m.TimestampStart)
			a.presence.Timestamps.Start = &ts
		}
	}

	if a.presence.Timestamps != nil {
		if m.TimestampEnd == 0 {
			a.presence.Timestamps.End = nil
		} else {
			te := time.UnixMilli(m.TimestampEnd)
			a.presence.Timestamps.End = &te
		}
	}

	if m.SmallImage.Clear {
		a.presence.SmallImage = ""
	}

	if m.SmallImage.AssetID != 0 {
		a.presence.SmallImage = "https://assetdelivery.roblox.com/v1/asset/?id" +
			strconv.FormatInt(m.SmallImage.AssetID, 10)
	}

	if m.LargeImage.Clear {
		a.presence.LargeImage = ""
	}

	if m.LargeImage.AssetID != 0 {
		a.presence.LargeImage = "https://assetdelivery.roblox.com/v1/asset/?id" +
			strconv.FormatInt(m.LargeImage.AssetID, 10)
	}
}

func (a *Activity) UpdatePresence() error {
	//	if a.presence == client.Activity{} {
	//		log.Println("Presence is empty, clearing")
	//		return ClearPresence()
	//	}

	log.Printf("Updating presence: %+v", a.presence)
	return client.SetActivity(a.presence)
}
