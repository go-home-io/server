package device

import (
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
)

type fakeCamera struct {
	initError   error
	updateError error
	state       *device.CameraState
	spec        *device.Spec
	name        string

	cmdInvoked    int
	updateInvoked int
	unloadInvoked int

	receivedSpeed uint8

	updateChan chan *device.StateUpdateData
}

func (f *fakeCamera) FakeInit(i interface{}) {
	f.updateChan = i.(*device.InitDataDevice).DeviceStateUpdateChan
}

func (f *fakeCamera) Init(d *device.InitDataDevice) error {
	f.updateChan = d.DeviceStateUpdateChan
	return nil
}

func (f *fakeCamera) Unload() {
	f.unloadInvoked++
}

func (f *fakeCamera) GetName() string {
	return f.name
}

func (f *fakeCamera) GetSpec() *device.Spec {
	return f.spec
}

func (f *fakeCamera) Input(common.Input) error {
	return nil
}

func (f *fakeCamera) Load() (*device.CameraState, error) {
	return f.state, f.updateError
}

func (f *fakeCamera) Update() (*device.CameraState, error) {
	f.updateInvoked++
	return f.state, f.updateError
}

func (f *fakeCamera) TakePicture() error {
	f.cmdInvoked++
	return nil
}

func (f *fakeCamera) On(_, _ bool) error {
	return nil
}

func (f *fakeCamera) Off() error {
	f.cmdInvoked++
	return errors.New("test")
}

func (f *fakeCamera) SetFanSpeed(p common.Percent) {
	f.cmdInvoked++
	f.receivedSpeed = p.Value
}

func getWrapper(isValidatorFailing bool, camera *fakeCamera) IDeviceWrapperProvider {
	ctor := &wrapperConstruct{
		DeviceType:       enums.DevCamera,
		DeviceConfigName: "test 1",
		DeviceProvider:   "test",
		DeviceInterface:  camera,
		DeviceState:      nil,
		Logger:           mocks.FakeNewLogger(nil),
		SystemLogger:     mocks.FakeNewLogger(nil),
		Secret:           mocks.FakeNewSecretStore(nil, false),
		WorkerID:         "test",
		LoadData: &device.InitDataDevice{
			Logger:                mocks.FakeNewLogger(nil),
			Secret:                mocks.FakeNewSecretStore(nil, false),
			UOM:                   0,
			DeviceStateUpdateChan: make(chan *device.StateUpdateData, 5),
			DeviceDiscoveredChan:  nil,
		},
		IsHubDevice:       false,
		Validator:         mocks.FakeNewValidator(!isValidatorFailing),
		UOM:               enums.UOMImperial,
		processor:         nil,
		RawConfig:         "",
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     nil,
	}

	return NewDeviceWrapper(ctor)
}

// Tests device generated name.
func TestName(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec:        nil,
		name:        "",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	assert.Equal(t, "test_1.camera", p.ID(), "ID is wrong")
	assert.Equal(t, "test 1", p.Name(), "name is wrong")
}

// Tests device generated name with a plugin name.
func TestNameWithPluginName(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec:        nil,
		name:        "test",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	assert.Equal(t, "test_1.camera.test", p.ID(), "ID is wrong")
	assert.Equal(t, "test 1", p.Name(), "name is wrong")
}

// Tests name without a config override.
func TestNameNoOverride(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec:        nil,
		name:        "",
	}

	ctor := &wrapperConstruct{
		DeviceType:      enums.DevCamera,
		DeviceProvider:  "test",
		DeviceInterface: c,
		DeviceState:     nil,
		Logger:          mocks.FakeNewLogger(nil),
		SystemLogger:    mocks.FakeNewLogger(nil),
		Secret:          mocks.FakeNewSecretStore(nil, false),
		WorkerID:        "test",
		LoadData: &device.InitDataDevice{
			Logger:                mocks.FakeNewLogger(nil),
			Secret:                mocks.FakeNewSecretStore(nil, false),
			UOM:                   0,
			DeviceStateUpdateChan: make(chan *device.StateUpdateData, 5),
			DeviceDiscoveredChan:  nil,
		},
		IsHubDevice:       false,
		Validator:         mocks.FakeNewValidator(true),
		UOM:               enums.UOMImperial,
		processor:         nil,
		RawConfig:         "",
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     nil,
	}

	p := NewDeviceWrapper(ctor)

	assert.NotNil(t, p, "nil wrapper")
}

