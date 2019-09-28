package device

import (
	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/providers"
	"testing"
	"time"
)

var (
	excludedDevices = []enums.DeviceType{enums.DevUnknown, enums.DevHub, enums.DevGroup, enums.DevTrigger}
)

// Tests that all device types are known.
func TestExpectedDeviceTypes(t *testing.T) {
	for _, v := range enums.DeviceTypeValues() {
		if enums.SliceContainsDeviceType(excludedDevices, v) {
			continue
		}

		_, err := getExpectedType(v)
		assert.NoError(t, err, v.String())
	}
}

// Tests panic with a wrong device load.
func TestLoadDevice(t *testing.T) {
	for _, v := range enums.DeviceTypeValues() {
		if enums.SliceContainsDeviceType(excludedDevices, v) {
			continue
		}

		assert.Panics(t, func() { loadDevice(nil, v) }, v.String())
	}
}

// Tests device update.
func TestDeviceUpdateChan(t *testing.T) {
	monkey.Patch(newDeviceProcessor, func(_ enums.DeviceType, _ string) IProcessor { return nil })
	defer monkey.UnpatchAll()

	c := &fakeCamera{
		initError:   nil,
		updateError: nil,
		state: &device.CameraState{
			Picture: "test1",
		},
		spec: &device.Spec{
			UpdatePeriod:           0,
			SupportedCommands:      []enums.Command{enums.CmdTakePicture},
			SupportedProperties:    []enums.Property{enums.PropPicture},
			PostCommandDeferUpdate: 0,
		},
		name: "",
	}

	s := mocks.FakeNewSettings(nil, true, []*providers.RawDevice{}, nil)
	s.(mocks.IFakeSettings).AddLoader(c)

	ctor := &ConstructDevice{
		DeviceName:        "test",
		DeviceType:        enums.DevCamera,
		ConfigName:        "test",
		RawConfig:         "",
		Settings:          s,
		UOM:               enums.UOMImperial,
		StatusUpdatesChan: nil,
		DiscoveryChan:     nil,
	}

	devs, err := LoadDevice(ctor)
	assert.NoError(t, err, "failed to load device")
	assert.Equal(t, 1, len(devs), "wrong device count")

	p := devs[0]

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "test1", p.GetUpdateMessage().State["picture"], "initial picture is incorrect")

	c.updateChan <- &device.StateUpdateData{State:&device.CameraState{
		Picture:            "test2",
	}}

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "test2", p.GetUpdateMessage().State["picture"], "updated picture is incorrect")
}
