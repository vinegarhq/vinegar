package bloxstraprpc

import (
	"encoding/json"
	"errors"
	"log"
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

	if !strings.Contains(line, GameMessageEntry) {
		return m, nil
	}

	msg := line[strings.Index(line, GameMessageEntry)+len(GameMessageEntry)+1:]

	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		return m, err
	}

	if m.Command == "" {
		return m, errors.New("command is empty")
	}

	if len(m.Data.Details) > 128 {
		return m, errors.New("details cannot be longer than 128 characters")
	}

	if len(m.Data.State) > 128 {
		return m, errors.New("details cannot be longer than 128 characters")
	}

	log.Printf("Received message: %+v", m)

	return m, nil
}
