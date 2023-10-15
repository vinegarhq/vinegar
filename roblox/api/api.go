package api

import (
	"encoding/json"
	"log"

	"github.com/vinegarhq/vinegar/util"
)

func UnmarshalBody(url string, v any) error {
	log.Printf("Sending API Request for %s", url)
	body, err := util.Body(url)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(body), &v)
}