// Tests wrong cmd execution.
func TestWrongCmd(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec:        nil,
		name:        "",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	p.InvokeCommand(enums.CmdOn, nil)
	assert.Equal(t, 0, c.cmdInvoked)
}

// Tests execution of a correct command with wrong arguments.
func TestWrongCmdAndArguments(t *testing.T) {
	enums.AllowedCommands = map[enums.DeviceType][]enums.Command{
		enums.DevCamera: {enums.CmdTakePicture},
	}

	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec: &device.Spec{
			UpdatePeriod:           1 * time.Second,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture, enums.CmdDock},
			SupportedProperties:    nil,
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	p.InvokeCommand(enums.CmdOn, map[string]interface{}{"test": "test"})
	assert.Equal(t, 0, c.cmdInvoked, "Wrong arguments passed")

	p.InvokeCommand(enums.CmdDock, nil)
	assert.Equal(t, 0, c.cmdInvoked)
}

// Tests non implemented command invocation.
func TestNonImplementedCmd(t *testing.T) {
	enums.AllowedCommands = map[enums.DeviceType][]enums.Command{
		enums.DevCamera: {enums.CmdDock, enums.CmdOn},
	}

	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec: &device.Spec{
			UpdatePeriod:           1 * time.Second,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture, enums.CmdDock, enums.CmdOn},
			SupportedProperties:    nil,
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	p.InvokeCommand(enums.CmdDock, nil)
	assert.Equal(t, 0, c.cmdInvoked, "dock invoked")

	p.InvokeCommand(enums.CmdOn, nil)
	assert.Equal(t, 0, c.cmdInvoked, "on invoked")
}

// Tests correct command with no params execution.
func TestCorrectCmdNoParams(t *testing.T) {
	enums.AllowedCommands = map[enums.DeviceType][]enums.Command{
		enums.DevCamera: {enums.CmdTakePicture},
	}

	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec: &device.Spec{
			UpdatePeriod:           1 * time.Second,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture},
			SupportedProperties:    nil,
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	p.InvokeCommand(enums.CmdTakePicture, nil)
	assert.Equal(t, 1, c.cmdInvoked)
}

// Tests correct command with params execution.
func TestCorrectCmdWithParams(t *testing.T) {
	enums.AllowedCommands = map[enums.DeviceType][]enums.Command{
		enums.DevCamera: {enums.CmdSetFanSpeed},
	}

	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec: &device.Spec{
			UpdatePeriod:           1 * time.Second,
			SupportedCommands:      []enums.Command{enums.CmdSetFanSpeed},
			SupportedProperties:    nil,
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	p.InvokeCommand(enums.CmdSetFanSpeed, nil)
	assert.Equal(t, 0, c.cmdInvoked, "wrong params invoked")

	p.InvokeCommand(enums.CmdSetFanSpeed, map[string]interface{}{"value": 10})
	assert.Equal(t, 0, c.cmdInvoked, "correct params was invoked with failed validator")

	p = getWrapper(false, c)

	p.InvokeCommand(enums.CmdSetFanSpeed, map[string]interface{}{"value": 10})
	assert.Equal(t, 1, c.cmdInvoked, "correct params was not invoked")

	assert.Equal(t, uint8(10), c.receivedSpeed, "wrong param conversion")
}

// Test correct errors handling.
func TestCorrectCmdError(t *testing.T) {
	enums.AllowedCommands = map[enums.DeviceType][]enums.Command{
		enums.DevCamera: {enums.CmdOff},
	}

	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec: &device.Spec{
			UpdatePeriod:           1 * time.Second,
			SupportedCommands:      []enums.Command{enums.CmdOff},
			SupportedProperties:    nil,
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(false, c)
	assert.NotNil(t, p, "nil wrapper")

	p.InvokeCommand(enums.CmdOff, nil)
	assert.Equal(t, 1, c.cmdInvoked, "command was not called")
	assert.Equal(t, 0, c.updateInvoked, "update was called")
}

// Tests defer update.
func TestDeferUpdateCmd(t *testing.T) {
	enums.AllowedCommands = map[enums.DeviceType][]enums.Command{
		enums.DevCamera: {enums.CmdTakePicture},
	}

	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec: &device.Spec{
			UpdatePeriod:           1 * time.Second,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture},
			SupportedProperties:    nil,
			PostCommandDeferUpdate: 100 * time.Millisecond,
		},
		name: "",
	}

	p := getWrapper(false, c)
	assert.NotNil(t, p, "nil wrapper")

	p.InvokeCommand(enums.CmdTakePicture, nil)
	assert.Equal(t, 1, c.cmdInvoked, "command was not called")
	assert.Equal(t, 1, c.updateInvoked, "update was not called")
}

