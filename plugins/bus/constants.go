//go:generate enumer -type=ChannelName -transform=kebab -trimprefix=Ch -text -json -yaml
//go:generate enumer -type=MessageType -transform=snake -trimprefix=Msg -text -json -yaml

// Package bus contains service bus plugin definitions.
package bus

// ChannelName describes enum with known service bus channels.
type ChannelName int

const (
	// ChDiscovery describes discovery messages channel.
	ChDiscovery ChannelName = iota
	// ChDeviceUpdates describes devices updates channel.
	ChDeviceUpdates
)

// MessageType describes enum with known service bus messages.
type MessageType int

const (
	// MsgPing describes ping/discovery message sent by worker.
	MsgPing MessageType = iota
	// MsgDeviceAssignment describes new devices assignment sent by master.
	MsgDeviceAssignment
	// MsgDeviceUpdate describes devices' state update sent by worker.
	MsgDeviceUpdate
	// MsgDeviceCommand describes device command sent by master.
	MsgDeviceCommand
)

const (
	// MsgTTLSeconds describes maximum valid message age.
	// Message will be discarded if it was send earlier than now()-MsgTTLSeconds.
	MsgTTLSeconds = 10
	// ChWorkerFormat describes formatter for worker messages channel.
	ChWorkerFormat = "worker%s"
)
