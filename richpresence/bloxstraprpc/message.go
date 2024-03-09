package bloxstraprpc

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/altfoxie/drpc"
)

type Timestamp int64

type RichPresenceImage struct {
	AssetID   *int64  `json:"assetId"`
	HoverText *string `json:"hoverText"`
	Clear     bool    `json:"clear"`
	Reset     bool    `json:"reset"`
}

type MessageData struct {
	Details        *string            `json:"details"`
	State          *string            `json:"state"`
	TimestampStart *Timestamp         `json:"timeStart"`
	TimestampEnd   *Timestamp         `json:"timeEnd"`
	SmallImage     *RichPresenceImage `json:"smallImage"`
	LargeImage     *RichPresenceImage `json:"largeImage"`
}

type Message struct {
	Command string      `json:"command"`
	Data    MessageData `json:"data"`
}

// NewMessage constructs a new Message from a BloxstrapRPC message
// log entry from the Roblox client.
func ParseMessage(line string) (*Message, error) {
	var m Message

	msg := line[strings.Index(line, bloxstrapRPCEntry)+len(bloxstrapRPCEntry)+1:]

	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		return nil, err
	}

	if m.Command == "" {
		return nil, errors.New("command is empty")
	}

	// discord RPC implementation requires a limit of 128 characters
	if m.Data.Details != nil && len(*m.Data.Details) > 128 {
		return nil, errors.New("details must be less than 128 characters")
	}

	if m.Data.State != nil && len(*m.Data.State) > 128 {
		return nil, errors.New("state must be less than 128 characters")
	}

	return &m, nil
}

// ApplyRichPresence applies/appends Message's properties to the given
// [drpc.BloxstrapRPC] for use in Discord's Rich Presence.
//
// UpdateGamePresence should be called as some of the properties are specific
// to BloxstrapRPC.
func (m *Message) ApplyRichPresence(p *drpc.Activity) {
	if m.Command != "SetRichPresence" {
		slog.Warn("Game sent invalid BloxstrapRPC command", "command", m.Command)
		return
	}

	if m.Data.Details != nil {
		p.Details = *m.Data.Details
	}

	if m.Data.State != nil {
		p.State = *m.Data.State
	}

	m.Data.TimestampStart.ApplyRichPresence(&p.Timestamps.Start)
	m.Data.TimestampEnd.ApplyRichPresence(&p.Timestamps.End)
	m.Data.SmallImage.ApplyRichPresence(&p.Assets.SmallImage, &p.Assets.SmallText)
	m.Data.LargeImage.ApplyRichPresence(&p.Assets.LargeImage, &p.Assets.LargeText)
}

// ApplyRichPresence applies/appends the Timestamp to the given drpc timestamp.
func (t Timestamp) ApplyRichPresence(drpcTimestamp *time.Time) {
	if drpcTimestamp == nil {
		return
	}

	*drpcTimestamp = time.Time{}
	if t != 0 {
		*drpcTimestamp = time.UnixMilli(int64(t))
	}
}

// ApplyRichPresence applies/appends the Timestamp to the given drpc Asset.
func (i *RichPresenceImage) ApplyRichPresence(drpcImage, drpcText *string) {
	if drpcImage == nil || drpcText == nil {
		return
	}

	if i.Clear {
		*drpcImage = ""
	}

	if i.Reset {
		*drpcImage = reset
		*drpcText = reset
	}

	if i.AssetID != nil {
		*drpcImage = "https://assetdelivery.roblox.com/v1/asset/?id=" +
			strconv.FormatInt(*i.AssetID, 10)
	}

	if i.HoverText != nil {
		*drpcText = *i.HoverText
	}
}
