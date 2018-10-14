package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/gobwas/glob"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type wsSuite struct {
	suite.Suite

	s       providers.ISettingsProvider
	worker1 bool
	worker2 bool
	group   bool

	ts *httptest.Server
	ws *websocket.Conn
}

//noinspection GoUnhandledErrorResult
func (w *wsSuite) SetupTest() {
	w.s = getFakeSettings(func(name string, msg ...interface{}) {
		switch name {
		case "1":
			w.worker1 = true
		case "2":
			w.worker2 = true
		}

	}, nil, nil)

	state := newServerState(w.s)
	state.KnownDevices = map[string]*knownDevice{
		"dev1": {ID: "dev1", Commands: []string{enums.CmdOn.String(), enums.CmdSetBrightness.String()},
			Worker: "1", State: map[string]interface{}{"test": "test"}},
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

	srv := &GoHomeServer{
		state:    state,
		Logger:   mocks.FakeNewLogger(nil),
		Settings: w.s,
		groups: map[string]providers.IGroupProvider{
			"g1": mocks.FakeNewGroupProvider("g1", []string{"dev1"}, func() {
				w.group = true
			}),
		},
	}

	w.ts = httptest.NewServer(http.HandlerFunc(srv.handleWS))

	monkey.Patch(getContextUser, func(request *http.Request) *providers.AuthenticatedUser {
		return user
	})
	defer monkey.UnpatchAll()

	u := "ws" + strings.TrimPrefix(w.ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(w.T(), err, "dial")
	w.ws = ws
}

//noinspection GoUnhandledErrorResult
func (w *wsSuite) TearDownTest() {
	if nil != w.ts {
		w.ts.Close()
	}
	if nil != w.ws {
		w.ws.Close()
	}
}

// Tests ping.
//noinspection GoUnhandledErrorResult
func (w *wsSuite) TestPing() {
	w.ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	w.ws.WriteMessage(websocket.TextMessage, []byte("ping"))
	wt, msg, err := w.ws.ReadMessage()
	require.NoError(w.T(), err, "ping")
	assert.Equal(w.T(), websocket.TextMessage, wt, "ping type")
	assert.Equal(w.T(), "pong", string(msg), "pong response")
}

// Tests proper updates.
//noinspection GoUnhandledErrorResult
func (w *wsSuite) TestCalls() {
	data := []struct {
		ID  string
		w1  bool
		w2  bool
		g   bool
		msg string
	}{
		{
			ID:  "dev1",
			w1:  true,
			w2:  false,
			g:   false,
			msg: "allowed",
		},
		{
			ID:  "dev2",
			w1:  false,
			w2:  false,
			g:   false,
			msg: "forbidden",
		},
		{
			ID: "g1",
			w1: false,
			w2: false,
			g:  true,
		},
	}

	for _, v := range data {
		w.worker1 = false
		w.worker2 = false
		w.group = false

		w.ws.WriteJSON(&wsCmd{
			ID:  v.ID,
			Cmd: "on",
			Val: nil,
		})

		time.Sleep(1 * time.Second)
		assert.Equal(w.T(), v.w1, w.worker1, v.msg+" worker 1")
		assert.Equal(w.T(), v.w2, w.worker2, v.msg+" worker 2")
		assert.Equal(w.T(), v.g, w.group, v.msg+" group")
	}
}

// Tests update callbacks.
//noinspection GoUnhandledErrorResult
func (w *wsSuite) TestUpdate() {
	w.s.FanOut().ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID: "dev1",
	}

	w.ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	wt, msg, err := w.ws.ReadMessage()
	require.NoError(w.T(), err, "error")
	require.Equal(w.T(), websocket.TextMessage, wt, "type")

	d := &knownDevice{}
	err = json.Unmarshal(msg, d)
	require.NoError(w.T(), err, "json")
	assert.Equal(w.T(), "dev1", d.ID, "wrong device")
	assert.Equal(w.T(), "test", d.State["test"].(string), "wrong state")
}

// Tests WS connection.
func TestWs(t *testing.T) {
	suite.Run(t, new(wsSuite))

}
