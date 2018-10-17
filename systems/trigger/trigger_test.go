package trigger

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-home-io/server/mocks"
	pluginTrigger "github.com/go-home-io/server/plugins/trigger"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Fake plugin implementation.
type fakePlugin struct {
	callback chan interface{}
}

func (f *fakePlugin) FakeInit(i interface{}) {
	f.callback = i.(*pluginTrigger.InitDataTrigger).Triggered
}

func (f *fakePlugin) Init(i *pluginTrigger.InitDataTrigger) error {

	return nil
}

func (f *fakePlugin) testInvoke() {
	f.callback <- true
}

// Tests active window parsing wrong format.
func TestActiveWindowFormat(t *testing.T) {
	w := wrapper{
		logger: mocks.FakeNewLogger(nil),
	}

	data := []struct {
		in   string
		gold bool
	}{
		{
			in:   "05:00PM",
			gold: false,
		},
		{
			in:   "05-08",
			gold: false,
		},
		{
			in:   "17:14-18:14",
			gold: false,
		},
		{
			in:   "12:00PM-17",
			gold: false,
		},
		{
			in:   "2:00PM-4:45PM",
			gold: true,
		},
		{
			in:   "4:45PM-2:00PM",
			gold: true,
		},
	}

	for _, v := range data {
		w.loadActiveWindow(v.in)
		assert.Equal(t, v.gold, w.activeWindow, v.in)
	}
}

// Test invokes within active window.
func TestWithinActiveWindowInvokes(t *testing.T) {
	called := 0
	w := wrapper{
		logger: mocks.FakeNewLogger(nil),
		server: mocks.FakeNewServer(func() {
			called++
		}).(providers.IServerProvider),
		deviceActions: []*triggerActionDevice{{}},
	}

	data := []string{
		fmt.Sprintf("%s-%s", time.Now().Add(-1*time.Hour).Format(time.Kitchen),
			time.Now().Add(1*time.Hour).Format(time.Kitchen)),
		fmt.Sprintf("%s-%s", time.Now().Add(-1*time.Hour).Format(time.Kitchen),
			time.Now().Add(22*time.Hour).Format(time.Kitchen)),
		fmt.Sprintf("11:00PM-%s", time.Now().Add(1*time.Minute).Format(time.Kitchen)),
		fmt.Sprintf("%s-11:00PM", time.Now().Add(-1*time.Minute).Format(time.Kitchen)),
	}

	for _, v := range data {
		called = 0
		w.loadActiveWindow(v)
		assert.True(t, w.activeWindow, "window %s", v)
		w.triggered(nil)
		assert.NotEqual(t, 0, called, "call %s", v)
	}
}

// Tests that no invokes will occur outside of active hrs.
func TestOutsideOfActiveWindowInvokes(t *testing.T) {
	called := 0
	w := wrapper{
		logger: mocks.FakeNewLogger(nil),
		server: mocks.FakeNewServer(func() {
			called++
		}).(providers.IServerProvider),
		deviceActions: []*triggerActionDevice{{}},
	}

	data := []string{
		fmt.Sprintf("%s-%s", time.Now().Add(-3*time.Hour).Format(time.Kitchen),
			time.Now().Add(-2*time.Hour).Format(time.Kitchen)),
		fmt.Sprintf("%s-%s", time.Now().Add(1*time.Hour).Format(time.Kitchen),
			time.Now().Add(-22*time.Hour).Format(time.Kitchen)),
	}

	for _, v := range data {
		called = 0
		w.loadActiveWindow(v)
		assert.True(t, w.activeWindow, "window %s", v)
		w.triggered(nil)
		assert.Equal(t, 0, called, "call %s", v)
	}
}

type wdSuite struct {
	suite.Suite

	ctr                *ConstructTrigger
	foundErrorActions  bool
	foundNoActions     bool
	foundCommand       bool
	foundUnMarshal     bool
	foundUnknownSystem bool
	foundRegexpError   bool
}

func (w *wdSuite) SetupTest() {
	w.ctr = &ConstructTrigger{
		Logger: mocks.FakeNewLogger(func(s string) {
			switch s {
			case "Failed to init trigger provider":
				w.foundNoActions = true
			case "Failed to validate action":
				w.foundErrorActions = true
			case "Failed to validate action properties: unknown command":
				w.foundCommand = true
			case "Failed to unmarshal trigger config":
				w.foundUnMarshal = true
			case "Unknown trigger system":
				w.foundUnknownSystem = true
			case "Failed to compile regexp":
				w.foundRegexpError = true
			}

		}),
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Loader:    mocks.FakeNewPluginLoader(&fakePlugin{}),
		Secret:    mocks.FakeNewSecretStore(nil, false),
		FanOut:    mocks.FakeNewFanOut(),
		Provider:  "test",
	}
}

//noinspection GoUnhandledErrorResult
func (w *wdSuite) newTrigger(data string) {
	w.ctr.RawConfig = []byte(data)
	NewTrigger(w.ctr)
}

