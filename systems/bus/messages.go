package bus

import (
	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/utils"
)

// MessageWithType helper type for initial service bus message parsing.
type MessageWithType struct {
	Type     bus.MessageType `json:"mt"`
	SendTime int64
}

// KeyValue helper type for key-value pair.
type KeyValue struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

// DiscoveryMessage used by worker for periodic server pings.
type DiscoveryMessage struct {
	MessageWithType
	NodeID       string            `json:"n"`
	Properties   map[string]string `json:"p"`
	IsFirstStart bool              `json:"f"`
	MaxDevices   int               `json:"m"`
}

// DeviceAssignment type with single device assignment.
type DeviceAssignment struct {
	Plugin string           `json:"p"`
	Type   enums.DeviceType `json:"t"`
	Config string           `json:"c"`
	Name   string           `json:"n"`
}

// DeviceAssignmentMessage used by server to send a new set of devices to worker.
type DeviceAssignmentMessage struct {
	MessageWithType
	Devices []*DeviceAssignment `json:"d"`
}

// DeviceUpdateMessage used by worker to update service with devices state update.
type DeviceUpdateMessage struct {
	MessageWithType
	DeviceType enums.DeviceType       `json:"t"`
	DeviceID   string                 `json:"i"`
	State      map[string]interface{} `json:"s"`
	Commands   []string               `json:"o"`
	WorkerID   string                 `json:"w"`
}

// DeviceCommandMessage used by server to invoke device command on a worker.
type DeviceCommandMessage struct {
	MessageWithType
	DeviceID string
	Command  enums.Command
	Payload  map[string]interface{}
}

// NewDiscoveryMessage constructs discovery message.
func NewDiscoveryMessage(nodeID string, firstStart bool, properties map[string]string,
	maxDevices int) *DiscoveryMessage {
	msg := DiscoveryMessage{
		MessageWithType: MessageWithType{
			Type:     bus.MsgPing,
			SendTime: utils.TimeNow(),
		},
		NodeID:       nodeID,
		IsFirstStart: firstStart,
		Properties:   make(map[string]string, len(properties)),
		MaxDevices:   maxDevices,
	}

	for k, v := range properties {
		msg.Properties[k] = v
	}

	return &msg
}

// NewDeviceAssignmentMessage constructs device assignment message.
func NewDeviceAssignmentMessage(devices []*DeviceAssignment) *DeviceAssignmentMessage {
	return &DeviceAssignmentMessage{
		MessageWithType: MessageWithType{
			Type:     bus.MsgDeviceAssignment,
			SendTime: utils.TimeNow(),
		},
		Devices: devices,
	}
}

// NewDeviceUpdateMessage constructs device update message.
func NewDeviceUpdateMessage() *DeviceUpdateMessage {
	return &DeviceUpdateMessage{
		MessageWithType: MessageWithType{
			Type:     bus.MsgDeviceUpdate,
			SendTime: utils.TimeNow(),
		},
	}
}

// NewDeviceCommandMessage constructs device command message.
func NewDeviceCommandMessage(deviceID string, command enums.Command,
	data map[string]interface{}) *DeviceCommandMessage {
	return &DeviceCommandMessage{
		MessageWithType: MessageWithType{
			Type:     bus.MsgDeviceCommand,
			SendTime: utils.TimeNow(),
		},
		Command:  command,
		Payload:  data,
		DeviceID: deviceID,
	}
}
