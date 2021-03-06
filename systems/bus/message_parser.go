package bus

import (
	"encoding/json"

	"go-home.io/x/server/plugins/bus"
	"go-home.io/x/server/plugins/common"
)

// IMessageParserProvider describes messages parser.
type IMessageParserProvider interface {
	ProcessIncomingMessage(r *bus.RawMessage)
}

// IMasterMessageParserProvider describes messages parser for server.
type IMasterMessageParserProvider interface {
	IMessageParserProvider

	GetDiscoveryMessageChan() chan *DiscoveryMessage
	GetDeviceUpdateMessageChan() chan *DeviceUpdateMessage
	GetEntityLoadStatueMessageChan() chan *EntityLoadStatusMessage
}

// IWorkerMessageParserProvider describes messages parser for worker.
type IWorkerMessageParserProvider interface {
	IMessageParserProvider

	GetDeviceAssignmentMessageChan() chan *DeviceAssignmentMessage
	GetDeviceCommandMessageChan() chan *DeviceCommandMessage
}

// Message parser implementation.
type messageParser struct {
	logger   common.ILoggerProvider
	isWorker bool

	deviceAssignmentChan chan *DeviceAssignmentMessage
	deviceCommandsChan   chan *DeviceCommandMessage

	discoveryMessageChan        chan *DiscoveryMessage
	deviceUpdateMessageChan     chan *DeviceUpdateMessage
	entityLoadStatusMessageChan chan *EntityLoadStatusMessage
}

// NewWorkerMessageParser constructs parser for worker.
func NewWorkerMessageParser(logger common.ILoggerProvider) IWorkerMessageParserProvider {
	return &messageParser{
		logger:               logger,
		deviceAssignmentChan: make(chan *DeviceAssignmentMessage, 5),
		deviceCommandsChan:   make(chan *DeviceCommandMessage, 20),
		isWorker:             true,
	}
}

// NewMasterMessageParser constructs parser for server.
func NewMasterMessageParser(logger common.ILoggerProvider) IMasterMessageParserProvider {
	return &messageParser{
		logger:                      logger,
		discoveryMessageChan:        make(chan *DiscoveryMessage, 5),
		deviceUpdateMessageChan:     make(chan *DeviceUpdateMessage, 50),
		entityLoadStatusMessageChan: make(chan *EntityLoadStatusMessage, 50),
		isWorker:                    false,
	}
}

// GetDeviceAssignmentMessageChan returns channel used for device assignment callbacks.
func (w *messageParser) GetDeviceAssignmentMessageChan() chan *DeviceAssignmentMessage {
	return w.deviceAssignmentChan
}

// GetDeviceCommandMessageChan returns channel used for device command callbacks.
func (w *messageParser) GetDeviceCommandMessageChan() chan *DeviceCommandMessage {
	return w.deviceCommandsChan
}

// GetDiscoveryMessageChan returns channel used for discovery callbacks.
func (w *messageParser) GetDiscoveryMessageChan() chan *DiscoveryMessage {
	return w.discoveryMessageChan
}

// GetDeviceUpdateMessageChan returns channel used for device updates callbacks.
func (w *messageParser) GetDeviceUpdateMessageChan() chan *DeviceUpdateMessage {
	return w.deviceUpdateMessageChan
}

func (w *messageParser) GetEntityLoadStatueMessageChan() chan *EntityLoadStatusMessage {
	return w.entityLoadStatusMessageChan
}

// ProcessIncomingMessage parses incoming service bus message.
func (w *messageParser) ProcessIncomingMessage(r *bus.RawMessage) {
	var err error
	b, err := parseRawMessage(r)
	if err != nil {
		w.logger.Error("Failed to parse incoming message", err, common.LogSystemToken, logSystem)
		return
	}

	if w.isWorker {
		err = w.processWorkerMessage(b, r)
	} else {
		err = w.processServerMessage(b, r)
	}

	if err != nil {
		w.logger.Error("Failed to parse incoming message", err, "type", b.Type.String(),
			common.LogSystemToken, logSystem)
	}
}

// Processes worker messages.
// nolint: dupl
func (w *messageParser) processWorkerMessage(b *MessageWithType, r *bus.RawMessage) error {
	var err error
	switch b.Type {
	case bus.MsgDeviceAssignment:
		var d DeviceAssignmentMessage
		err = json.Unmarshal(r.Body, &d)
		if err == nil {
			w.deviceAssignmentChan <- &d
		}
	case bus.MsgDeviceCommand:
		var d DeviceCommandMessage
		err = json.Unmarshal(r.Body, &d)
		if err == nil {
			w.deviceCommandsChan <- &d
		}
	default:
		w.logger.Warn("Received unknown message type", "type", b.Type.String(),
			common.LogSystemToken, logSystem)
		err = &ErrUnknownType{}
	}

	return err
}

// Processes server messages.
// nolint: dupl
func (w *messageParser) processServerMessage(b *MessageWithType, r *bus.RawMessage) error {
	var err error

	switch b.Type {
	case bus.MsgPing:
		var m DiscoveryMessage
		err := json.Unmarshal(r.Body, &m)
		if err == nil {
			w.discoveryMessageChan <- &m
		}
	case bus.MsgDeviceUpdate:
		var m DeviceUpdateMessage
		err := json.Unmarshal(r.Body, &m)
		if err == nil {
			w.deviceUpdateMessageChan <- &m
		}
	case bus.MsgEntityLoadStatus:
		var m EntityLoadStatusMessage
		err := json.Unmarshal(r.Body, &m)
		if err == nil {
			w.entityLoadStatusMessageChan <- &m
		}
	default:
		w.logger.Warn("Received unknown message type", "type", b.Type.String(),
			common.LogSystemToken, logSystem)
		err = &ErrUnknownType{}
	}

	return err
}
