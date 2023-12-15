package bloxstraprpc

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
)

type RichPresenceImage struct {
	AssetID   int64 `json:"assetId"`
	HoverText int64 `json:"hoverText"`
	Clear     bool  `json:"clear"`
	Reset     bool  `json:"reset"`
}

type Data struct {
	Details        string            `json:"details"`
	State          string            `json:"state"`
	TimestampStart int64             `json:"timeStart"`
	TimestampEnd   int64             `json:"timeEnd"`
	SmallImage     RichPresenceImage `json:"smallImage"`
	LargeImage     RichPresenceImage `json:"largeImage"`
}

type Message struct {
	Command string `json:"command"`
	Data    `json:"data"`
}

func ParseMessage(line string) (Message, error) {
	var m Message

	msg, err := strconv.Unquote(`"` + line[strings.Index(line, GameMessageEntry)+len(GameMessageEntry)+1:] + `"`)
	if err != nil {
		return Message{}, err
	}

	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		return Message{}, err
	}

	if m.Command == "" {
		return Message{}, errors.New("command is empty")
	}

	// discord RPC implementation requires a limit of 128 characters
	if len(m.Data.Details) > 128 || len(m.Data.State) > 128 {
		return Message{}, errors.New("details or state cannot be longer than 128 characters")
	}

	log.Printf("Received message: %+v", m)

	return m, nil
}
