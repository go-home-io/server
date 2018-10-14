package server

import (
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests correct users' selection.
func TestDeviceOperation(t *testing.T) {
	worker1 := false
	worker2 := false
	s := getFakeSettings(func(name string, msg ...interface{}) {
		switch name {
		case "1":
			worker1 = true
		case "2":
			worker2 = true
		}

	}, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1":   {ID: "dev1", Commands: []string{enums.CmdOn.String(), enums.CmdSetBrightness.String()}, Worker: "1"},
		"device": {ID: "device", Commands: []string{enums.CmdOn.String()}, Worker: "2"},
		"g1":     {ID: "g1", Type: enums.DevGroup, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
	}

	user := &providers.AuthenticatedUser{
		Username: "usr1",
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("dev?"),
						compileRegexp("g?"),
					},
				},
			},
		},
	}

	groupCalled := false

	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
		groups: map[string]providers.IGroupProvider{
			"g1": mocks.FakeNewGroupProvider("g1", []string{"dev1"}, func() {
				groupCalled = true
			}),
		},
	}

	data := []struct {
		deviceID string
		cmd      enums.Command
		data     string
		result   *bool
		gold     bool
		isGroup  bool
		isError  bool
	}{
		{
			cmd:      enums.CmdOn,
			deviceID: "dev1",
			data:     "",
			result:   &worker1,
			gold:     true,
			isGroup:  false,
			isError:  false,
		},
		{
			deviceID: "dev1",
			cmd:      enums.CmdSetBrightness,
			data:     "20",
			result:   &worker1,
			gold:     true,
			isGroup:  false,
			isError:  false,
		},
		{
			deviceID: "device",
			cmd:      enums.CmdOn,
			data:     "",
			result:   &worker2,
			gold:     false,
			isGroup:  false,
			isError:  true,
		},
		{
			deviceID: "g1",
			cmd:      enums.CmdOn,
			data:     "",
			result:   &groupCalled,
			gold:     true,
			isGroup:  false,
		},
		{
			deviceID: "g2",
			cmd:      enums.CmdOn,
			data:     "",
			result:   &groupCalled,
			gold:     false,
			isGroup:  true,
			isError:  true,
		},
	}

	for _, v := range data {
		worker1 = false
		worker2 = false
		groupCalled = false

		var err error
		if v.isGroup {
			err = srv.commandGroupCommand(user, v.deviceID, v.cmd, nil)
		} else {
			err = srv.commandInvokeDeviceCommand(user, v.deviceID, v.cmd.String(), []byte(v.data))
		}

		if v.isError {
			assert.Error(t, err, "%s invoke no error %s", v.deviceID, v.cmd.String())
		} else {
			require.NoError(t, err, "%s invoke error %s", v.deviceID, v.cmd.String())
		}

		assert.Equal(t, v.gold, *v.result, "%s result %s", v.deviceID, v.cmd.String())
	}
}

// Tests unknown device error.
func TestUnknownDevice(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{}
	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
	}

	user := &providers.AuthenticatedUser{
		Username: "usr1",
		Rules:    map[providers.SecSystem][]*providers.BakedRule{},
	}

	err := srv.commandInvokeDeviceCommand(user, "dev1", "on", []byte(""))
	require.Error(t, err)
	assert.IsType(t, &ErrUnknownDevice{}, err)
}

// Tests forbidden device error.
func TestDeviceForbidden(t *testing.T) {
	logFound := false
	s := getFakeSettings(nil, nil, func(s string) {
		if s == "User doesn't have access to this device" {
			logFound = true
		}
	})
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1": {ID: "dev1", Commands: []string{enums.CmdOn.String()}, Worker: "1"},
	}
	srv := &GoHomeServer{
		state:    state,
		Logger:   state.Logger,
		Settings: s,
	}

	user := &providers.AuthenticatedUser{
		Username: "usr1",
		Rules:    map[providers.SecSystem][]*providers.BakedRule{},
	}

	err := srv.commandInvokeDeviceCommand(user, "dev1", "on", []byte(""))
	require.Error(t, err)
	assert.IsType(t, &ErrUnknownDevice{}, err)
	assert.True(t, logFound, "log not found")
}

// Tests unknown and non-supported commands error.
func TestWrongCommand(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1": {ID: "dev1", Commands: []string{enums.CmdOn.String()}, Worker: "1"},
	}
	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
	}

	user := &providers.AuthenticatedUser{
		Username: "usr1",
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("dev?"),
					},
				},
			},
		},
	}

	data := []struct {
		cmd  string
		data string
		err  interface{}
	}{
		{
			cmd:  "on1",
			data: "",
			err:  &ErrUnknownCommand{},
		},
		{
			cmd:  "off",
			data: "",
			err:  &ErrUnsupportedCommand{},
		},
		{
			cmd:  "on",
			data: "wrong data",
			err:  &ErrBadRequest{},
		},
	}

	for _, v := range data {
		err := srv.commandInvokeDeviceCommand(user, "dev1", v.cmd, []byte(v.data))
		require.Error(t, err, "%s: %s", v.cmd, v.data)
		assert.IsType(t, v.err, err, "%s: %s", v.cmd, v.data)
	}
}

