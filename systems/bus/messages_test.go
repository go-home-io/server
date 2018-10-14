package bus

import (
	"math"
	"testing"

	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/utils"
	"github.com/stretchr/testify/assert"
)

func checkTime(t *testing.T, sendTime int64) {
	assert.False(t, math.Abs(float64(sendTime)-float64(utils.TimeNow())) > 1)
}

// Test discovery ctor.
func TestNewDiscoveryMessage(t *testing.T) {
	m := NewDiscoveryMessage("test", true, map[string]string{"test": "data"}, 100)
	checkTime(t, m.SendTime)
	assert.Equal(t, 1, len(m.Properties))
}

// Tests device assignment ctor.
func TestNewDeviceAssignmentMessage(t *testing.T) {
	m := NewDeviceAssignmentMessage([]*DeviceAssignment{{Name: "test"}}, enums.UOMMetric)
	checkTime(t, m.SendTime)
	assert.Equal(t, 1, len(m.Devices))
}

// Tests device update ctor.
func TestNewDeviceUpdateMessage(t *testing.T) {
	m := NewDeviceUpdateMessage()
	checkTime(t, m.SendTime)
}

// Tests device command ctor.
func TestNewDeviceCommandMessage(t *testing.T) {
	m := NewDeviceCommandMessage("test", enums.CmdOn, map[string]interface{}{"test": 1})
	checkTime(t, m.SendTime)
	assert.Equal(t, 1, len(m.Payload))
}
