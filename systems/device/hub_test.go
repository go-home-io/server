package device

import (
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/providers"
)

type fakeHub struct {
	spec         *device.Spec
	loadResult   *device.HubLoadResult
	updateResult *device.HubLoadResult

	unloadInvoked int

	discoveredChan chan *device.DiscoveredDevices
}

func (f *fakeHub) FakeInit(i interface{}) {
	f.discoveredChan = i.(*device.InitDataDevice).DeviceDiscoveredChan
}

func (f *fakeHub) Init(*device.InitDataDevice) error {
	return nil
}

func (f *fakeHub) Unload() {
	f.unloadInvoked++
}

func (f *fakeHub) GetName() string {
	return "test"
}

func (f *fakeHub) GetSpec() *device.Spec {
	return f.spec
}

func (f *fakeHub) Input(common.Input) error {
	return nil
}

func (f *fakeHub) Load() (*device.HubLoadResult, error) {
	return f.loadResult, nil
}

func (f *fakeHub) Update() (*device.HubLoadResult, error) {
	return f.updateResult, nil
}

// Tests whether unload was invoked after the load.
func TestUnloadInvokedDuringLoad(t *testing.T) {
	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	cam := &fakeCamera{
		initError:     nil,
		updateError:   nil,
		state:         &device.CameraState{Picture: "test"},
		spec:          nil,
		name:          "dev123",
		cmdInvoked:    0,
		updateInvoked: 0,
		receivedSpeed: 0,
	}

	f := &fakeHub{
		spec: &device.Spec{},
		loadResult: &device.HubLoadResult{
			State: &device.HubState{
				GenericDeviceState: device.GenericDeviceState{},
				NumDevices:         1,
			},
			Devices: []*device.DiscoveredDevices{
				{
					Type:      enums.DevCamera,
					Interface: cam,
					State:     &device.CameraState{Picture: "test"},
				},
			},
		},
		updateResult: nil,
	}

	s.(mocks.IFakeSettings).AddLoader(f)

	ctor := &ConstructDevice{
		DeviceName:        "",
		DeviceType:        enums.DevHub,
		ConfigName:        "hub1",
		RawConfig:         "",
		Settings:          s,
		UOM:               enums.UOMMetric,
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     make(chan *NewDeviceDiscoveredEvent, 5),
	}

	devs, err := LoadDevice(ctor)

	assert.NoError(t, err, "hub load failed")
	assert.Equal(t, 2, len(devs), "incorrect devices number")

	devs[0].Unload()

	assert.Equal(t, 1, cam.unloadInvoked, "camera unload was not called")
	assert.Equal(t, 1, f.unloadInvoked, "hub unload was not called")
}

// Tests whether unload was invoked after the update.
func TestUnloadInvokedDuringUpdate(t *testing.T) {
	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	cam := &fakeCamera{
		initError:     nil,
		updateError:   nil,
		state:         &device.CameraState{Picture: "test"},
		spec:          nil,
		name:          "dev123",
		cmdInvoked:    0,
		updateInvoked: 0,
		receivedSpeed: 0,
	}

	f := &fakeHub{
		spec: &device.Spec{
			UpdatePeriod: 100 * time.Millisecond,
		},
		updateResult: &device.HubLoadResult{
			State: &device.HubState{
				GenericDeviceState: device.GenericDeviceState{},
				NumDevices:         1,
			},
			Devices: []*device.DiscoveredDevices{
				{
					Type:      enums.DevCamera,
					Interface: cam,
					State:     &device.CameraState{Picture: "test"},
				},
			},
		},
		loadResult: &device.HubLoadResult{},
	}

	s.(mocks.IFakeSettings).AddLoader(f)

	ctor := &ConstructDevice{
		DeviceName:        "",
		DeviceType:        enums.DevHub,
		ConfigName:        "hub1",
		RawConfig:         "",
		Settings:          s,
		UOM:               enums.UOMMetric,
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     make(chan *NewDeviceDiscoveredEvent, 5),
	}

	devs, err := LoadDevice(ctor)

	assert.NoError(t, err, "hub load failed")
	assert.Equal(t, 1, len(devs), "incorrect devices number")

	time.Sleep(120 * time.Millisecond)

	devs[0].Unload()

	assert.Equal(t, 1, cam.unloadInvoked, "camera unload was not called")
	assert.Equal(t, 1, f.unloadInvoked, "hub unload was not called")
}

