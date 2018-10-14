package group

import (
	"testing"
	"time"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type grSuite struct {
	suite.Suite

	invoked int
	f       providers.IInternalFanOutProvider
	prov    providers.IGroupProvider
	srv     mocks.IFakeServer
}

func (g *grSuite) SetupTest() {
	var config = `
system: device
provider: group
name: cabinet lights
devices:
  - device1
  - device*
  - otherdevice
`
	s := mocks.FakeNewSettings(nil, false, nil, nil)
	g.f = s.FanOut()
	g.invoked = 0
	g.srv = mocks.FakeNewServer(func() {
		g.invoked++
	})

	ctor := &ConstructGroup{
		Settings:  s,
		Server:    g.srv.(providers.IServerProvider),
		RawConfig: []byte(config),
	}

	g.prov, _ = NewGroupProvider(ctor)

	g.f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID: "device1",
	}

	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), 0, len(g.prov.Devices()))

	g.srv.AddDevice(&providers.KnownDevice{
		Commands: []string{"On"},
		Worker:   "worker-1",
	})
}

func (g *grSuite) getMsg() *common.MsgDeviceUpdate {
	return &common.MsgDeviceUpdate{
		ID:        "device1",
		Name:      "device 1",
		Type:      enums.DevLight,
		State:     map[enums.Property]interface{}{enums.PropOn: true},
		FirstSeen: false,
	}
}

// Tests devices update.
func (g *grSuite) TestUpdate() {
	l := len(g.prov.Devices())
	msg := g.getMsg()

	g.f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), l+1, len(g.prov.Devices()), "not added on a first call")

	g.f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), l+1, len(g.prov.Devices()), "extra device")

	assert.Equal(g.T(), 0, g.invoked, "invokes mismatch")
}

// Tests that device update is not triggered with a wrong device.
func (g *grSuite) TestNoUpdatesWrongDevice() {
	l := len(g.prov.Devices())
	g.f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID: "wrongdevice",
	}

	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), l, len(g.prov.Devices()), "wrong device was added on a first call")

	g.f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID: "wrongdevice",
	}
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), l, len(g.prov.Devices()), "wrong device was added on a second call")

	assert.Equal(g.T(), 0, g.invoked, "invokes mismatch")
}

// Tests correct name.
func (g *grSuite) TestName() {
	assert.Equal(g.T(), "group.cabinet_lights", g.prov.ID())
}

// Tests correct command invoke.
func (g *grSuite) TestCommand() {
	g.f.ChannelInDeviceUpdates() <- g.getMsg()
	time.Sleep(1 * time.Second)
	g.prov.InvokeCommand(enums.CmdOn, nil)
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), 1, g.invoked, "invokes mismatch")
}

// Tests group provider.
func TestGroupProvider(t *testing.T) {
	suite.Run(t, new(grSuite))
}

// Tests wrong settings.
func TestWrongSettings(t *testing.T) {
	var config = `
system: device
provider: group
name: cabinet lights
devices: devices:
`

	s := mocks.FakeNewSettings(nil, false, nil, nil)
	srv := mocks.FakeNewServer(nil)

	ctor := &ConstructGroup{
		Settings:  s,
		Server:    srv.(providers.IServerProvider),
		RawConfig: []byte(config),
	}

	_, err := NewGroupProvider(ctor)
	assert.Error(t, err)
}
