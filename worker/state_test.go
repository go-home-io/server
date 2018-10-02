package worker

import (
	"errors"
	"testing"
	"time"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/api"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/utils"
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
	if !d.loadCalled || 1 != len(state.devices) {
		t.Error("Plugin not loaded")
		t.Fail()
	}

	time.Sleep(1 * time.Second)
	d.loadCalled = false
	state.retryLoad()
	time.Sleep(1 * time.Second)
	if d.loadCalled {
		t.Error("Retry called")
		t.Fail()
	}
	state.checkStaleMaster()
	time.Sleep(1 * time.Second)
	if !d.unloadCalled || len(state.devices) != 0 {
		t.Error("Unload not called")
		t.Fail()
	}
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
	if !d.loadCalled || 0 != len(state.devices) || nil == state.failedDevices || 1 != len(state.failedDevices.Devices) {
		t.Error("Plugin was loaded")
		t.Fail()
	}

	d.loadCalled = false
	d.loadError = nil
	state.retryLoad()
	time.Sleep(1 * time.Second)
	if !d.loadCalled || 1 != len(state.devices) || nil != state.failedDevices {
		t.Error("Plugin was not loaded")
		t.Fail()
	}
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
	if !d.loadCalled || 0 != len(state.devices) ||
		!d.unloadCalled || nil == state.failedDevices || 1 != len(state.failedDevices.Devices) {
		t.Error("Plugin was loaded")
		t.Fail()
	}

	d.loadTimeout = 0
	d.loadCalled = false
	d.unloadCalled = false

	state.retryLoad()
	time.Sleep(1 * time.Second)
	if !d.loadCalled || d.unloadCalled || 1 != len(state.devices) || nil != state.failedDevices {
		t.Error("Plugin was not loaded")
		t.Fail()
	}
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

	if !d.loadCalled || 0 != len(state.devices) ||
		!d.unloadCalled || nil == state.failedDevices || 1 != len(state.failedDevices.Devices) {
		t.Error("Plugin was loaded")
		t.Fail()
	}

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
	if d.loadCalled || 1 != len(state.devices) || nil == state.devices["fake_device2.sensor.fake_device"] {
		t.Error("First plugin was loaded")
		t.Fail()
	}

	if !d2.loadCalled || d2.unloadCalled {
		t.Error("Second plugin was not loaded")
		t.Fail()
	}
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
	if !d.loadCalled || 1 != len(state.devices) || d.unloadCalled || nil != state.failedDevices {
		t.Error("Plugin was not loaded")
		t.Fail()
	}

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
	if d.loadCalled || !d.unloadCalled || 1 != len(state.devices) ||
		nil == state.devices["fake_device2.sensor.fake_device"] {
		t.Error("First plugin was loaded")
		t.Fail()
	}

	if !d2.loadCalled || d2.unloadCalled {
		t.Error("Second plugin was not loaded")
		t.Fail()
	}
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
	if !d.loadCalled || 1 != len(state.devices) || d.unloadCalled || nil != state.failedDevices {
		t.Error("Plugin was not loaded")
		t.Fail()
	}

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
	if d.unloadCalled || d.loadCalled || 1 != len(state.devices) {
		t.Error("Device was unloaded")
		t.Fail()
	}
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
	if 1 != len(state.devices) || nil != state.failedDevices {
		t.Error("Hub was not loaded")
		t.Fail()
	}

	s := &fakeSwitch{}
	settings.(mocks.IFakeSettings).AddLoader(s)
	h.Disco(s)
	time.Sleep(1 * time.Second)
	if !busCalled {
		t.Error("Bus was not called during discovery")
		t.Fail()
	}

	if 2 != len(state.devices) {
		t.Error("Discovery didn't work")
		t.Fail()
	}

	busCalled = false
	s.Push()
	time.Sleep(1 * time.Second)
	if !busCalled {
		t.Error("Bus update was not called during update")
		t.Fail()
	}

	state.DevicesCommandMessage(&bus.DeviceCommandMessage{
		Command:  enums.CmdOn,
		DeviceID: "fake_hub.switch.fake_switch",
	})

	time.Sleep(1 * time.Second)

	if !s.onCalled {
		t.Error("Command was not called")
		t.Fail()
	}

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
	if !h.unloadCalled || !s.unloadCalled {
		t.Error("Unload was not called")
		t.Fail()
	}

	if !d.loadCalled || 1 != len(state.devices) || d.unloadCalled || nil != state.failedDevices {
		t.Error("Plugin was not loaded")
		t.Fail()
	}

	busCalled = false
	state.DevicesCommandMessage(&bus.DeviceCommandMessage{
		Command:  enums.CmdOn,
		DeviceID: "fake_hub.switch.fake_switch",
	})

	if busCalled {
		t.Error("Unknown device bus called")
		t.Fail()
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
	if a.unloadCalled || 1 != len(state.extendedAPIs) || 0 != len(state.devices) || nil != state.failedDevices {
		t.Error("API was not loaded")
		t.Fail()
	}

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
	if !a.unloadCalled || 0 != len(state.extendedAPIs) || 1 != len(state.devices) || nil != state.failedDevices {
		t.Error("API was not unloaded")
		t.Fail()
	}
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
	if nil == state.failedDevices || 1 != len(state.failedDevices.Devices) {
		t.Error("Bad API error")
		t.Fail()
	}

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
	if nil == state.failedDevices || 1 != len(state.failedDevices.Devices) {
		t.Error("Bad API error")
		t.Fail()
	}

	state.retryLoad()

	time.Sleep(1 * time.Second)
	if a.unloadCalled || 1 != len(state.extendedAPIs) || 0 != len(state.devices) || nil != state.failedDevices {
		t.Error("API was not loaded")
		t.Fail()
	}
}
