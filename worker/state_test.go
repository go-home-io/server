package worker

import (
	"errors"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/api"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/systems/bus"
	"go-home.io/x/server/utils"
)

// Fake sensor plugin.
type fakeDevicePlugin struct {
	loadError   error
	updateError error
	loadTimeout time.Duration

	unloadCalled  bool
	getSpecCalled bool
	loadCalled    bool
	updateCalled  bool
}

func (f *fakeDevicePlugin) Init(*device.InitDataDevice) error {
	return nil
}

func (f *fakeDevicePlugin) Unload() {
	f.unloadCalled = true
}

func (f *fakeDevicePlugin) GetName() string {
	return "fake device"
}

func (f *fakeDevicePlugin) GetSpec() *device.Spec {
	f.getSpecCalled = true
	return &device.Spec{
		UpdatePeriod:        1 * time.Second,
		SupportedCommands:   make([]enums.Command, 0),
		SupportedProperties: []enums.Property{enums.PropOn},
	}
}

func (f *fakeDevicePlugin) Load() (*device.SensorState, error) {
	f.loadCalled = true
	time.Sleep(f.loadTimeout)
	return &device.SensorState{
		On: true,
	}, f.loadError
}

func (f *fakeDevicePlugin) Update() (*device.SensorState, error) {
	f.updateCalled = true
	return &device.SensorState{
		On: true,
	}, f.updateError
}

// Fake hub plugin.
type fakeHub struct {
	unloadCalled bool
	disc         chan *device.DiscoveredDevices
}

func (f *fakeHub) FakeInit(data interface{}) {
	f.disc = data.(*device.InitDataDevice).DeviceDiscoveredChan
}

func (f *fakeHub) Init(*device.InitDataDevice) error {
	return nil
}

func (f *fakeHub) Unload() {
	f.unloadCalled = true
}

func (*fakeHub) GetName() string {
	return "hub"
}

func (*fakeHub) GetSpec() *device.Spec {
	return &device.Spec{}
}

func (*fakeHub) Load() (*device.HubLoadResult, error) {
	return &device.HubLoadResult{
		Devices: []*device.DiscoveredDevices{},
		State: &device.HubState{
			NumDevices: 0,
		},
	}, nil
}

func (*fakeHub) Update() (*device.HubLoadResult, error) {
	return &device.HubLoadResult{
		Devices: []*device.DiscoveredDevices{},
		State: &device.HubState{
			NumDevices: 0,
		},
	}, nil
}

func (f *fakeHub) Disco(s *fakeSwitch) {
	f.disc <- &device.DiscoveredDevices{
		Type: enums.DevSwitch,
		State: &device.SwitchState{
			On: true,
		},
		Interface: s,
	}
}

// Fake switch plugin.
type fakeSwitch struct {
	update       chan *device.StateUpdateData
	loadCalled   bool
	unloadCalled bool
	onCalled     bool
}

func (f *fakeSwitch) Init(data *device.InitDataDevice) error {
	f.update = data.DeviceStateUpdateChan
	return nil
}

func (f *fakeSwitch) Unload() {
	f.unloadCalled = true
}

func (*fakeSwitch) GetName() string {
	return "fake switch"
}

func (*fakeSwitch) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedCommands:   []enums.Command{enums.CmdOn},
		SupportedProperties: []enums.Property{enums.PropOn},
	}
}

func (f *fakeSwitch) Load() (*device.SwitchState, error) {
	f.loadCalled = true
	return &device.SwitchState{On: false}, nil
}

func (f *fakeSwitch) On() error {
	f.onCalled = true
	return nil
}

func (*fakeSwitch) Off() error {
	return nil
}

func (*fakeSwitch) Toggle() error {
	return nil
}

func (*fakeSwitch) Update() (*device.SwitchState, error) {
	return &device.SwitchState{On: false}, nil
}

func (f *fakeSwitch) Push() {
	f.update <- &device.StateUpdateData{
		State: &device.SwitchState{On: true},
	}
}

// Fake API plugin.
type fakeAPI struct {
	unloadCalled bool
}

func (f *fakeAPI) Init(*api.InitDataAPI) error {
	return nil
}

func (f *fakeAPI) Routes() []string {
	return []string{}
}

func (f *fakeAPI) Unload() {
	f.unloadCalled = true
}