// Tests simple field value actual value.
func TestGetFieldValueOrNilSimple(t *testing.T) {
	w := reflect.ValueOf(struct {
		St string
	}{
		St: "test",
	})

	v := getFieldValueOrNil(w.Field(0))

	assert.Equal(t, "test", v)
}

// Tests empty string field value actual value.
func TestGetFieldValueOrNilEmptyString(t *testing.T) {
	w := reflect.ValueOf(struct {
		St string
	}{
		St: "",
	})

	v := getFieldValueOrNil(w.Field(0))

	assert.Nil(t, v)
}

// Tests empty object value actual value.
func TestGetFieldValueOrNilNilObject(t *testing.T) {
	w := reflect.ValueOf(struct {
		St interface{}
	}{
		St: nil,
	})

	v := getFieldValueOrNil(w.Field(0))

	assert.Nil(t, v)
}

// Tests that update was called and results are processed.
func TestUpdateCalledAndProcessed(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state: &device.CameraState{
			GenericDeviceState: device.GenericDeviceState{
				Input: &common.Input{
					Title:  "",
					Params: map[string]string{"k1": "v1", "k2": ""},
				},
			},
			Picture:  "test",
			Distance: 10,
		},
		spec: &device.Spec{
			UpdatePeriod:           100 * time.Millisecond,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture},
			SupportedProperties:    []enums.Property{enums.PropPicture, enums.PropInput},
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(false, c)
	assert.NotNil(t, p, "nil wrapper")

	time.Sleep(120 * time.Millisecond)

	assert.Equal(t, 1, c.updateInvoked, "update was not called")

	m := p.GetUpdateMessage()

	v, ok := m.State["picture"]
	assert.True(t, ok, "picture was not processed")
	assert.Equal(t, "test", v, "picture has incorrect data")

	_, ok = m.State["distance"]
	assert.False(t, ok, "distance was processed")

	v, ok = m.State["input"]
	assert.True(t, ok, "input was not processed")
	assert.Equal(t, "v1", v.(*common.Input).Params["k1"], "input is incorrect")
}

// Tests that update was invoked but nil value was not processed.
func TestUpdateCalledAndProcessedWithNilValues(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state: &device.CameraState{
			Picture:  "",
			Distance: 10,
		},
		spec: &device.Spec{
			UpdatePeriod:           100 * time.Millisecond,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture},
			SupportedProperties:    []enums.Property{enums.PropPicture, enums.PropInput},
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(false, c)
	assert.NotNil(t, p, "nil wrapper")

	time.Sleep(120 * time.Millisecond)

	assert.Equal(t, 1, c.updateInvoked, "update was not called")

	m := p.GetUpdateMessage()

	_, ok := m.State["picture"]

	assert.False(t, ok)
}

type fakeCameraProcessor struct {
}

func (f fakeCameraProcessor) IsExtraProperty(p enums.Property) bool {
	return p == enums.PropDistance
}

func (f fakeCameraProcessor) GetExtraSupportPropertiesSpec() []enums.Property {
	return []enums.Property{enums.PropDistance}
}

