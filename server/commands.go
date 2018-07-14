package server

import (
	"encoding/json"
	"errors"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/systems/bus"
)

// Invokes device command if it's allowed for the user.
func (s *GoHomeServer) commandInvokeDeviceCommand(deviceID string, opName string, data []byte) error {
	knownDevice := s.state.GetDevice(deviceID)
	if nil == knownDevice {
		s.Logger.Warn("Failed to find device", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID)
		return errors.New("unknown device")
	}

	command, err := enums.CommandString(opName)
	if err != nil {
		s.Logger.Warn("Received unknown command", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID, common.LogDeviceCommandToken, opName)
		return errors.New("unknown command")
	}

	inputData := make(map[string]interface{})
	if len(data) > 0 {
		err := json.Unmarshal(data, &inputData)
		if err != nil {
			s.Logger.Error("Failed to unmarshal input request", err,
				common.LogSystemToken, logSystem)
			return errors.New("bad request data")
		}
	}

	s.Settings.ServiceBus().PublishToWorker(knownDevice.Worker, bus.NewDeviceCommandMessage(deviceID, command, inputData))
	return nil
}