// Tests name overrides after the load.
func TestNameOverridesDuringLoad(t *testing.T) {
	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	cam := &fakeCamera{
		initError:     nil,
		updateError:   nil,
		state:         &device.CameraState{Picture: "test"},
		spec:          nil,
		name:          "dev123",
		cmdInvoked:    0,
		updateInvoked: 0,
		receivedSpeed: 0,
	}

	f := &fakeHub{
		spec: &device.Spec{},
		loadResult: &device.HubLoadResult{
			State: &device.HubState{
				GenericDeviceState: device.GenericDeviceState{},
				NumDevices:         1,
			},
			Devices: []*device.DiscoveredDevices{
				{
					Type:      enums.DevCamera,
					Interface: cam,
					State:     &device.CameraState{Picture: "test"},
				},
			},
		},
		updateResult: nil,
	}

	s.(mocks.IFakeSettings).AddLoader(f)

	ctor := &ConstructDevice{
		DeviceName: "",
		DeviceType: enums.DevHub,
		ConfigName: "hub1",
		RawConfig: `
nameOverrides: 
  "*dev*": test1
  "dev2*": test2
`,
		Settings:          s,
		UOM:               enums.UOMMetric,
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     make(chan *NewDeviceDiscoveredEvent, 5),
	}

	devs, err := LoadDevice(ctor)

	assert.NoError(t, err, "hub load failed")
	assert.Equal(t, 2, len(devs), "incorrect devices number")

	assert.Equal(t, "hub1", devs[0].Name(), "hub name is incorrect")
	assert.Equal(t, "hub1.hub.test", devs[0].ID(), "hub ID is incorrect")

	assert.Equal(t, "test1", devs[1].Name(), "camera name is incorrect")
	assert.Equal(t, "hub1.camera.dev123", devs[1].ID(), "camera ID is incorrect")
}

// Tests name overrides after the update.
func TestNameOverridesDuringUpdate(t *testing.T) {
	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	cam := &fakeCamera{
		initError:     nil,
		updateError:   nil,
		state:         &device.CameraState{Picture: "test"},
		spec:          nil,
		name:          "dev123",
		cmdInvoked:    0,
		updateInvoked: 0,
		receivedSpeed: 0,
	}

	f := &fakeHub{
		spec: &device.Spec{
			UpdatePeriod: 100 * time.Millisecond,
		},
		updateResult: &device.HubLoadResult{
			State: &device.HubState{
				GenericDeviceState: device.GenericDeviceState{},
				NumDevices:         1,
			},
			Devices: []*device.DiscoveredDevices{
				{
					Type:      enums.DevCamera,
					Interface: cam,
					State:     &device.CameraState{Picture: "test"},
				},
			},
		},
		loadResult: &device.HubLoadResult{},
	}

	s.(mocks.IFakeSettings).AddLoader(f)

	dc := make(chan *NewDeviceDiscoveredEvent, 5)

	ctor := &ConstructDevice{
		DeviceName: "",
		DeviceType: enums.DevHub,
		ConfigName: "hub1",
		RawConfig: `
nameOverrides: 
  "*dev*": test1
  "dev2*": test2
`,
		Settings:          s,
		UOM:               enums.UOMMetric,
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     dc,
	}

	devs, err := LoadDevice(ctor)

	assert.NoError(t, err, "hub load failed")
	assert.Equal(t, 1, len(devs), "incorrect devices number")

	assert.Equal(t, "hub1", devs[0].Name(), "hub name is incorrect")
	assert.Equal(t, "hub1.hub.test", devs[0].ID(), "hub ID is incorrect")

	timeout := time.After(120 * time.Millisecond)

	select {
	case <-timeout:
		t.Fatal("didn't get device update")
	case c := <-dc:
		{
			assert.Equal(t, "test1", c.Provider.Name(), "camera name is incorrect")
			assert.Equal(t, "hub1.camera.dev123", c.Provider.ID(), "camera ID is incorrect")
			return
		}
	}
}

// Tests proper device update.
func TestHubUpdate(t *testing.T) {
	monkey.Patch(newDeviceProcessor, func(_ enums.DeviceType, _ string) IProcessor { return nil })
	defer monkey.UnpatchAll()

	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	cam := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       &device.CameraState{Picture: "test1"},
		spec: &device.Spec{
			UpdatePeriod:        100 * time.Millisecond,
			SupportedProperties: []enums.Property{enums.PropPicture},
		},
		name:          "dev123",
		cmdInvoked:    0,
		updateInvoked: 0,
		receivedSpeed: 0,
	}

	f := &fakeHub{
		spec: &device.Spec{},
		loadResult: &device.HubLoadResult{
			State: &device.HubState{
				GenericDeviceState: device.GenericDeviceState{},
				NumDevices:         1,
			},
			Devices: []*device.DiscoveredDevices{
				{
					Type:      enums.DevCamera,
					Interface: cam,
					State:     &device.CameraState{Picture: "test1"},
				},
			},
		},
		updateResult: nil,
	}

	s.(mocks.IFakeSettings).AddLoader(f)

	ctor := &ConstructDevice{
		DeviceName:        "",
		DeviceType:        enums.DevHub,
		ConfigName:        "hub1",
		RawConfig:         "",
		Settings:          s,
		UOM:               enums.UOMMetric,
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     make(chan *NewDeviceDiscoveredEvent, 5),
	}

	devs, err := LoadDevice(ctor)

	assert.NoError(t, err, "hub load failed")
	assert.Equal(t, 2, len(devs), "incorrect devices number")

	assert.Equal(t, "test1", devs[1].GetUpdateMessage().State["picture"], "initial state is incorrect")

	cam.state.Picture = "test2"
	time.Sleep(120 * time.Millisecond)

	assert.Equal(t, "test2", devs[1].GetUpdateMessage().State["picture"], "updated state is incorrect")
}

