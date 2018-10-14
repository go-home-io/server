package bus

import (
	"encoding/json"

	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/utils"
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
