package bloxstraprpc

import (
	"log"

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

func (a *Activity) updateGamePresence() error {
	if a.universeID == "" {
		log.Println("Not in game, clearing presence!")

		return a.client.SetActivity(a.presence)
	}

	tn, err := api.GetGameIcon(a.universeID, "PlaceHolder", "512x512", "Png", false)
	if err != nil {
		return err
	}

	buttons := []drpc.Button{{
		Label: "See game page",
		URL:   "https://www.roblox.com/games/" + a.placeID,
	}}

	joinurl := "roblox://experiences/start?placeId=" + a.placeID + "&gameInstanceId=" + a.jobID
	
	gd, err := api.GetGameDetails(a.universeID)
	if err != nil {
		return err
	}

	status := "by " + gd.Creator.Name

	switch a.server {
	case Public:
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

	log.Printf("Updating Discord presence: %#v", a.presence)
	return a.client.SetActivity(a.presence)
}