// Tests proper devices unloading.
func TestStaleMaster(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	state := newWorkerState(settings)
	d := &fakeDevicePlugin{}
	settings.(mocks.IFakeSettings).AddLoader(d)

	utils.LongTimeNoSee = 1
	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.True(t, d.loadCalled, "load failed")
	assert.Equal(t, 1, len(state.devices), "load failed")

	time.Sleep(1 * time.Second)
	d.loadCalled = false
	state.retryLoad()
	time.Sleep(1 * time.Second)
	assert.False(t, d.loadCalled, "retry called")

	state.checkStaleMaster()
	time.Sleep(1 * time.Second)
	assert.True(t, d.unloadCalled, "unload was not called")
	assert.Equal(t, 0, len(state.devices), "unload was not called")
}

// Tests that if plugin failed to load, it will be correctly reloaded next time.
func TestFailedReload(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	state := newWorkerState(settings)
	d := &fakeDevicePlugin{
		loadError: errors.New("load"),
	}
	settings.(mocks.IFakeSettings).AddLoader(d)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.True(t, d.loadCalled, "load not called")
	assert.Equal(t, 0, len(state.devices), "loaded wrong device")
	assert.NotNil(t, state.failedDevices, "no failed device")
	assert.Equal(t, 1, len(state.failedDevices.Devices), "loaded wrong device")

	d.loadCalled = false
	d.loadError = nil
	state.retryLoad()
	time.Sleep(1 * time.Second)
	assert.True(t, d.loadCalled, "load not called")
	assert.Equal(t, 1, len(state.devices), "didn't load device")
	assert.Nil(t, state.failedDevices, "failed not nil")
}

// Tests reload.
func TestSelfReload(t *testing.T) {
	monkey.Patch(getNextRetryTime, func(_ int) time.Time { return time.Now().Add(500 * time.Millisecond) })
	defer monkey.UnpatchAll()

	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	state := newWorkerState(settings)
	d := &fakeDevicePlugin{
		loadError: errors.New("load"),
	}
	settings.(mocks.IFakeSettings).AddLoader(d)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.True(t, d.loadCalled, "load not called")
	assert.Equal(t, 0, len(state.devices), "loaded wrong device")
	assert.NotNil(t, state.failedDevices, "no failed device")
	assert.Equal(t, 1, len(state.failedDevices.Devices), "loaded wrong device")

	d.loadCalled = false
	d.loadError = nil

	time.Sleep(1200 * time.Millisecond)

	assert.True(t, d.loadCalled, "load not called")
	assert.Equal(t, 1, len(state.devices), "didn't load device")
	assert.Nil(t, state.failedDevices, "failed not nil")
}

// Tests proper unload
func TestProperUnloadAfterTimeout(t *testing.T) {
	deviceLoadTimeout = 1 * time.Second

	settings := mocks.FakeNewSettings(nil, true, nil, nil)
	state := newWorkerState(settings)
	d1 := &fakeDevicePlugin{
		loadTimeout: 3 * time.Second,
	}
	settings.(mocks.IFakeSettings).AddLoader(d1)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "fake device",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.True(t, d1.loadCalled, "load not called")
	assert.False(t, d1.unloadCalled, "unload called")
	assert.Equal(t, 0, len(state.devices), "loaded wrong device")
	assert.NotNil(t, state.failedDevices, "no failed device")
	assert.Equal(t, 1, len(state.failedDevices.Devices), "loaded wrong device")

	d2 := &fakeDevicePlugin{}

	settings.(mocks.IFakeSettings).AddLoader(d2)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device2",
				IsAPI:  false,
				Plugin: "fake device2",
				Config: "fake device2",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.True(t, d2.loadCalled, "load not called")
	assert.Equal(t, 1, len(state.devices), "loaded wrong device")
	assert.Nil(t, state.failedDevices, "no failed device")

	time.Sleep(2200 * time.Millisecond)
	assert.True(t, d1.unloadCalled, "unload not called")
}

// Tests timeout during device load.
func TestLoadTimeout(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)

	deviceLoadTimeout = 1 * time.Second

	state := newWorkerState(settings)
	d := &fakeDevicePlugin{
		loadTimeout: 2 * time.Second,
	}
	settings.(mocks.IFakeSettings).AddLoader(d)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
			},
		},
	})

	time.Sleep(3 * time.Second)
	assert.True(t, d.loadCalled, "load not called")
	assert.Equal(t, 0, len(state.devices), "device num")
	assert.True(t, d.unloadCalled, "unload not called")
	assert.NotNil(t, state.failedDevices, "has failed")
	assert.Equal(t, 1, len(state.failedDevices.Devices), "failed num")

	d.loadTimeout = 0
	d.loadCalled = false
	d.unloadCalled = false
	state.retryLoad()
	time.Sleep(1 * time.Second)
	assert.True(t, d.loadCalled, "load not called")
	assert.False(t, d.unloadCalled, "unload called")
	assert.Nil(t, state.failedDevices, "has failed")
	assert.Equal(t, 1, len(state.devices), "device num")
}

