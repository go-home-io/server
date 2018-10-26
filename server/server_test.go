package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/api"
	"go-home.io/x/server/plugins/bus"
	"go-home.io/x/server/plugins/trigger"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/utils"
)

type fakeAPI struct {
}

func (f *fakeAPI) Init(*api.InitDataAPI) error {
	return nil
}

func (*fakeAPI) Routes() []string {
	return []string{}
}

func (*fakeAPI) Unload() {
}

type fakeTrigger struct {
}

func (*fakeTrigger) Init(*trigger.InitDataTrigger) error {
	return nil
}

// Tests success components loading.
func TestSuccessGroupsAPILoad(t *testing.T) {
	s := getFakeSettings(func(_ string, _ ...interface{}) {}, nil, nil)
	s.(mocks.IFakeSettings).AddMasterComponents(
		[]*providers.RawMasterComponent{{Name: "1", RawConfig: []byte("")}},
		[]*providers.RawMasterComponent{{Name: "1"}, {Name: "2"}},
		nil)
	s.(mocks.IFakeSettings).AddLoader(&fakeAPI{})
	s.(mocks.IFakeSettings).AddMasterSettings(&providers.MasterSettings{
		Locations: []*providers.RawMasterComponent{{Name: "1", RawConfig: []byte("")},
			{Name: "1", RawConfig: []byte(":wrong yaml")}}})

	srv, _ := NewServer(s)
	go srv.Start()

	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, len(srv.(*GoHomeServer).groups), "wrong group count")
	assert.Equal(t, 1, len(srv.(*GoHomeServer).locations), "wrong locations count")
	require.Equal(t, 2, len(srv.(*GoHomeServer).extendedAPIs), "wrong extended api count")
	assert.True(t, srv.(*GoHomeServer).extendedAPIs[0].Loaded, "api not loaded")
	assert.True(t, srv.(*GoHomeServer).extendedAPIs[1].Loaded, "api not loaded")

	srv.PushMasterDeviceUpdate(&providers.MasterDeviceUpdate{
		ID:       "1",
		Commands: []string{"test"},
	})

	group := srv.GetDevice("1")
	assert.NotNil(t, group, "no group")
	assert.Equal(t, "test", group.Commands[0], "group state was not updated")
}

// Tests failed components.
func TestFailedGroupAPILoad(t *testing.T) {
	s := getFakeSettings(func(_ string, _ ...interface{}) {}, nil, nil)
	s.(mocks.IFakeSettings).AddMasterComponents(
		[]*providers.RawMasterComponent{{Name: "1", RawConfig: []byte(":wrong yaml")}},
		[]*providers.RawMasterComponent{{Name: "1"}, {Name: "2"}},
		nil,
	)
	s.(mocks.IFakeSettings).AddLoader(nil)

	srv, _ := NewServer(s)
	go srv.Start()

	time.Sleep(1 * time.Second)
	assert.Equal(t, 0, len(srv.(*GoHomeServer).groups), "wrong group count")
	require.Equal(t, 2, len(srv.(*GoHomeServer).extendedAPIs), "wrong extended api count")
	assert.False(t, srv.(*GoHomeServer).extendedAPIs[0].Loaded, "api loaded")
	assert.False(t, srv.(*GoHomeServer).extendedAPIs[1].Loaded, "api loaded")
}

// Tests trigger loading.
func TestTriggersLoad(t *testing.T) {
	s := getFakeSettings(func(_ string, _ ...interface{}) {}, nil, nil)
	s.(mocks.IFakeSettings).AddMasterComponents(
		nil, nil,
		[]*providers.RawMasterComponent{{Name: "1", RawConfig: []byte(`
name: 1
actions:
  - system: device
    command: "on"
`)}, {Name: "2", RawConfig: []byte(`
name: 2
actions:
  - system: device
`)}})
	s.(mocks.IFakeSettings).AddLoader(&fakeTrigger{})

	srv, _ := NewServer(s)
	go srv.Start()

	time.Sleep(1 * time.Second)
	require.Equal(t, 2, len(srv.(*GoHomeServer).triggers), "wrong trigger count")
	assert.NotEqual(t, srv.(*GoHomeServer).triggers[0].Loaded, srv.(*GoHomeServer).triggers[1].Loaded,
		"wrong number of loaded triggers")
}

// Tests discovery.
func TestDiscovery(t *testing.T) {
	s := getFakeSettings(func(_ string, _ ...interface{}) {}, nil, nil)
	srv, _ := NewServer(s)
	go srv.Start()

	time.Sleep(1 * time.Second)
	srv.(*GoHomeServer).incomingChan <- bus.RawMessage{
		Body: []byte(fmt.Sprintf(`{"mt": "ping",  "st": %d, "n": "test"}`, utils.TimeNow())),
	}

	time.Sleep(1 * time.Second)
	wks := srv.(*GoHomeServer).state.GetWorkers()
	require.Equal(t, 1, len(wks), "wrong workers count")
	require.Equal(t, "test", wks[0].ID, "wrong worker name")
}
