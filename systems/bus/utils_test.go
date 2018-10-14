package bus

import (
	"fmt"
	"testing"

	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/utils"
	"github.com/stretchr/testify/assert"
)

// Tests incorrect JSON.
func TestWrongJson(t *testing.T) {
	msg := &bus.RawMessage{
		Body: []byte("wrong json"),
	}

	m, err := parseRawMessage(msg)
	assert.Error(t, err)
	assert.Nil(t, m)
}

// Tests an old message.
func TestOldMessage(t *testing.T) {
	msg := &bus.RawMessage{
		Body: []byte(fmt.Sprintf(`{ "mt": "ping", "st":  %d }`, utils.TimeNow()-(bus.MsgTTLSeconds+1))),
	}

	m, err := parseRawMessage(msg)
	assert.Error(t, err)
	assert.Nil(t, m)
}

// Tests correct message.
func TestCorrectMessage(t *testing.T) {
	msg := &bus.RawMessage{
		Body: []byte(fmt.Sprintf(`{ "mt": "ping", "st":  %d }`, utils.TimeNow()-(bus.MsgTTLSeconds-1))),
	}

	m, err := parseRawMessage(msg)
	assert.NoError(t, err)
	assert.NotNil(t, m)
}
