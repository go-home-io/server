package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/providers"
)

type grSuite struct {
	suite.Suite

	prov providers.ILocationProvider
	f    providers.IInternalFanOutProvider
}

func (g *grSuite) SetupTest() {
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
	g.f = s.FanOut()

	ctor := &ConstructLocation{
		RawConfig: []byte(config),
		FanOut:    g.f,
	}

	g.prov, _ = NewLocationProvider(ctor)
}

// Tests empty device message.
func (g *grSuite) TestEmptyDevice() {
	g.f.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID: "device1",
	}

	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), 0, len(g.prov.Devices()))
}

// Tests double invoke.
func (g *grSuite) TestDoubleMessage() {
	msg := &common.MsgDeviceUpdate{
		ID:        "device1",
		Name:      "device 1",
		Type:      enums.DevLight,
		FirstSeen: true,
	}

	g.f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), 1, len(g.prov.Devices()), "first add")

	g.f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), 1, len(g.prov.Devices()), "second add")
}

// Tests double invoke with incorrect device.
func (g *grSuite) TestDoubleIncorrect() {
	msg := &common.MsgDeviceUpdate{
		ID:        "wrongdevice",
		FirstSeen: true,
	}

	g.f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), 0, len(g.prov.Devices()), "first add")

	g.f.ChannelInDeviceUpdates() <- msg
	time.Sleep(1 * time.Second)
	assert.Equal(g.T(), 0, len(g.prov.Devices()), "second add")
}

// Tests ID.
func (g *grSuite) TestID() {
	assert.Equal(g.T(), "cabinet loc", g.prov.ID())
}

// Tests location provider.
func TestGroupProvider(t *testing.T) {
	suite.Run(t, new(grSuite))
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
	assert.Error(t, err)
}
