package bus

import (
	"encoding/json"

	"go-home.io/x/server/plugins/bus"
	"go-home.io/x/server/utils"
)

// Parses raw message and checks whether it should be skipped due to the age.
func parseRawMessage(r *bus.RawMessage) (*MessageWithType, error) {
	var b MessageWithType
	if err := json.Unmarshal(r.Body, &b); err != nil {
		return nil, &ErrCorruptedMessage{}
	}

	if utils.TimeNow()-b.SendTime > bus.MsgTTLSeconds {
		return nil, &ErrOldMessage{}
	}

	return &b, nil
}
