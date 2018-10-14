package bus

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/utils"
	"github.com/stretchr/testify/assert"
)

// Tests master server messages parsing.
func TestMasterServerParser(t *testing.T) {
	p := NewMasterMessageParser(mocks.FakeNewLogger(nil))
	disco := false
	upd := false

	go func() {
		for {
			select {
			case <-p.GetDiscoveryMessageChan():
				disco = true
			case <-p.GetDeviceUpdateMessageChan():
				upd = true
			}
		}
	}()

	data := []struct {
		msg   string
		disco bool
		upd   bool
		err   string
	}{
		{
			msg:   fmt.Sprintf(`{"mt": "ping",  "st": %d}`, utils.TimeNow()),
			disco: true,
			upd:   false,
			err:   "discovery",
		},
		{
			msg:   fmt.Sprintf(`{"mt": "device_update",  "st": %d}`, utils.TimeNow()),
			disco: false,
			upd:   true,
			err:   "device update",
		},
		{
			msg:   fmt.Sprintf(`{"mt": "device_command",  "st": %d}`, utils.TimeNow()),
			disco: false,
			upd:   false,
			err:   "wrong message",
		},
		{
			msg:   fmt.Sprintf(`{"mt": "device_update",  "st": %d}`, utils.TimeNow()-(bus.MsgTTLSeconds+1)),
			disco: false,
			upd:   false,
			err:   "old message",
		},
	}

	for _, v := range data {
		disco = false
		upd = false
		p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(v.msg)})
		time.Sleep(1 * time.Second)
		assert.Equal(t, v.upd, upd, "update %s", v.err)
		assert.Equal(t, v.disco, disco, "discovery %s", v.err)
	}
}

// Tests worker server messages parser.
func TestWorkerServerParser(t *testing.T) {
	p := NewWorkerMessageParser(mocks.FakeNewLogger(nil))
	assign := false
	cmd := false

	go func() {
		for {
			select {
			case <-p.GetDeviceAssignmentMessageChan():
				assign = true
			case <-p.GetDeviceCommandMessageChan():
				cmd = true
			}
		}
	}()

	data := []struct {
		msg    string
		assign bool
		cmd    bool
		err    string
	}{
		{
			msg:    fmt.Sprintf(`{"mt": "device_assignment",  "st": %d}`, utils.TimeNow()),
			assign: true,
			cmd:    false,
			err:    "device assignment",
		},
		{
			msg:    fmt.Sprintf(`{"mt": "device_command",  "st": %d}`, utils.TimeNow()),
			assign: false,
			cmd:    true,
			err:    "device command",
		},
		{
			msg:    fmt.Sprintf(`{"mt": "ping",  "st": %d}`, utils.TimeNow()),
			assign: false,
			cmd:    false,
			err:    "incorrect message",
		},
		{
			msg:    fmt.Sprintf(`{"mt": "device_command",  "st": %d}`, utils.TimeNow()-(bus.MsgTTLSeconds+1)),
			assign: false,
			cmd:    false,
			err:    "old message",
		},
	}

	for _, v := range data {
		assign = false
		cmd = false
		p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(v.msg)})
		time.Sleep(1 * time.Second)
		assert.Equal(t, v.cmd, cmd, "command %s", v.err)
		assert.Equal(t, v.assign, assign, "assignment %s", v.err)
	}
}
