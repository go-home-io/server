package trigger

import (
	"testing"
	"github.com/go-home-io/server/mocks"
	"fmt"
	"time"
	"github.com/go-home-io/server/utils"
	pluginTrigger "github.com/go-home-io/server/plugins/trigger"
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

	w.loadActiveWindow("05:00PM")
	if w.activeWindow {
		t.Fail()
	}

	w.loadActiveWindow("05-08")
	if w.activeWindow {
		t.Fail()
	}

	w.loadActiveWindow("17:14-18:14")
	if w.activeWindow {
		t.Fail()
	}

	w.loadActiveWindow("12:00PM-17")
	if w.activeWindow {
		t.Fail()
	}

	w.loadActiveWindow("2:00PM-4:45PM")
	if !w.activeWindow {
		t.Fail()
	}

	w.loadActiveWindow("4:45PM-2:00PM")
	if !w.activeWindow {
		t.Fail()
	}
}

// Test invokes within active window.
func TestWithinActiveWindowInvokes(t *testing.T) {
	called := 0
	w := wrapper{
		logger: mocks.FakeNewLogger(nil),
		server: mocks.FakeNewServer(func() {
			called ++
		}),
		deviceActions: []*triggerActionDevice{{}},
	}

	timeF := fmt.Sprintf("%s-%s", time.Now().Local().Add(-1 * time.Hour).Format(time.Kitchen),
		time.Now().Local().Add(1 * time.Hour).Format(time.Kitchen))

	w.loadActiveWindow(timeF)
	if !w.activeWindow {
		t.Error("firs window")
		t.Fail()
	}

	w.triggered(nil)

	if 0 == called {
		t.Error("firs call")
		t.Fail()
	}

	called = 0

	timeF = fmt.Sprintf("%s-%s", time.Now().Local().Add(-1 * time.Hour).Format(time.Kitchen),
		time.Now().Local().Add(22 * time.Hour).Format(time.Kitchen))

	w.loadActiveWindow(timeF)
	if !w.activeWindow {
		t.Error("second window")
		t.Fail()
	}

	w.triggered(nil)

	if 0 == called {
		t.Error("second call")
		t.Fail()
	}

	called = 0

	timeF = fmt.Sprintf("11:00PM-%s", time.Now().Local().Add(1 * time.Hour).Format(time.Kitchen))

	w.loadActiveWindow(timeF)
	if !w.activeWindow {
		t.Error("third window")
		t.Fail()
	}

	w.triggered(nil)

	if 0 == called {
		t.Error("third call")
		t.Fail()
	}

	called = 0

	timeF = fmt.Sprintf("%s-11:00PM", time.Now().Local().Add(-1 * time.Hour).Format(time.Kitchen))

	w.loadActiveWindow(timeF)
	if !w.activeWindow {
		t.Error("forth window")
		t.Fail()
	}

	w.triggered(nil)

	if 0 == called {
		t.Error("forth call")
		t.Fail()
	}
}

// Tests that no invokes will occur outside of active hrs.
func TestOutsideOfActiveWindowInvokes(t *testing.T) {
	called := 0
	w := wrapper{
		logger: mocks.FakeNewLogger(nil),
		server: mocks.FakeNewServer(func() {
			called ++
		}),
		deviceActions: []*triggerActionDevice{{}},
	}

	timeF := fmt.Sprintf("%s-%s", time.Now().Local().Add(-3 * time.Hour).Format(time.Kitchen),
		time.Now().Local().Add(-2 * time.Hour).Format(time.Kitchen))

	w.loadActiveWindow(timeF)
	if !w.activeWindow {
		t.Fail()
	}

	w.triggered(nil)

	if 0 != called {
		t.Fail()
	}

	called = 0

	timeF = fmt.Sprintf("%s-%s", time.Now().Local().Add(1 * time.Hour).Format(time.Kitchen),
		time.Now().Local().Add(-22 * time.Hour).Format(time.Kitchen))

	w.loadActiveWindow(timeF)
	if !w.activeWindow {
		t.Fail()
	}

	w.triggered(nil)

	if 0 != called {
		t.Fail()
	}
}