// Tests long load with a new assignment.
func TestLoadTimeoutWithNewMessage(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)

	deviceLoadTimeout = 1 * time.Second

	state := newWorkerState(settings)
	d := &fakeDevicePlugin{
		loadTimeout: 2 * time.Second,
	}

	d2 := &fakeDevicePlugin{}

	settings.(mocks.IFakeSettings).AddLoader(d)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 1",
			},
		},
	})

	time.Sleep(3 * time.Second)
	assert.True(t, d.loadCalled, "first load")
	assert.True(t, d.unloadCalled, "first unload load")
	assert.Equal(t, 0, len(state.devices), "fist wrong num")
	assert.NotNil(t, state.failedDevices, "first failed")
	assert.Equal(t, 1, len(state.failedDevices.Devices), "first failed num")

	d.loadTimeout = 0
	d.loadCalled = false
	d.unloadCalled = false

	settings.(mocks.IFakeSettings).AddLoader(d2)
	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device2",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 2",
			},
		},
	})

	state.retryLoad()
	time.Sleep(1 * time.Second)

	assert.False(t, d.loadCalled, "first load")
	assert.True(t, d2.loadCalled, "second load")
	assert.False(t, d2.unloadCalled, "second unload")
	assert.Equal(t, 1, len(state.devices), "wrong num")
	assert.NotNil(t, state.devices["fake_device2.sensor.fake_device"], "wrong state")
}

// Tests that device will be unloaded when a new assignment is received.
func TestProperUnloadOnANewAssignment(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)

	state := newWorkerState(settings)
	d := &fakeDevicePlugin{}
	d2 := &fakeDevicePlugin{}

	settings.(mocks.IFakeSettings).AddLoader(d)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 1",
			},
		},
	})

	time.Sleep(1 * time.Second)
	require.True(t, d.loadCalled, "plugin was not loaded")
	require.Equal(t, 1, len(state.devices), "incorrect devices number")
	require.False(t, d.unloadCalled, "unload called")
	require.Nil(t, state.failedDevices, "found failed devices")

	d.loadCalled = false
	d.unloadCalled = false

	settings.(mocks.IFakeSettings).AddLoader(d2)
	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device2",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 2",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.False(t, d.loadCalled, "first plugin called")
	assert.True(t, d.unloadCalled, "unload was not called")
	assert.Equal(t, 1, len(state.devices), "device num is incorrect")

	assert.True(t, d2.loadCalled, "second plugin was not called")
	assert.False(t, d2.unloadCalled, "second unload was called")
}

// Tests that same assignment won't trigger reload.
func TestSameAssignment(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)

	state := newWorkerState(settings)
	d := &fakeDevicePlugin{}
	settings.(mocks.IFakeSettings).AddLoader(d)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 1",
			},
		},
	})

	time.Sleep(1 * time.Second)
	require.True(t, d.loadCalled, "plugin was not loaded")
	require.Equal(t, 1, len(state.devices), "incorrect devices number")
	require.False(t, d.unloadCalled, "unload called")
	require.Nil(t, state.failedDevices, "found failed devices")

	d.loadCalled = false
	d.unloadCalled = false

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 1",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.False(t, d.unloadCalled, "unload called")
	assert.False(t, d.loadCalled, "load called")
	assert.Equal(t, 1, len(state.devices), "device num is incorrect")
}

