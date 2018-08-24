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
)

// Tests WS connection.
func TestWsConnection(t *testing.T) {
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

	ts := httptest.NewServer(http.HandlerFunc(srv.handleWS))
	defer ts.Close()

	monkey.Patch(getContextUser, func(request *http.Request) *providers.AuthenticatedUser {
		return user
	})

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	ws.WriteMessage(websocket.TextMessage, []byte("ping"))
	wt, msg, err := ws.ReadMessage()
	if err != nil || wt != websocket.TextMessage || string(msg) != "pong" {
		t.Error("Ping failed")
		t.FailNow()
	}

	ws.WriteJSON(&wsCmd{
		ID:  "dev1",
		Cmd: "on",
		Val: nil,
	})

	time.Sleep(1 * time.Second)

	if !worker1 || worker2 {
		t.Error("Allowed device failed")
		t.Fail()
	}

	ws.WriteJSON(&wsCmd{
		ID:  "dev2",
		Cmd: "on",
		Val: nil,
	})

	worker1 = false
	worker2 = false

	time.Sleep(1 * time.Second)

	if worker1 || worker2 {
		t.Error("Forbidden device failed")
		t.Fail()
	}

	ws.WriteJSON(&wsCmd{
		ID:  "g1",
		Cmd: "on",
		Val: nil,
	})

	time.Sleep(1 * time.Second)

	if !groupCalled {
		t.Error("Group failed")
		t.Fail()
	}

	s.FanOut().ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{
		ID: "dev1",
	}

	ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	wt, msg, err = ws.ReadMessage()
	if err != nil || wt != websocket.TextMessage {
		t.Error("Update failed")
		t.FailNow()
	}

	d := &knownDevice{}
	err = json.Unmarshal(msg, d)

	if err != nil || d.ID != "dev1" || d.State["test"].(string) != "test" {
		t.Error("Update failed with wrong data")
		t.Fail()
	}
}
