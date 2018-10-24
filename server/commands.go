package server

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gobwas/glob"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems/bus"
)

const (
	// Default location name.
	defaultLocationName = "Default"
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
				common.LogIDToken, v.ID, common.LogDeviceCommandToken, cmd.String())
			continue
		}

		s.Logger.Debug("Invoking device operation", common.LogSystemToken, logSystem,
			common.LogIDToken, v.ID, common.LogDeviceCommandToken, cmd.String())

		if v.Type == enums.DevGroup {
			g, ok := s.groups[v.ID]
			if !ok {
				s.Logger.Warn("Received unknown group", common.LogSystemToken, logSystem,
					common.LogIDToken, v.ID, common.LogDeviceCommandToken, cmd.String())
				continue
			}
			g.InvokeCommand(cmd, data)
		} else {
			s.Settings.ServiceBus().PublishToWorker(v.Worker,
				bus.NewDeviceCommandMessage(v.ID, cmd, data))
		}
	}
}

// Invokes device command if it's allowed for the user.
func (s *GoHomeServer) commandInvokeDeviceCommand(user providers.IAuthenticatedUser,
	deviceID string, cmdName string, data []byte) error {
	knownDevice := s.state.GetDevice(deviceID)
	if nil == knownDevice {
		s.Logger.Warn("Failed to find device", common.LogSystemToken, logSystem,
			common.LogIDToken, deviceID, common.LogUserNameToken, user.Name())
		return &ErrUnknownDevice{ID: deviceID}
	}

	// We don't want to allow to brute-forth device names, so returning generic error
	if !user.DeviceCommand(knownDevice.ID) {
		s.Logger.Warn("User doesn't have access to this device", common.LogSystemToken, logSystem,
			common.LogIDToken, deviceID, common.LogUserNameToken, user.Name())
		return &ErrUnknownDevice{ID: deviceID}
	}

	command, err := enums.CommandString(cmdName)
	if err != nil {
		s.Logger.Warn("Received unknown command", common.LogSystemToken, logSystem,
			common.LogIDToken, deviceID, common.LogDeviceCommandToken, cmdName,
			common.LogUserNameToken, user.Name())
		return &ErrUnknownCommand{Name: cmdName}
	}

	if !helpers.SliceContainsString(knownDevice.Commands, cmdName) {
		s.Logger.Warn("Received command is not supported", common.LogSystemToken, logSystem,
			common.LogIDToken, deviceID, common.LogDeviceCommandToken, cmdName,
			common.LogUserNameToken, user.Name())
		return &ErrUnsupportedCommand{Name: cmdName}
	}

	inputData := make(map[string]interface{})
	if len(data) > 0 {
		err := json.Unmarshal(data, &inputData)
		if err != nil {
			data = []byte(fmt.Sprintf(`{ "value" : %s  }`, string(data)))
			err := json.Unmarshal(data, &inputData)
			if err != nil {
				s.Logger.Error("Failed to unmarshal input request", err,
					common.LogSystemToken, logSystem)
				return &ErrBadRequest{}
			}
		}
	}

	if knownDevice.Type == enums.DevGroup {
		return s.commandGroupCommand(user, knownDevice.ID, command, inputData)
	}

	s.Logger.Debug("Invoking device operation", common.LogSystemToken, logSystem,
		common.LogIDToken, deviceID, common.LogDeviceCommandToken, cmdName,
		common.LogUserNameToken, user.Name())
	s.Settings.ServiceBus().PublishToWorker(knownDevice.Worker,
		bus.NewDeviceCommandMessage(deviceID, command, inputData))
	return nil
}

// Invokes group command
func (s *GoHomeServer) commandGroupCommand(user providers.IAuthenticatedUser,
	groupID string, cmd enums.Command, data map[string]interface{}) error {
	g, ok := s.groups[groupID]
	if !ok {
		s.Logger.Warn("Received unknown group", common.LogSystemToken, logSystem,
			common.LogIDToken, groupID, common.LogDeviceCommandToken, cmd.String(),
			common.LogUserNameToken, user.Name())
		return &ErrUnknownGroup{Name: groupID}
	}

	g.InvokeCommand(cmd, data)
	return nil
}

