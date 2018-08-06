package group

import (
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/plugins/device/enums"
	"time"
)

// Tests group provider
func TestGroupProvider(t *testing.T) {
	var config = `
system: device
provider: group
name: cabinet lights
devices:
  - device1
  - device*
  - otherdevice
`

	invoked := 0

	s := mocks.FakeNewSettings(nil, false, nil, nil)
	f := s.FanOut()

	srv := mocks.FakeNewServer(func() {
		invoked++
	})

	ctor := &ConstructGroup{
		Settings: s,
		Server: srv,
		RawConfig: []byte(config),
	}

	prov, _ := NewGroupProvider(ctor)
	f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID:"device1",
	}

	time.Sleep(1 * time.Second)

	if len(prov.Devices()) != 0 {
		t.Error("Wrong device")
		t.FailNow()
	}

	srv.AddDevice(&providers.KnownDevice{
		Commands:[]string{"On"},
		Worker:"worker-1",
	})
	msg := &common.MsgDeviceUpdate{
		ID:"device1",
		Name:"device 1",
		Type:enums.DevLight,
		State:map[enums.Property]interface{} {enums.PropOn: true},
		FirstSeen:false,
	}

	f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)

	if 1 != len(prov.Devices()){
		t.Error("Device was not added")
		t.FailNow()
	}

	f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)

	if 1 != len(prov.Devices()){
		t.Error("Device was not added")
		t.FailNow()
	}

	f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID:"wrongdevice",
	}
	time.Sleep(1 * time.Second)
	f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID:"wrongdevice",
	}
	time.Sleep(1 * time.Second)

	if 1 != len(prov.Devices()){
		t.Error("Device was not added")
		t.FailNow()
	}

	if prov.ID() != "group.cabinet_lights" {
		t.Error("Wrong name")
		t.Fail()
	}

	prov.InvokeCommand(enums.CmdOn, nil)

	if 1 != invoked {
		t.Error("Not invoked")
		t.Fail()
	}
}
