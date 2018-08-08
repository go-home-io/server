package server

import (
	"encoding/json"
	"errors"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/gobwas/glob"
)

// InternalCommandInvokeDeviceCommand invokes devices operations.
// This command is used strictly internally.
func (s *GoHomeServer) InternalCommandInvokeDeviceCommand(
	deviceRegexp glob.Glob, cmd enums.Command, data map[string]interface{}) {
	if nil == data {
		data = make(map[string]interface{})
	}

	for _, v := range s.state.GetAllDevices() {
		if !deviceRegexp.Match(v.ID) {
			continue
		}

		if !helpers.SliceContainsString(v.Commands, cmd.String()) {
			s.Logger.Warn("Received command is not supported", common.LogSystemToken, logSystem,
				common.LogDeviceNameToken, v.ID, common.LogDeviceCommandToken, cmd.String())
			continue
		}

		s.Logger.Debug("Invoking device operation", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, v.ID, common.LogDeviceCommandToken, cmd.String())
		s.Settings.ServiceBus().PublishToWorker(v.Worker,
			bus.NewDeviceCommandMessage(v.ID, cmd, data))
	}

}

// Invokes device command if it's allowed for the user.
func (s *GoHomeServer) commandInvokeDeviceCommand(user *providers.AuthenticatedUser,
	deviceID string, cmdName string, data []byte) error {
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

	command, err := enums.CommandString(cmdName)
	if err != nil {
		s.Logger.Warn("Received unknown command", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID, common.LogDeviceCommandToken, cmdName,
			common.LogUserNameToken, user.Username)
		return errors.New("unknown command")
	}

	if !helpers.SliceContainsString(knownDevice.Commands, cmdName) {
		s.Logger.Warn("Received command is not supported", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, deviceID, common.LogDeviceCommandToken, cmdName,
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

	if knownDevice.Type == enums.DevGroup {
		return s.commandGroupCommand(user, knownDevice.ID, command, inputData)
	}

	s.Logger.Debug("Invoking device operation", common.LogSystemToken, logSystem,
		common.LogDeviceNameToken, deviceID, common.LogDeviceCommandToken, cmdName,
		common.LogUserNameToken, user.Username)
	s.Settings.ServiceBus().PublishToWorker(knownDevice.Worker,
		bus.NewDeviceCommandMessage(deviceID, command, inputData))
	return nil
}

// Invokes group command
func (s *GoHomeServer) commandGroupCommand(user *providers.AuthenticatedUser,
	groupID string, cmd enums.Command, data map[string]interface{}) error {
	g, ok := s.groups[groupID]
	if !ok {
		s.Logger.Warn("Received unknown group", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, groupID, common.LogDeviceCommandToken, cmd.String(),
			common.LogUserNameToken, user.Username)
		return errors.New("unknown group")
	}

	g.InvokeCommand(cmd, data)
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

// Returns all allowed for the user locations.
func (s *GoHomeServer) commandGetAllGroups(user *providers.AuthenticatedUser) []*knownGroup {
	devices := s.commandGetAllDevices(user)
	response := make([]*knownGroup, 0)
	for _, v := range devices {
		if v.Type != enums.DevGroup {
			continue
		}

		g, ok := s.groups[v.ID]
		if !ok {
			continue
		}

		group := &knownGroup{
			ID: v.ID,
			knownLocation: knownLocation{
				Name:    v.Name,
				Devices: g.Devices(),
			},
		}

		response = append(response, group)
	}

	return response
}
