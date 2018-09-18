package bus

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/utils"
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

	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "ping",  "st": %d}`, utils.TimeNow()))})
	time.Sleep(1 * time.Second)
	if !disco || upd {
		t.Error("Discovery failed")
		t.Fail()
	}

	disco = false
	upd = false
	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "device_update",  "st": %d}`, utils.TimeNow()))})
	time.Sleep(1 * time.Second)
	if !upd || disco {
		t.Error("Device update failed")
		t.Fail()
	}

	disco = false
	upd = false
	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "device_command",  "st": %d}`, utils.TimeNow()))})
	time.Sleep(1 * time.Second)
	if upd || disco {
		t.Error("Incorrect message failed")
		t.Fail()
	}

	disco = false
	upd = false
	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "device_update",  "st": %d}`, utils.TimeNow()-(bus.MsgTTLSeconds+1)))})
	time.Sleep(1 * time.Second)
	if upd || disco {
		t.Error("Old message failed")
		t.Fail()
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

	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "device_assignment",  "st": %d}`, utils.TimeNow()))})
	time.Sleep(1 * time.Second)
	if !assign || cmd {
		t.Error("Device assignment failed")
		t.Fail()
	}

	assign = false
	cmd = false
	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "device_command",  "st": %d}`, utils.TimeNow()))})
	time.Sleep(1 * time.Second)
	if !cmd || assign {
		t.Error("Device command failed")
		t.Fail()
	}

	assign = false
	cmd = false
	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "ping",  "st": %d}`, utils.TimeNow()))})
	time.Sleep(1 * time.Second)
	if assign || cmd {
		t.Error("Incorrect message failed")
		t.Fail()
	}

	assign = false
	cmd = false
	p.ProcessIncomingMessage(&bus.RawMessage{Body: []byte(
		fmt.Sprintf(`{"mt": "device_command",  "st": %d}`, utils.TimeNow()-(bus.MsgTTLSeconds+1)))})
	time.Sleep(1 * time.Second)
	if cmd || assign {
		t.Error("Old message failed")
		t.Fail()
	}
}