// Tests proper device update channeling.
func TestHubUpdateChan(t *testing.T) {
	monkey.Patch(newDeviceProcessor, func(_ enums.DeviceType, _ string) IProcessor { return nil })
	defer monkey.UnpatchAll()

	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	cam := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       &device.CameraState{Picture: "test1"},
		spec: &device.Spec{
			UpdatePeriod:        0,
			SupportedProperties: []enums.Property{enums.PropPicture},
		},
		name:          "dev123",
		cmdInvoked:    0,
		updateInvoked: 0,
		receivedSpeed: 0,
	}

	f := &fakeHub{
		spec: &device.Spec{},
		loadResult: &device.HubLoadResult{
			State: &device.HubState{
				GenericDeviceState: device.GenericDeviceState{},
				NumDevices:         1,
			},
			Devices: []*device.DiscoveredDevices{
				{
					Type:      enums.DevCamera,
					Interface: cam,
					State:     &device.CameraState{Picture: "test1"},
				},
			},
		},
		updateResult: nil,
	}

	s.(mocks.IFakeSettings).AddLoader(f)

	ctor := &ConstructDevice{
		DeviceName:        "",
		DeviceType:        enums.DevHub,
		ConfigName:        "hub1",
		RawConfig:         "",
		Settings:          s,
		UOM:               enums.UOMMetric,
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     make(chan *NewDeviceDiscoveredEvent, 5),
	}

	devs, err := LoadDevice(ctor)

	assert.NoError(t, err, "hub load failed")
	assert.Equal(t, 2, len(devs), "incorrect devices number")

	assert.Equal(t, "test1", devs[1].GetUpdateMessage().State["picture"], "initial state is incorrect")

	cam.updateChan <- &device.StateUpdateData{State: &device.CameraState{Picture: "test2"}}

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "test2", devs[1].GetUpdateMessage().State["picture"], "updated state is incorrect")
}

// Tests proper discovery chan operations.
func TestUpdateAndUnloadAfterDiscoveryThroughChan(t *testing.T) {
	monkey.Patch(newDeviceProcessor, func(_ enums.DeviceType, _ string) IProcessor { return nil })
	defer monkey.UnpatchAll()

	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	cam := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       &device.CameraState{Picture: "test1"},
		spec: &device.Spec{
			UpdatePeriod:        0,
			SupportedProperties: []enums.Property{enums.PropPicture},
		},
		name:          "dev123",
		cmdInvoked:    0,
		updateInvoked: 0,
		receivedSpeed: 0,
	}

	f := &fakeHub{
		spec: &device.Spec{},
		loadResult: &device.HubLoadResult{
			State: &device.HubState{
				NumDevices: 0,
			},
			Devices: []*device.DiscoveredDevices{},
		},
		updateResult: nil,
	}

	s.(mocks.IFakeSettings).AddLoader(f)

	dc := make(chan *NewDeviceDiscoveredEvent, 5)

	ctor := &ConstructDevice{
		DeviceName:        "",
		DeviceType:        enums.DevHub,
		ConfigName:        "hub1",
		RawConfig:         "",
		Settings:          s,
		UOM:               enums.UOMMetric,
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     dc,
	}

	devs, err := LoadDevice(ctor)

	assert.NoError(t, err, "hub load failed")
	assert.Equal(t, 1, len(devs), "incorrect devices number")

	f.discoveredChan <- &device.DiscoveredDevices{
		Type:      enums.DevCamera,
		Interface: cam,
		State:     &device.CameraState{Picture: "test1"},
	}

	timeout := time.After(100 * time.Millisecond)
	var discWrapper IDeviceWrapperProvider = nil
	select {
	case <-timeout:
		t.Fatal("timeout waiting for the device")
	case c := <-dc:
		{
			discWrapper = c.Provider
			assert.Equal(t, "test1",
				discWrapper.GetUpdateMessage().State["picture"], "initial state is incorrect")
			break
		}
	}

	cam.updateChan <- &device.StateUpdateData{State: &device.CameraState{Picture: "test2"}}

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "test2",
		discWrapper.GetUpdateMessage().State["picture"], "updated state is incorrect")

	devs[0].Unload()

	assert.Equal(t, 1, cam.unloadInvoked, "camera unload was not called")
	assert.Equal(t, 1, f.unloadInvoked, "hub unload was not called")
}
