package bus

import (
	"encoding/json"
	"errors"

	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/utils"
)

// Parses raw message and checks whether it should be skipped due to the age.
func parseRawMessage(r *bus.RawMessage) (*MessageWithType, error) {
	var b MessageWithType
	if err := json.Unmarshal(r.Body, &b); err != nil {
		return nil, errors.New("failed to unmarshal bus message")
	}

	if utils.TimeNow()-b.SendTime > bus.MsgTTLSeconds {
		return nil, errors.New("message is too old")
	}

	return &b, nil
}