func (f fakeCameraProcessor) IsPropertyGood(p enums.Property, v interface{}) (bool, map[enums.Property]interface{}) {
	if p == enums.PropInput {
		return false, nil
	}

	if p != enums.PropPicture {
		return true, map[enums.Property]interface{}{p: v}
	}

	return true, map[enums.Property]interface{}{
		enums.PropDistance: 1,
		enums.PropPicture:  "test1",
	}
}

// Tests whether update ignores incorrect state.
func TestUpdateStateWithWrongStructure(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state: &device.CameraState{
			Picture:  "",
			Distance: 10,
		},
		spec: &device.Spec{
			UpdatePeriod:           100 * time.Millisecond,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture},
			SupportedProperties:    []enums.Property{enums.PropPicture},
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	p := getWrapper(false, c)
	assert.NotNil(t, p, "nil wrapper")

	p.(*deviceWrapper).processUpdate(&struct {
		On      bool
		Wrong   string `json:"wrong"`
		Picture string `json:"picture"`
	}{
		On:      true,
		Wrong:   "test",
		Picture: "test",
	})

	m := p.GetUpdateMessage()

	_, ok := m.State["on"]
	assert.False(t, ok, "on field is present")

	_, ok = m.State["wrong"]
	assert.False(t, ok, "wrong field is present")

	v, ok := m.State["picture"]
	assert.True(t, ok, "picture field is not present")
	assert.Equal(t, "test", v, "picture has incorrect value")
}

// Tests whether processors are correctly invoked.
func TestProcessors(t *testing.T) {
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state: &device.CameraState{
			GenericDeviceState: device.GenericDeviceState{
				Input: &common.Input{
					Title:  "",
					Params: map[string]string{"k1": "v1", "k2": ""},
				},
			},
			Picture:  "test",
			Distance: 10,
		},
		spec: &device.Spec{
			UpdatePeriod:           100 * time.Millisecond,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture},
			SupportedProperties:    []enums.Property{enums.PropPicture, enums.PropInput},
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	ctor := &wrapperConstruct{
		DeviceType:       enums.DevCamera,
		DeviceConfigName: "test 1",
		DeviceProvider:   "test",
		DeviceInterface:  c,
		DeviceState:      nil,
		Logger:           mocks.FakeNewLogger(nil),
		SystemLogger:     mocks.FakeNewLogger(nil),
		Secret:           mocks.FakeNewSecretStore(nil, false),
		WorkerID:         "test",
		LoadData: &device.InitDataDevice{
			Logger:                mocks.FakeNewLogger(nil),
			Secret:                mocks.FakeNewSecretStore(nil, false),
			UOM:                   0,
			DeviceStateUpdateChan: make(chan *device.StateUpdateData, 5),
			DeviceDiscoveredChan:  nil,
		},
		IsHubDevice:       false,
		Validator:         mocks.FakeNewValidator(true),
		UOM:               enums.UOMImperial,
		processor:         &fakeCameraProcessor{},
		RawConfig:         "",
		StatusUpdatesChan: make(chan *UpdateEvent, 5),
		DiscoveryChan:     nil,
	}

	p := NewDeviceWrapper(ctor)
	assert.NotNil(t, p, "nil wrapper")

	time.Sleep(120 * time.Millisecond)
	assert.Equal(t, 1, c.updateInvoked, "update was not called")

	m := p.GetUpdateMessage()

	assert.Equal(t, 1, m.State["distance"], "distance is wrong")
	assert.Equal(t, "test1", m.State["picture"], "picture is wrong")

	_, ok := m.State["input"]
	assert.False(t, ok, "input is present")
}

// Tests whether Unload being called.
func TestUnload(t *testing.T){
	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state:       nil,
		spec:        nil,
		name:        "test",
	}

	p := getWrapper(true, c)
	assert.NotNil(t, p, "nil wrapper")

	p.AppendChild(p)

	p.Unload()
	assert.Equal(t, 1, c.unloadInvoked, "unload was not called")
}