package worker

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-home.io/x/server/mocks"
	busPlugin "go-home.io/x/server/plugins/bus"
	"go-home.io/x/server/systems/bus"
	"go-home.io/x/server/utils"
)

// Tests worker discovery
func TestNewWorker(t *testing.T) {
	busCalled := false
	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	settings.(mocks.IFakeSettings).AddSBCallback(func(i ...interface{}) {
		assert.Equal(t, settings.NodeID(), i[0].(*bus.DiscoveryMessage).NodeID, "wrong ID")
		busCalled = true
	})
	w, _ := NewWorker(settings)
	go w.Start()

	time.Sleep(1 * time.Second)
	assert.True(t, busCalled, "bus was not called")
}

type dSuite struct {
	suite.Suite

	w *GoHomeWorker
	s *fakeSwitch
}

func (d *dSuite) SetupTest() {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	d.w, _ = NewWorker(settings)
	go d.w.Start()
	d.s = &fakeSwitch{}
	settings.(mocks.IFakeSettings).AddLoader(d.s)
}

// Tests device assignment.
func (d *dSuite) TestAssignment() {
	d.w.workerChan <- busPlugin.RawMessage{Body: []byte(fmt.Sprintf(`
{ 
"mt": "device_assignment",
"d": [
	{"t": "switch", "n": "test", "a": false}
],
"st": %d 
}
`, utils.TimeNow()))}

	time.Sleep(1 * time.Second)
	assert.True(d.T(), d.s.loadCalled, "load")
	assert.False(d.T(), d.s.unloadCalled, "unload")
}

// Tests device command.
func (d *dSuite) TestCommand() {
	d.TestAssignment()
	d.w.workerChan <- busPlugin.RawMessage{Body: []byte(fmt.Sprintf(`
{ 
"mt": "device_command",
"i": "test.switch.fake_switch",
"c": "on",
"st": %d 
}
`, utils.TimeNow()))}

	time.Sleep(1 * time.Second)
	assert.True(d.T(), d.s.onCalled)
}

// Tests devices assignments.
func TestWorker(t *testing.T) {
	suite.Run(t, new(dSuite))
}