// Returns all allowed for the user devices.
func (s *GoHomeServer) commandGetAllDevices(user providers.IAuthenticatedUser) []*knownDevice {
	allowedDevices := make([]*knownDevice, 0)
	isWorkerAllowed := user.Workers()

	for _, v := range s.state.GetAllDevices() {
		worker := v.Worker
		if !isWorkerAllowed {
			worker = ""
		}

		if user.DeviceGet(v.ID) {
			d := &knownDevice{
				ID:         v.ID,
				Type:       v.Type,
				State:      v.State,
				Name:       v.Name,
				Worker:     worker,
				Commands:   v.Commands,
				LastSeen:   v.LastSeen,
				IsReadOnly: !user.DeviceCommand(v.ID),
			}
			allowedDevices = append(allowedDevices, d)
		}
	}

	return allowedDevices
}

// Returns all allowed for the user groups.
func (s *GoHomeServer) commandGetAllGroups(user providers.IAuthenticatedUser) []*knownGroup {
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
			ID:      v.ID,
			Name:    v.Name,
			Devices: make([]string, 0),
		}

		for _, dev := range g.Devices() {
			d := s.state.GetDevice(dev)
			if nil == d || !user.DeviceGet(v.ID) {
				continue
			}

			group.Devices = append(group.Devices, dev)
		}

		if 0 != len(group.Devices) {
			sort.Strings(group.Devices)
			response = append(response, group)
		}
	}

	sort.Slice(response, func(i, j int) bool {
		return response[i].Name < response[j].Name
	})
	return response
}

// Returns all allowed for the user locations.
// nolint: gocyclo
func (s *GoHomeServer) commandGetAllLocations(user providers.IAuthenticatedUser) []*knownLocation {
	devices := s.commandGetAllDevices(user)
	groups := s.commandGetAllGroups(user)
	response := make([]*knownLocation, 0)
	devicesProcessed := make([]string, 0)
	var defaultLocation *knownLocation

	for _, v := range s.locations {
		location := &knownLocation{
			Name:    v.ID(),
			Icon:    v.Icon(),
			Devices: make([]string, 0),
		}

		for _, dev := range v.Devices() {
			d := s.state.GetDevice(dev)
			if nil == d || !user.DeviceGet(d.ID) {
				continue
			}

			location.Devices = append(location.Devices, dev)
			devicesProcessed = append(devicesProcessed, dev)
		}

		if 0 != len(location.Devices) {
			sort.Strings(location.Devices)
			response = append(response, location)
		}

		if defaultLocationName == location.Name {
			defaultLocation = location
		}
	}

	devicesLeft := make([]string, 0)
	for _, v := range devices {
		if helpers.SliceContainsString(devicesProcessed, v.ID) || groupsHasDevice(groups, v.ID) {
			continue
		}

		devicesLeft = append(devicesLeft, v.ID)
	}

	if 0 == len(devicesLeft) {
		return response
	}

	if nil == defaultLocation {
		defaultLocation = &knownLocation{
			Name:    defaultLocationName,
			Devices: make([]string, 0),
		}

		response = append(response, defaultLocation)
	}

	defaultLocation.Devices = append(defaultLocation.Devices, devicesLeft...)
	sort.Strings(defaultLocation.Devices)

	sort.Slice(response, func(i, j int) bool {
		if defaultLocationName == response[i].Name {
			return true
		}

		if defaultLocationName == response[j].Name {
			return false
		}

		return response[i].Name < response[j].Name
	})
	return response
}

// Checks whether this device has been claimed as a part of a group.
func groupsHasDevice(groups []*knownGroup, deviceID string) bool {
	for _, g := range groups {
		if helpers.SliceContainsString(g.Devices, deviceID) {
			return true
		}
	}

	return false
}
