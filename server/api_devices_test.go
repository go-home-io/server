package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/gobwas/glob"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems/security"
	"go-home.io/x/server/utils"
)

type fakeTriggerWrapper struct {
	id string
}

func (f *fakeTriggerWrapper) GetID() string {
	return f.id
}

func (f *fakeTriggerWrapper) GetLastTriggeredTime() int64 {
	return utils.TimeNow()
}

func getFakeRootUser(_ *http.Request) providers.IAuthenticatedUser {
	return &security.AuthenticatedUser{
		Username: "test",
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					Command: true,
					History: true,
					Resources: []glob.Glob{
						compileRegexp("*"),
					},
				},
			},
			providers.SecSystemCore: {
				{
					Get: true,
					Resources: []glob.Glob{
						compileRegexp("*"),
					},
				},
			},
			providers.SecSystemTrigger: {
				{
					Get:     true,
					History: true,
					Resources: []glob.Glob{
						compileRegexp("trigger1t*.trigger")},
				},
			},
		},
	}
}

func getServer() *GoHomeServer {
	s := getFakeSettings(func(_ string, _ ...interface{}) {}, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1":   {ID: "dev1", Commands: []string{enums.CmdOn.String(), enums.CmdSetBrightness.String()}, Worker: "1"},
		"device": {ID: "device", Commands: []string{enums.CmdOn.String()}, Worker: "2"},
		"g1":     {ID: "g1", Type: enums.DevGroup, Commands: []string{enums.CmdOn.String()}, Worker: "1"},
	}

	return &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: s,
		groups: map[string]providers.IGroupProvider{
			"g1": mocks.FakeNewGroupProvider("g1", []string{"dev1"}, func() {}),
		},
		triggers: []*knownMasterComponent{
			{
				Loaded: true,
				Name:   "trigger1",
				Interface: &fakeTriggerWrapper{
					id: "trigger1test.trigger",
				},
			},
			{
				Loaded: true,
				Name:   "0trigger",
				Interface: &fakeTriggerWrapper{
					id: "trigger123.trigger",
				},
			},
		},
	}
}

// Tests get devices.
func TestGetDevicesAPI(t *testing.T) {
	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getDevices)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "response code")
	data := make([]*knownDevice, 0)
	err = json.Unmarshal(r.Body.Bytes(), &data)
	assert.NoError(t, err, "wrong response")
	assert.Equal(t, 3, len(data), "incorrect devices num")
}

// Tests get groups.
func TestGetGroupsAPI(t *testing.T) {
	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	http.HandlerFunc(srv.getGroups).ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "response code")
	data := make([]*knownGroup, 0)
	err = json.Unmarshal(r.Body.Bytes(), &data)
	assert.NoError(t, err, "wrong response")
	assert.Equal(t, 1, len(data), "incorrect group num")
}

// Tests get state.
func TestGetCurrentStateAPI(t *testing.T) {
	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	http.HandlerFunc(srv.getCurrentState).ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "response code")
	data := &currentState{}
	err = json.Unmarshal(r.Body.Bytes(), data)
	assert.NoError(t, err, "wrong response")
	assert.Equal(t, 3, len(data.Devices), "incorrect devices num")
	assert.Equal(t, 1, len(data.Groups), "incorrect groups num")
	assert.Equal(t, 1, len(data.Triggers), "incorrect triggers num")
	assert.Equal(t, "trigger1test.trigger", data.Triggers[0].ID, "incorrect trigger")
}

// Tests device command.
func TestDeviceCommandAPI(t *testing.T) {
	input := map[string]int{
		"dev1": http.StatusOK,
		"dev2": http.StatusInternalServerError,
		"g1":   http.StatusOK,
		"g2":   http.StatusInternalServerError,
	}

	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	for k, v := range input {
		req, err := http.NewRequest("POST", "/test", strings.NewReader(""))
		require.NoError(t, err, "setup failed %s", k)
		req = mux.SetURLVars(req, map[string]string{
			string(urlDeviceID):    k,
			string(urlCommandName): enums.CmdOn.String(),
		})

		r := httptest.NewRecorder()
		http.HandlerFunc(srv.deviceCommand).ServeHTTP(r, req)
		assert.Equal(t, v, r.Code, "response code %s", k)
	}
}

// Test getting the history.
func TestGetDeviceStateHistoryAPI(t *testing.T) {
	input := map[string]int{
		"dev1": http.StatusOK,
		"dev2": http.StatusInternalServerError,
		"g1":   http.StatusOK,
		"g2":   http.StatusInternalServerError,
	}

	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	for k, v := range input {
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err, "setup failed %s", k)
		req = mux.SetURLVars(req, map[string]string{string(urlDeviceID): k})

		r := httptest.NewRecorder()
		http.HandlerFunc(srv.getDeviceStateHistory).ServeHTTP(r, req)
		assert.Equal(t, v, r.Code, "response code %s", k)
	}
}

// Test getting trigger history.
func TestGetTriggerStateHistoryAPI(t *testing.T) {
	input := map[string]int{
		"trigger1test.trigger": http.StatusOK,
		"trigger123.trigger": http.StatusForbidden,
		"dev2": http.StatusInternalServerError,
	}

	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	for k, v := range input {
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err, "setup failed %s", k)
		req = mux.SetURLVars(req, map[string]string{string(urlTriggerID): k})

		r := httptest.NewRecorder()
		http.HandlerFunc(srv.getTriggerStateHistory).ServeHTTP(r, req)
		assert.Equal(t, v, r.Code, "response code %s", k)
	}
}

// Tests forbidden device history.
func TestGetStateHistoryForbidden(t *testing.T) {
	input := map[string]int{
		"dev1": http.StatusForbidden,
		"dev2": http.StatusInternalServerError,
		"g1":   http.StatusForbidden,
		"g2":   http.StatusInternalServerError,
	}

	monkey.Patch(getContextUser, func(_ *http.Request) providers.IAuthenticatedUser {
		return &security.AuthenticatedUser{
			Username: "test",
			Rules: map[providers.SecSystem][]*providers.BakedRule{
				providers.SecSystemDevice: {
					{
						Get:     true,
						Command: true,
						History: false,
						Resources: []glob.Glob{
							compileRegexp("*"),
						},
					},
				},
			},
		}
	})
	defer monkey.UnpatchAll()

	srv := getServer()
	for k, v := range input {
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err, "setup failed %s", k)
		req = mux.SetURLVars(req, map[string]string{string(urlDeviceID): k})

		r := httptest.NewRecorder()
		http.HandlerFunc(srv.getDeviceStateHistory).ServeHTTP(r, req)
		assert.Equal(t, v, r.Code, "response code %s", k)
	}
}
