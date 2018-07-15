package server

import (
	"encoding/json"
	"errors"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/utils"
)

// Invokes device command if it's allowed for the user.
func (s *GoHomeServer) commandInvokeDeviceCommand(user *providers.AuthenticatedUser,
	deviceID string, opName string, data []byte) error {
	knownDevice := s.state.GetDevice(deviceID)
	if nil == knownDevice {
		s.Logger.Warn("Failed to find device", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID, common.LogUserNameToken, user.Username)
		return errors.New("unknown device")
	}

	// We don't want to allow to brute-forth device names, so returning generic error
	if !knownDevice.Command(user) {
		s.Logger.Warn("User doesn't have access to this device", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID, common.LogUserNameToken, user.Username)
		return errors.New("unknown device")
	}

	command, err := enums.CommandString(opName)
	if err != nil {
		s.Logger.Warn("Received unknown command", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID, common.LogDeviceCommandToken, opName,
			common.LogUserNameToken, user.Username)
		return errors.New("unknown command")
	}

	if !utils.SliceContainsString(knownDevice.Commands, opName) {
		s.Logger.Warn("Received command is not supported", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID, common.LogDeviceCommandToken, opName,
			common.LogUserNameToken, user.Username)
		return errors.New("command is not supported")
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

	s.Logger.Debug("Invoking device operation", common.LogSystemToken, logSystem,
		common.LogDeviceNameToken, deviceID, common.LogDeviceCommandToken, opName,
		common.LogUserNameToken, user.Username)
	s.Settings.ServiceBus().PublishToWorker(knownDevice.Worker,
		bus.NewDeviceCommandMessage(deviceID, command, inputData))
	return nil
}

// Returns all allowed for the user devices.
func (s *GoHomeServer) commandGetAllDevices(user *providers.AuthenticatedUser) []*knownDevice {
	allowedDevices := make([]*knownDevice, 0)

	for _, v := range s.state.GetAllDevices() {
		if v.Get(user) {
			allowedDevices = append(allowedDevices, v)
		}
	}

	return allowedDevices
}