// Tests no action in config.
func (w *wdSuite) TestNoAction() {
	data := `
name: test
actions:
  - system: device
`
	w.newTrigger(data)
	assert.True(w.T(), w.foundErrorActions)
	assert.True(w.T(), w.foundNoActions)
}

// Tests un-marshal error.
func (w *wdSuite) TestUnMarshal() {
	data := `
name: test
actions:
  - system: device
  entity: cab_led.light*
`
	w.newTrigger(data)
	assert.True(w.T(), w.foundUnMarshal)
}

// Tests compilation error.
func (w *wdSuite) TestRegexp() {
	data := `
actions:
  - system: device
    entity: "[!]"
    command: on
`
	w.newTrigger(data)
	assert.True(w.T(), w.foundRegexpError)
}

// Tests unknown command
func (w *wdSuite) TestCommand() {
	data := `
actions:
    - system: device
      entity: hub
      command: on1
`
	w.newTrigger(data)
	assert.True(w.T(), w.foundCommand)
}

// Tests unknown system.
func (w *wdSuite) TestSystem() {
	data := `
actions:
    - system: test
      entity: hub
      command: on
`
	w.newTrigger(data)
	assert.True(w.T(), w.foundUnknownSystem)
}

// Tests wrong yaml data parsing.
func TestWrongData(t *testing.T) {
	suite.Run(t, new(wdSuite))
}

// Tests correct invokes.
func TestInvoke(t *testing.T) {
	invoked := false
	fakePlugin := &fakePlugin{}
	ctr := &ConstructTrigger{
		Logger:    mocks.FakeNewLogger(nil),
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Loader:    mocks.FakeNewPluginLoader(fakePlugin),
		Secret:    mocks.FakeNewSecretStore(nil, false),
		FanOut:    mocks.FakeNewFanOut(),
		Provider:  "test",
		Server: mocks.FakeNewServer(func() {
			invoked = true
		}).(providers.IServerProvider),
	}

	ctr.RawConfig = []byte(fmt.Sprintf(`
actions:
    - system: device
      entity: hub
      command: set-color
      args:
         r: 5`))

	_, err := NewTrigger(ctr)
	require.NoError(t, err)
	fakePlugin.testInvoke()
	time.Sleep(1 * time.Second)
	assert.True(t, invoked)
}

// Tests incorrect data.
func TestNoSystemData(t *testing.T) {
	fakePlugin := &fakePlugin{}
	ctr := &ConstructTrigger{
		Logger:    mocks.FakeNewLogger(nil),
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Loader:    mocks.FakeNewPluginLoader(fakePlugin),
		Secret:    mocks.FakeNewSecretStore(nil, false),
		FanOut:    mocks.FakeNewFanOut(),
		Provider:  "test",
	}

	ctr.RawConfig = []byte(fmt.Sprintf(`
actions:
    - entity: hub
      command: "on"`))

	_, err := NewTrigger(ctr)
	assert.Error(t, err)
}

// Tests error loading plugin.
func TestNoPlugin(t *testing.T) {
	ctr := &ConstructTrigger{
		Logger:    mocks.FakeNewLogger(nil),
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Loader:    mocks.FakeNewPluginLoader(nil),
		Secret:    mocks.FakeNewSecretStore(nil, false),
		FanOut:    mocks.FakeNewFanOut(),
		Provider:  "test",
	}
	ctr.RawConfig = []byte(fmt.Sprintf(`
actions:
    - system: device
      entity: hub
      command: set-color
      args:
         r: 5`))

	_, err := NewTrigger(ctr)
	assert.Error(t, err)
}

// Tests invokes outside fo active hrs.
func TestInvokeOutsideOfActiveWindow(t *testing.T) {
	invoked := false
	fakePlugin := &fakePlugin{}
	ctr := &ConstructTrigger{
		Logger:    mocks.FakeNewLogger(nil),
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Name:      "test",
		Loader:    mocks.FakeNewPluginLoader(fakePlugin),
		Secret:    mocks.FakeNewSecretStore(nil, false),
		FanOut:    mocks.FakeNewFanOut(),
		Provider:  "test",
		Server: mocks.FakeNewServer(func() {
			invoked = true
		}).(providers.IServerProvider),
	}
	ctr.RawConfig = []byte(fmt.Sprintf(`
activeHrs: %s-%s
actions:
    - system: device
      entity: hub
      command: "on"`, time.Now().Add(-3*time.Hour).Format(time.Kitchen),
		time.Now().Add(-2*time.Hour).Format(time.Kitchen)))

	pl, err := NewTrigger(ctr)
	assert.NoError(t, err)
	fakePlugin.testInvoke()
	time.Sleep(1 * time.Second)
	assert.False(t, invoked, "wrong invoke")
	assert.Equal(t, "test.trigger", pl.GetID(), "wrong ID")
}
