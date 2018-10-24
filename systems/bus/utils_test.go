package bus

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/plugins/bus"
	"go-home.io/x/server/utils"
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
