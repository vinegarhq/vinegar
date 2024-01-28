package bloxstraprpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/altfoxie/drpc"
)

// RichPresenceImage holds game image information sent
// from a BloxstrapRPC message
type RichPresenceImage struct {
	AssetID   int64  `json:"assetId"`
	HoverText string `json:"hoverText"`
	Clear     bool   `json:"clear"`
	Reset     bool   `json:"reset"`
}

// Data holds game information sent from a BloxstrapRPC message
type Data struct {
	Details        string            `json:"details"`
	State          string            `json:"state"`
	TimestampStart int64             `json:"timeStart"`
	TimestampEnd   int64             `json:"timeEnd"`
	SmallImage     RichPresenceImage `json:"smallImage"`
	LargeImage     RichPresenceImage `json:"largeImage"`
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
		return Message{}, fmt.Errorf("bloxstraprpc message: %w", err)
	}

	if m.Command == "" {
		return Message{}, errors.New("bloxstraprpc message: command is empty")
	}

	// discord RPC implementation requires a limit of 128 characters
	if len(m.Data.Details) > 128 || len(m.Data.State) > 128 {
		return Message{}, errors.New("bloxstraprpc message: details or state cannot be longer than 128 characters")
	}

	return m, nil
}

func (m Message) ApplyPresence(p *drpc.Activity) {
	if m.Command != "SetRichPresence" {
		log.Printf("WARNING: Game sent invalid BloxstrapRPC command: %s", m.Command)
		return
	}

	if m.Data.Details != "" {
		p.Details = m.Data.Details
	}

	if m.Data.State != "" {
		p.State = m.Data.State
	}

	if m.TimestampStart == 0 && m.TimestampEnd == 0 {
		p.Timestamps = nil
	}

	// Upstream RPC lib checks with time.IsZero
	if p.Timestamps != nil {
		if m.TimestampStart == 0 {
			p.Timestamps.Start = time.Time{}
		} else {
			p.Timestamps.Start = time.UnixMilli(m.TimestampStart)
		}
		if m.TimestampEnd == 0 {
			p.Timestamps.End = time.Time{}
		} else {
			p.Timestamps.End = time.UnixMilli(m.TimestampEnd)
		}
	}

	if m.SmallImage.Clear {
		p.Assets.SmallImage = ""
	}

	if m.SmallImage.AssetID != 0 {
		p.Assets.SmallImage = "https://assetdelivery.roblox.com/v1/asset/?id=" +
			strconv.FormatInt(m.SmallImage.AssetID, 10)
	}

	if m.LargeImage.Clear {
		p.Assets.LargeImage = ""
	}

	if m.LargeImage.AssetID != 0 {
		p.Assets.LargeImage = "https://assetdelivery.roblox.com/v1/asset/?id=" +
			strconv.FormatInt(m.LargeImage.AssetID, 10)
	}
}
