package server

import (
	"testing"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/mocks"
	"github.com/gobwas/glob"
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

	err := srv.commandInvokeDeviceCommand(user, "dev1", "on", []byte(""))

	if err != nil || !worker1 {
		t.Fail()
	}

	err = srv.commandInvokeDeviceCommand(user, "device", "on", []byte(""))
	if err == nil || worker2 {
		t.Fail()
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

	if err == nil || err.Error() != "unknown device" {
		t.Fail()
	}
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

	if err == nil || err.Error() != "unknown device" || !logFound {
		t.Fail()
	}
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

	err := srv.commandInvokeDeviceCommand(user, "dev1", "on1", []byte(""))
	if err == nil || err.Error() != "unknown command" {
		t.Fail()
	}

	err = srv.commandInvokeDeviceCommand(user, "dev1", "off", []byte(""))
	if err == nil || err.Error() != "command is not supported" {
		t.Fail()
	}

	err = srv.commandInvokeDeviceCommand(user, "dev1", "on", []byte("wrong data"))
	if err == nil || err.Error() != "bad request data" {
		t.Fail()
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
	if 1 != len(devices) || "dev1" != devices[0].ID {
		t.Fail()
	}
}

func TestInternalInvokeCommand(t *testing.T) {
	numCalled := 0
	s := getFakeSettings(func(name string, msg ...interface{}) {
		numCalled ++
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
	if 1 != numCalled{
		t.Fail()
	}
}