// Tests device discovery and commands.
func TestDeviceDiscoveryAndCommands(t *testing.T) {
	busCalled := false
	settings := mocks.FakeNewSettings(nil, true, nil, nil)

	settings.(mocks.IFakeSettings).AddSBCallback(func(_ ...interface{}) {
		busCalled = true
	})

	state := newWorkerState(settings)
	h := &fakeHub{}
	settings.(mocks.IFakeSettings).AddLoader(h)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevHub,
				Name:   "fake hub",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "hub",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, len(state.devices), "hub was not loaded")
	assert.Nil(t, state.failedDevices, "hub failed to load")

	s := &fakeSwitch{}
	settings.(mocks.IFakeSettings).AddLoader(s)
	h.Disco(s)
	time.Sleep(1 * time.Second)
	assert.True(t, busCalled, "bus was not called during the discovery")
	assert.Equal(t, 2, len(state.devices), "discovery didn't work")

	busCalled = false
	s.Push()
	time.Sleep(1 * time.Second)
	assert.True(t, busCalled, "bus was not called during the update")

	state.DevicesCommandMessage(&bus.DeviceCommandMessage{
		Command:  enums.CmdOn,
		DeviceID: "fake_hub.switch.fake_switch",
	})

	time.Sleep(1 * time.Second)
	assert.True(t, s.onCalled, "command was not invoked")

	s.onCalled = false
	s.unloadCalled = false
	h.unloadCalled = false

	d := &fakeDevicePlugin{}
	settings.(mocks.IFakeSettings).AddLoader(d)
	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 1",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.True(t, h.unloadCalled, "unload was not called")
	assert.True(t, s.unloadCalled, "unload was not called on a switch")

	require.True(t, d.loadCalled, "plugin was not loaded")
	require.Equal(t, 1, len(state.devices), "incorrect devices number")
	require.False(t, d.unloadCalled, "unload called")
	require.Nil(t, state.failedDevices, "found failed devices")

	time.Sleep(1 * time.Second)
	busCalled = false
	assert.True(t, wait(&busCalled), "updates were not called")

	busCalled = false
	state.DevicesCommandMessage(&bus.DeviceCommandMessage{
		Command:  enums.CmdOn,
		DeviceID: "fake_hub.switch.fake_switch",
	})
	time.Sleep(100 * time.Millisecond)
	assert.False(t, busCalled, "unknown device bus called")
}

// Waiting for a callback.
func wait(called *bool) bool {
	wait := time.NewTicker(10 * time.Millisecond)
	defer wait.Stop()
	for {
		select {
		case <-wait.C:
			if !*called {
				continue
			}

			return true
		case <-time.After(3 * time.Second):
			return false
		}
	}
}

// Tests API processing.
func TestAPIUnload(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)

	state := newWorkerState(settings)
	a := &fakeAPI{}
	settings.(mocks.IFakeSettings).AddLoader(a)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Name:   "fake api",
				IsAPI:  true,
				Plugin: "fake api",
				Config: "api",
			},
		},
	})

	time.Sleep(1 * time.Second)
	require.Equal(t, 1, len(state.extendedAPIs), "incorrect api number")
	require.Equal(t, 0, len(state.devices), "incorrect devices number")
	require.False(t, a.unloadCalled, "unload called")
	require.Nil(t, state.failedDevices, "found failed api")

	a.unloadCalled = false

	d := &fakeDevicePlugin{}
	settings.(mocks.IFakeSettings).AddLoader(d)
	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Type:   enums.DevSensor,
				Name:   "fake device",
				IsAPI:  false,
				Plugin: "fake device",
				Config: "device 1",
			},
		},
	})

	time.Sleep(1 * time.Second)
	assert.True(t, a.unloadCalled, "unload was not called")
	assert.Equal(t, 0, len(state.extendedAPIs), "incorrect api number")
	assert.Equal(t, 1, len(state.devices), "incorrect devices number")
	assert.Nil(t, state.failedDevices, "found failed devices")
}

// Test API load error.
func TestAPILoadError(t *testing.T) {
	settings := mocks.FakeNewSettings(nil, true, nil, nil)

	state := newWorkerState(settings)
	settings.(mocks.IFakeSettings).AddLoader(nil)
	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Name:   "fake api",
				IsAPI:  true,
				Plugin: "fake api",
				Config: "api",
			},
		},
	})

	time.Sleep(1 * time.Second)
	require.NotNil(t, state.failedDevices, "first was loaded")
	assert.Equal(t, 1, len(state.failedDevices.Devices), "first was loaded")

	a := &fakeAPI{}
	settings.(mocks.IFakeSettings).AddLoader(a)

	state.DevicesAssignmentMessage(&bus.DeviceAssignmentMessage{
		Devices: []*bus.DeviceAssignment{
			{
				Name:   "fake api",
				IsAPI:  true,
				Plugin: "fake api",
				Config: "api",
			},
		},
	})

	time.Sleep(1 * time.Second)
	require.NotNil(t, state.failedDevices, "second was loaded")
	assert.Equal(t, 1, len(state.failedDevices.Devices), "second was loaded")

	state.retryLoad()
	time.Sleep(1 * time.Second)
	assert.False(t, a.unloadCalled, "third unload called")
	assert.Equal(t, 1, len(state.extendedAPIs), "trird api num")
	assert.Equal(t, 0, len(state.devices), "third device num")
	assert.Nil(t, state.failedDevices, "third failed num")
}