// Tests wrong yaml data parsing.
func TestWrongData(t *testing.T) {
	foundErrorActions := false
	foundNoActions := false
	foundCommand := false
	foundUnMarshal := false
	foundUnknownSystem := false
	foundRegexpError := false

	ctr := &ConstructTrigger{
		Logger: mocks.FakeNewLogger(func(s string) {
			switch s {
			case "Failed to init trigger provider":
				foundNoActions = true
			case "Failed to validate action":
				foundErrorActions = true
			case "Failed to validate action properties: unknown command":
				foundCommand = true
			case "Failed to unmarshal trigger config":
				foundUnMarshal = true
			case "Unknown trigger system":
				foundUnknownSystem = true
			case "Failed to compile regexp":
				foundRegexpError = true
			}

		}),
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Loader:    mocks.FakeNewPluginLoader(&fakePlugin{}),
		Secret:    mocks.FakeNewSecretStore(nil, false),
		FanOut:    mocks.FakeNewFanOut(),
		Provider:  "test",
	}

	ctr.RawConfig = []byte(`
name: test
actions:
  - system: device
`)
	NewTrigger(ctr)
	if !foundErrorActions || !foundNoActions {
		t.Error("action")
		t.Fail()
	}

	ctr.RawConfig = []byte(`
name: test
actions:
  - system: device
  entity: cab_led.light*
`)
	NewTrigger(ctr)
	if !foundUnMarshal {
		t.Error("unmarshal")
		t.Fail()
	}
	ctr.RawConfig = []byte(`
actions:
  - system: device
    entity: "[!]"
    command: on`)
	NewTrigger(ctr)
	if !foundRegexpError{
		t.Error("regexp")
		t.Fail()
	}
	ctr.RawConfig = []byte(`
actions:
    - system: device
      entity: hub
      command: on1`)
	NewTrigger(ctr)
	if !foundCommand {
		t.Error("command")
		t.Fail()
	}

	ctr.RawConfig = []byte(`
actions:
    - system: test
      entity: hub
      command: on`)
	NewTrigger(ctr)
	if !foundUnknownSystem {
		t.Error("system")
		t.Fail()
	}
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
		Server:mocks.FakeNewServer(func() {
			invoked = true
		}),
	}

	ctr.RawConfig = []byte(fmt.Sprintf(`
actions:
    - system: device
      entity: hub
      command: set-color
      args:
         r: 5`))

	_, err := NewTrigger(ctr)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	fakePlugin.testInvoke()

	time.Sleep(1 * time.Second)

	if !invoked {
		t.Fail()
	}
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
	if err == nil {
		t.FailNow()
	}

}

// Tests error loading plugin.
func TestNoPlugin(t *testing.T){
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
	if err == nil {
		t.FailNow()
	}
}

// Tests invokes outside fo active hrs.
func TestInvokeOutsideOfActiveWindow(t *testing.T) {
	invoked := false
	fakePlugin := &fakePlugin{}
	ctr := &ConstructTrigger{
		Logger:    mocks.FakeNewLogger(nil),
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Loader:    mocks.FakeNewPluginLoader(fakePlugin),
		Secret:    mocks.FakeNewSecretStore(nil, false),
		FanOut:    mocks.FakeNewFanOut(),
		Provider:  "test",
		Server:mocks.FakeNewServer(func() {
			invoked = true
		}),
	}
	ctr.RawConfig = []byte(fmt.Sprintf(`
name: test
activeHrs: %s-%s
actions:
    - system: device
      entity: hub
      command: "on"`, time.Now().Local().Add(-3 * time.Hour).Format(time.Kitchen),
		time.Now().Local().Add(-2 * time.Hour).Format(time.Kitchen)))

	pl, err := NewTrigger(ctr)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	fakePlugin.testInvoke()
	time.Sleep(1 * time.Second)

	if invoked {
		t.Fail()
	}

	if pl.GetID() != "test.trigger" {
		t.Fail()
	}
}
