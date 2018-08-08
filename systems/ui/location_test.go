package ui

import (
	"testing"
	"time"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
)

// Tests location provider.
func TestGroupProvider(t *testing.T) {
	var config = `
system: ui
provider: location
name: cabinet loc
devices:
  - device1
  - device*
  - otherdevice
`
	s := mocks.FakeNewSettings(nil, false, nil, nil)
	f := s.FanOut()

	ctor := &ConstructLocation{
		RawConfig: []byte(config),
		FanOut:    f,
	}

	prov, _ := NewLocationProvider(ctor)
	f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID: "device1",
	}

	time.Sleep(1 * time.Second)

	if len(prov.Devices()) != 0 {
		t.Error("Wrong device")
		t.FailNow()
	}

	msg := &common.MsgDeviceUpdate{
		ID:        "device1",
		Name:      "device 1",
		Type:      enums.DevLight,
		FirstSeen: true,
	}

	f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)

	if 1 != len(prov.Devices()) {
		t.Error("Device was not added")
		t.FailNow()
	}

	f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)

	if 1 != len(prov.Devices()) {
		t.Error("Device was not added")
		t.FailNow()
	}

	f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID:        "wrongdevice",
		FirstSeen: true,
	}
	time.Sleep(1 * time.Second)
	f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID:        "wrongdevice",
		FirstSeen: true,
	}
	time.Sleep(1 * time.Second)

	if 1 != len(prov.Devices()) {
		t.Error("Device was not added")
		t.FailNow()
	}

	if prov.ID() != "cabinet loc" {
		t.Error("Wrong name")
		t.Fail()
	}
}

// Test wrong config.
func TestWrongSettings(t *testing.T) {
	var config = `ad`
	s := mocks.FakeNewSettings(nil, false, nil, nil)
	f := s.FanOut()

	ctor := &ConstructLocation{
		RawConfig: []byte(config),
		FanOut:    f,
		Logger:    mocks.FakeNewLogger(nil),
	}

	_, err := NewLocationProvider(ctor)
	if err == nil {
		t.Fail()
	}
}
