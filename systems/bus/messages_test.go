package bus

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/utils"
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

// Tests entity load ctor.
func TestNewEntityLoadStatusMessage(t *testing.T) {
	m := NewEntityLoadStatusMessage("test", "test_node", true)
	checkTime(t, m.SendTime)
	assert.Equal(t, "test", m.Name, "name")
	assert.Equal(t, "test_node", m.NodeID, "node")
	assert.True(t, m.IsSuccess, "success")
}
