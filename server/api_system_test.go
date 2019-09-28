package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bou.ke/monkey"
	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems/bus"
	"go-home.io/x/server/systems/security"
)

// Tests working ping.
func TestPingAPI(t *testing.T) {
	srv := getServer()
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.ping)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "response code")
}

// Tests failed ping.
func TestFailedPingAPI(t *testing.T) {
	srv := getServer()
	srv.Settings.ServiceBus().(mocks.IFakeServiceBus).SetPingError(errors.New("test"))
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.ping)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusInternalServerError, r.Code, "response code")
}

// Tests get workers.
func TestGetWorkersAPI(t *testing.T) {
	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	srv.state.Discovery(&bus.DiscoveryMessage{
		NodeID:       "test",
		MaxDevices:   99,
		IsFirstStart: false,
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getWorkers)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "response code")
	data := make([]*knownWorker, 0)
	err = json.Unmarshal(r.Body.Bytes(), &data)
	assert.NoError(t, err, "wrong response")
	// One master
	assert.Equal(t, 2, len(data), "incorrect devices num")
}

// Tests forbidden workers.
func TestGetWorkersForbiddenAPI(t *testing.T) {
	monkey.Patch(getContextUser, func(_ *http.Request) providers.IAuthenticatedUser {
		return &security.AuthenticatedUser{
			Username: "test",
			Rules: map[providers.SecSystem][]*providers.BakedRule{
				providers.SecSystemAll: {
					{
						Get:     false,
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
	srv.state.Discovery(&bus.DiscoveryMessage{
		NodeID:       "test",
		MaxDevices:   99,
		IsFirstStart: false,
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getWorkers)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusForbidden, r.Code, "response code")
}

// Test get entities.
func TestGetEntitiesStatusAPI(t *testing.T) {
	monkey.Patch(getContextUser, getFakeRootUser)
	defer monkey.UnpatchAll()

	srv := getServer()
	srv.triggers = []*knownMasterComponent{{
		Name: "test",
	}}
	srv.extendedAPIs = []*knownMasterComponent{{
		Name:   "test1",
		Loaded: true,
	}}

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getStatus)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "response code")
	data := make([]*knownEntity, 0)
	err = json.Unmarshal(r.Body.Bytes(), &data)
	assert.NoError(t, err, "wrong response")
	assert.Equal(t, 2, len(data), "incorrect devices num")
}

// Tests forbidden entities.
func TestForbiddenEntitiesStatusAPI(t *testing.T) {
	monkey.Patch(getContextUser, func(_ *http.Request) providers.IAuthenticatedUser {
		return &security.AuthenticatedUser{
			Username: "test",
			Rules: map[providers.SecSystem][]*providers.BakedRule{
				providers.SecSystemAll: {
					{
						Get:     false,
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
	srv.state.Discovery(&bus.DiscoveryMessage{
		NodeID:       "test",
		MaxDevices:   99,
		IsFirstStart: false,
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getStatus)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusForbidden, r.Code, "response code")
}

// Tests forbidden logs.
func TestLogsForbidden(t *testing.T) {
	monkey.Patch(getContextUser, func(_ *http.Request) providers.IAuthenticatedUser {
		return &security.AuthenticatedUser{
			Username: "test",
			Rules: map[providers.SecSystem][]*providers.BakedRule{
				providers.SecSystemCore: {
					{
						Get:     false,
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
	srv.Logger.(mocks.IFakeLogger).HistorySupported(true)

	logR := &common.LogHistoryRequest{}
	j, _ := json.Marshal(logR)

	req, err := http.NewRequest("POST", "/test", bytes.NewReader(j))
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getLogs)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusForbidden, r.Code, "response code")
}

// Tests bad logs request.
func TestLogsBadRequestAnd(t *testing.T) {
	monkey.Patch(getContextUser, func(_ *http.Request) providers.IAuthenticatedUser {
		return &security.AuthenticatedUser{
			Username: "test",
			Rules: map[providers.SecSystem][]*providers.BakedRule{
				providers.SecSystemCore: {
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
	srv.Logger.(mocks.IFakeLogger).HistorySupported(true)

	req, err := http.NewRequest("POST", "/test", bytes.NewReader([]byte("{P")))
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getLogs)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusInternalServerError, r.Code, "response code")
}

// Tests correct logs query.
func TestLogsCorrectQuery(t *testing.T) {
	monkey.Patch(getContextUser, func(_ *http.Request) providers.IAuthenticatedUser {
		return &security.AuthenticatedUser{
			Username: "test",
			Rules: map[providers.SecSystem][]*providers.BakedRule{
				providers.SecSystemCore: {
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
	srv.Logger.(mocks.IFakeLogger).HistorySupported(true)

	req, err := http.NewRequest("POST", "/test", bytes.NewReader([]byte("{}")))
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getLogs)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "response code")
}

// Tests logs query while logs history is not supported.
func TestLogsNotSupported(t *testing.T) {
	monkey.Patch(getContextUser, func(_ *http.Request) providers.IAuthenticatedUser {
		return &security.AuthenticatedUser{
			Username: "test",
			Rules: map[providers.SecSystem][]*providers.BakedRule{
				providers.SecSystemCore: {
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
	srv.Logger.(mocks.IFakeLogger).HistorySupported(false)

	req, err := http.NewRequest("POST", "/test", bytes.NewReader([]byte("{}")))
	require.NoError(t, err, "setup failed")

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.getLogs)
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusForbidden, r.Code, "response code")
}
