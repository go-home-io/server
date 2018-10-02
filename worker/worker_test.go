package worker

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-home-io/server/utils"

	"github.com/go-home-io/server/mocks"
	busPlugin "github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/systems/bus"
)

// Tests worker discovery
func TestNewWorker(t *testing.T) {
	busCalled := false
	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	settings.(mocks.IFakeSettings).AddSBCallback(func(i ...interface{}) {
		if i[0].(*bus.DiscoveryMessage).NodeID != settings.NodeID() {
			t.Error("Node ID failed")
			t.Fail()
		}
		busCalled = true
	})
	w, _ := NewWorker(settings)
	go w.Start()

	time.Sleep(1 * time.Second)
	if !busCalled {
		t.Error("Bus was not called")
		t.Fail()
	}
}

// Tests devices assignment.
func TestDeviceAssignmentAndCommand(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	w, _ := NewWorker(settings)
	go w.Start()

	s := &fakeSwitch{}
	settings.(mocks.IFakeSettings).AddLoader(s)

	time.Sleep(1 * time.Second)
	w.workerChan <- busPlugin.RawMessage{Body: []byte(fmt.Sprintf(`
{ 
"mt": "device_assignment",
"d": [
	{"t": "switch", "n": "test", "a": false}
],
"st": %d 
}
`, utils.TimeNow()))}

	time.Sleep(1 * time.Second)
	if !s.loadCalled || s.unloadCalled {
		t.Error("Load failed")
		t.FailNow()
	}

	w.workerChan <- busPlugin.RawMessage{Body: []byte(fmt.Sprintf(`
{ 
"mt": "device_command",
"i": "test.switch.fake_switch",
"c": "on",
"st": %d 
}
`, utils.TimeNow()))}

	time.Sleep(1 * time.Second)
	if !s.onCalled {
		t.Error("Command failed")
		t.Fail()
	}
}
