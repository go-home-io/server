package bus

import (
	"math"
	"testing"

	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/utils"
)

// Test discovery ctor.
func TestNewDiscoveryMessage(t *testing.T) {
	m := NewDiscoveryMessage("test", true, map[string]string{"test": "data"}, 100)
	if math.Abs(float64(m.SendTime)-float64(utils.TimeNow())) > 1 || len(m.Properties) != 1 {
		t.Fail()
	}
}

// Tests device assignment ctor.
func TestNewDeviceAssignmentMessage(t *testing.T) {
	m := NewDeviceAssignmentMessage([]*DeviceAssignment{&DeviceAssignment{Name: "test"}}, enums.UOMMetric)
	if math.Abs(float64(m.SendTime)-float64(utils.TimeNow())) > 1 || len(m.Devices) != 1 {
		t.Fail()
	}
}

// Tests device update ctor.
func TestNewDeviceUpdateMessage(t *testing.T) {
	m := NewDeviceUpdateMessage()
	if math.Abs(float64(m.SendTime)-float64(utils.TimeNow())) > 1 {
		t.Fail()
	}
}

// Tests device command ctor.
func TestNewDeviceCommandMessage(t *testing.T) {
	m := NewDeviceCommandMessage("test", enums.CmdOn, map[string]interface{}{"test": 1})
	if math.Abs(float64(m.SendTime)-float64(utils.TimeNow())) > 1 || len(m.Payload) != 1 {
		t.Fail()
	}
}
