package bloxstraprpc

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/altfoxie/drpc"
)

// RichPresenceImage holds game image information sent
// from a BloxstrapRPC message
type RichPresenceImage struct {
	AssetID   *int64  `json:"assetId"`
	HoverText *string `json:"hoverText"`
	Clear     bool    `json:"clear"`
	Reset     bool    `json:"reset"`
}

// Data holds game information sent from a BloxstrapRPC message
type Data struct {
	Details        *string            `json:"details"`
	State          *string            `json:"state"`
	TimestampStart *int64             `json:"timeStart"`
	TimestampEnd   *int64             `json:"timeEnd"`
	SmallImage     *RichPresenceImage `json:"smallImage"`
	LargeImage     *RichPresenceImage `json:"largeImage"`
}

// Message is a representation of a BloxstrapRPC message sent
// from a Roblox game using the BloxstrapRPC SDK.
type Message struct {
	Command string `json:"command"`
	Data    `json:"data"`
}

// NewMessage constructs a new Message from a BloxstrapRPC message
// log entry from the Roblox client.
func NewMessage(line string) (Message, error) {
	var m Message

	msg := line[strings.Index(line, BloxstrapRPCEntry)+len(BloxstrapRPCEntry)+1:]

	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		return Message{}, err
	}

	if m.Command == "" {
		return Message{}, errors.New("command is empty")
	}

	// discord RPC implementation requires a limit of 128 characters
	if m.Data.Details != nil && len(*m.Data.Details) > 128 {
		return Message{}, errors.New("details must be less than 128 characters")
	}

	if m.Data.State != nil && len(*m.Data.State) > 128 {
		return Message{}, errors.New("state must be less than 128 characters")
	}

	return m, nil
}

// ApplyRichPresence applies/appends Message's properties to the given
// [drpc.Activity] for use in Discord's Rich Presence.
func (m Message) ApplyRichPresence(p *drpc.Activity) {
	if m.Command != "SetRichPresence" {
		log.Printf("WARNING: Game sent invalid BloxstrapRPC command: %s", m.Command)
		return
	}

	if m.Data.Details != nil {
		p.Details = *m.Data.Details
	}

	if m.Data.State != nil {
		p.State = *m.Data.State
	}

	if m.TimestampStart != nil {
		if *m.TimestampStart == 0 {
			p.Timestamps.Start = time.Time{}
		} else {
			p.Timestamps.Start = time.UnixMilli(*m.TimestampStart)
		}
	}
	if m.TimestampEnd != nil {
		if *m.TimestampEnd == 0 {
			p.Timestamps.End = time.Time{}
		} else {
			p.Timestamps.End = time.UnixMilli(*m.TimestampEnd)
		}
	}

	if m.SmallImage != nil {
		if m.SmallImage.Clear {
			p.Assets.SmallImage = ""
		}

		if m.SmallImage.Reset {
			p.Assets.SmallImage = Reset
			*m.SmallImage.HoverText = Reset
		}

		if m.SmallImage.AssetID != nil {
			p.Assets.SmallImage = "https://assetdelivery.roblox.com/v1/asset/?id=" +
				strconv.FormatInt(*m.SmallImage.AssetID, 10)
		}

		if m.SmallImage.HoverText != nil {
			p.Assets.SmallText = *m.SmallImage.HoverText
		}
	}

	if m.LargeImage != nil {
		if m.LargeImage.Clear {
			p.Assets.LargeImage = ""
		}

		if m.LargeImage.Reset {
			p.Assets.LargeImage = Reset
			*m.LargeImage.HoverText = Reset
		}

		if m.LargeImage.AssetID != nil {
			p.Assets.LargeImage = "https://assetdelivery.roblox.com/v1/asset/?id=" +
				strconv.FormatInt(*m.LargeImage.AssetID, 10)
		}

		if m.LargeImage.HoverText != nil {
			p.Assets.SmallText = *m.LargeImage.HoverText
		}
	}
}