// Tests correct filtration of devices.
func TestGetAllDevices(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1":   {ID: "dev1", Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"device": {ID: "device", Commands: []string{enums.CmdOn.String()}, Worker: "2"},
	}

	user := &providers.AuthenticatedUser{
		Username: "usr1",
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("dev?"),
					},
				},
			},
		},
	}

	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
	}

	devices := srv.commandGetAllDevices(user)
	require.Equal(t, 1, len(devices), "len")
	assert.Equal(t, "dev1", devices[0].ID, "name")
}

// Tests internal command invoke.
func TestInternalInvokeCommand(t *testing.T) {
	numCalled := 0
	s := getFakeSettings(func(name string, msg ...interface{}) {
		numCalled++
	}, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1":   {ID: "dev1", Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"dev2":   {ID: "dev2", Commands: []string{enums.CmdOff.String()}, Worker: "1"},
		"device": {ID: "device", Commands: []string{enums.CmdOn.String()}, Worker: "2"},
	}

	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
	}

	srv.InternalCommandInvokeDeviceCommand(glob.MustCompile("dev[1-2]*"), enums.CmdOn, nil)
	assert.Equal(t, 1, numCalled)
}

// Tests group invoke.
func TestInternalInvokeCommandGroup(t *testing.T) {
	numCalled := 0
	s := getFakeSettings(func(name string, msg ...interface{}) {
		numCalled++
	}, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1":   {ID: "dev1", Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"dev2":   {ID: "dev2", Type: enums.DevGroup, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"device": {ID: "device", Commands: []string{enums.CmdOn.String()}, Worker: "2"},
	}

	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
	}

	srv.InternalCommandInvokeDeviceCommand(glob.MustCompile("dev[1-2]*"), enums.CmdOn, nil)
	require.Equal(t, 1, numCalled, "device call w/o group")

	groupCalled := 0
	fg := mocks.FakeNewGroupProvider("dev2", nil, func() {
		groupCalled++
	})
	numCalled = 0
	srv.groups = map[string]providers.IGroupProvider{"dev2": fg}

	srv.InternalCommandInvokeDeviceCommand(glob.MustCompile("dev[1-2]*"), enums.CmdOn, nil)
	assert.Equal(t, 1, numCalled, "device call")
	assert.Equal(t, 1, groupCalled, "group call")
}

// Tests groups invocation.
func TestGetGroups(t *testing.T) {
	user := &providers.AuthenticatedUser{
		Username: "usr1",
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("dev?"),
						compileRegexp("g?"),
					},
				},
			},
		},
	}

	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1":   {ID: "dev1", Type: enums.DevSwitch, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"dev2":   {ID: "dev2", Type: enums.DevSwitch, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"dev3":   {ID: "dev3", Type: enums.DevSwitch, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"g1":     {ID: "g1", Type: enums.DevGroup, Commands: []string{enums.CmdOff.String()}, Worker: "1"},
		"g2":     {ID: "g2", Type: enums.DevGroup, Commands: []string{enums.CmdOn.String()}, Worker: "2"},
		"group3": {ID: "group3", Type: enums.DevGroup, Commands: []string{enums.CmdOn.String()}, Worker: "2"},
	}

	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
		groups: map[string]providers.IGroupProvider{
			"g1":     mocks.FakeNewGroupProvider("g1", []string{"dev1"}, nil),
			"g2":     mocks.FakeNewGroupProvider("g2", []string{"dev2"}, nil),
			"group3": mocks.FakeNewGroupProvider("group3", []string{"dev3"}, nil),
		},
	}

	groups := srv.commandGetAllGroups(user)
	assert.Equal(t, 2, len(groups), "groups")

	for _, v := range groups {
		assert.Equal(t, 1, len(v.Devices), "devices for %s", v.Name)
	}
}

// Tests locations invocation.
func TestGetLocations(t *testing.T) {
	user := &providers.AuthenticatedUser{
		Username: "usr1",
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("dev?"),
						compileRegexp("g?"),
					},
				},
			},
		},
	}

	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1":   {ID: "dev1", Type: enums.DevSwitch, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"dev2":   {ID: "dev2", Type: enums.DevSwitch, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"dev3":   {ID: "dev3", Type: enums.DevSwitch, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
		"g1":     {ID: "g1", Type: enums.DevGroup, Commands: []string{enums.CmdOff.String()}, Worker: "1"},
		"g2":     {ID: "g2", Type: enums.DevGroup, Commands: []string{enums.CmdOn.String()}, Worker: "2"},
		"group3": {ID: "group3", Type: enums.DevGroup, Commands: []string{enums.CmdOn.String()}, Worker: "2"},
	}

	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
		groups: map[string]providers.IGroupProvider{
			"g1":     mocks.FakeNewGroupProvider("g1", []string{"dev1"}, nil),
			"g2":     mocks.FakeNewGroupProvider("g2", []string{"dev2"}, nil),
			"group3": mocks.FakeNewGroupProvider("group3", []string{"dev3"}, nil),
		},
		locations: []providers.ILocationProvider{
			mocks.FakeNewLocationProvider("l1", []string{"g1"}, nil),
			mocks.FakeNewLocationProvider("l2", []string{"group3"}, nil),
		},
	}

	locations := srv.commandGetAllLocations(user)
	assert.Equal(t, 2, len(locations), "number of groups")

	found := false
	for _, v := range locations {
		if v.Name == "Default" {
			assert.Equal(t, 2, len(v.Devices), "number of devices")
			found = true
			break
		}
	}

	assert.True(t, found, "group")
}
