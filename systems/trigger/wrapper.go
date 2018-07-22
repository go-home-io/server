// Package trigger contains wrappers around system triggers.
package trigger

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	pluginTrigger "github.com/go-home-io/server/plugins/trigger"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"github.com/go-home-io/server/utils"
	"github.com/gobwas/glob"
	"gopkg.in/yaml.v2"
)

// Describes trigger wrapper object.
type wrapper struct {
	trigger   pluginTrigger.ITrigger
	logger    common.ILoggerProvider
	validator providers.IValidatorProvider
	ID        string
	name      string
	server    providers.IServerProvider

	deviceActions []*triggerActionDevice

	triggerChan chan interface{}

	activeWindow bool
	from         int
	to           int
}

// ConstructTrigger has data required to create a new trigger.
type ConstructTrigger struct {
	Logger    common.ILoggerProvider
	Loader    providers.IPluginLoaderProvider
	Secret    common.ISecretProvider
	Validator providers.IValidatorProvider
	Provider  string
	RawConfig []byte
	FanOut    providers.IInternalFanOutProvider
	Server    providers.IServerProvider
}

// NewTrigger creates a new trigger.
func NewTrigger(ctor *ConstructTrigger) (providers.ITriggerProvider, error) {
	cfg := &trigger{}
	err := yaml.Unmarshal(ctor.RawConfig, cfg)
	if err != nil {
		ctor.Logger.Error("Failed to unmarshal trigger config", err)
		return nil, err
	}

	if "" == cfg.Name {
		cfg.Name = strconv.FormatInt(utils.TimeNow()+rand.Int63(), 10)
		ctor.Logger.Warn("Trigger name is not specified, going to user random one", "name", cfg.Name)
	}

	w := &wrapper{
		logger:    ctor.Logger,
		name:      cfg.Name,
		validator: ctor.Validator,
		server:    ctor.Server,
	}
	err = w.loadActions(cfg.Actions)
	if err != nil {
		ctor.Logger.Error("Failed to init trigger provider", err, "name", cfg.Name)
		return nil, err
	}

	w.loadActiveWindow(cfg.ActiveHrs)

	callback := make(chan interface{}, 5)

	initData := &pluginTrigger.InitDataTrigger{
		FanOut:    ctor.FanOut,
		Logger:    ctor.Logger,
		Triggered: callback,
		Secret:    ctor.Secret,
	}

	request := &providers.PluginLoadRequest{
		PluginProvider: ctor.Provider,
		RawConfig:      ctor.RawConfig,
		SystemType:     systems.SysTrigger,
		ExpectedType:   pluginTrigger.TypeTrigger,
		InitData:       initData,
	}

	plugin, err := ctor.Loader.LoadPlugin(request)
	if err != nil {
		ctor.Logger.Error("Failed to load trigger provider", err, "name", cfg.Name)
		return nil, err
	}

	w.trigger = plugin.(pluginTrigger.ITrigger)
	w.triggerChan = callback

	go w.processTriggers()

	return w, nil
}

// GetID returns trigger id.
// Format is name.trigger.
func (w *wrapper) GetID() string {
	if "" == w.ID {
		w.ID = fmt.Sprintf("%s.%s", utils.NormalizeDeviceName(w.name), systems.SysTrigger.String())
	}

	return w.ID
}

// Loads all trigger actions.
func (w *wrapper) loadActions(data []map[string]interface{}) error {
	w.deviceActions = make([]*triggerActionDevice, 0)

	for _, v := range data {
		s, ok := v[system]
		if !ok {
			continue
		}

		sys, err := triggerSystemString(s.(string))
		if err != nil {
			w.logger.Warn("Unknown trigger system", common.LogSystemToken, s.(string), "name", w.ID)
			continue
		}

		switch sys {
		case triggerDevice:
			w.loadDeviceAction(v)
		case triggerScript:
			w.loadScriptAction(v)
		}
	}

	if 0 == len(w.deviceActions) {
		return errors.New("no actions defined")
	}

	return nil
}

// Loads single device action.
func (w *wrapper) loadDeviceAction(data map[string]interface{}) {
	tmp, err := yaml.Marshal(data)
	if err != nil {
		w.logger.Error("Failed to parse device action: marshal error", err, "name", w.ID)
		return
	}

	action := &triggerActionDevice{}
	err = yaml.Unmarshal(tmp, action)
	if err != nil {
		w.logger.Error("Failed to parse device action: un-marshal error", err, "name", w.ID)
		return
	}

	if !w.validator.Validate(action) {
		err := errors.New("invalid action config")
		w.logger.Error("Failed to validate action", err, "name", w.ID)
		return
	}

	action.prepEntity, err = glob.Compile(action.Entity)
	if err != nil {
		w.logger.Error("Failed to compile regexp", err, "name", w.ID)
	}

	action.cmd, err = enums.CommandString(action.Command)
	if err != nil {
		w.logger.Error("Failed to validate action properties: unknown command", err,
			"name", w.ID, common.LogDeviceCommandToken, action.Command)
		return
	}

	action.Args, err = helpers.CommandPropertyFixYaml(action.Args, action.cmd)
	if err != nil {
		w.logger.Error("Failed to validate action properties", err, "name", w.ID)
		return
	}

	action.prepArgs = make(map[string]interface{})
	if nil != action.Args {
		tmp, err = yaml.Marshal(action.Args)
		if err != nil {
			w.logger.Error("Failed to parse device action: marshal error", err, "name", w.ID)
			return
		}

		err = yaml.Unmarshal(tmp, action.prepArgs)
		if err != nil {
			w.logger.Error("Failed to parse device action: un-marshal error", err, "name", w.ID)
			return
		}

	}

	w.deviceActions = append(w.deviceActions, action)
}

// Loads single script action.
func (w *wrapper) loadScriptAction(data map[string]interface{}) {

}

// Processes trigger provider callback-channel messages.
func (w *wrapper) processTriggers() {
	for msg := range w.triggerChan {
		go w.triggered(msg)
	}
}

// Processes actual event.
// TODO: Remove nolint after scripts implementation.
func (w *wrapper) triggered(msg interface{}) { //nolint:unparam
	if !w.isInActiveTimeWindow() {
		w.logger.Debug("Triggered but outside of active window", "name", w.ID, common.LogDeviceNameToken)
		return
	}

	for _, v := range w.deviceActions {
		w.logger.Info("Invoking trigger device action", "name", w.ID, common.LogDeviceNameToken,
			v.Entity, common.LogDeviceCommandToken, v.Command)
		w.server.InternalCommandInvokeDeviceCommand(v.prepEntity, v.cmd, v.prepArgs)
	}
}

// Determines whether local time is within operation hours.
func (w *wrapper) isInActiveTimeWindow() bool {
	if !w.activeWindow {
		return true
	}

	nowT := time.Now().Local()
	now := nowT.Hour()*60 + nowT.Minute()
	if w.from > w.to {
		return now <= w.to || now >= w.from
	}

	return w.from <= now && now <= w.to
}

// Loads active time window.
func (w *wrapper) loadActiveWindow(window string) {
	w.activeWindow = false
	if len(window) > 0 {
		parts := strings.Split(window, "-")
		if len(parts) != 2 {
			w.logger.Warn("Active windows is wrong, ignoring", "name", w.ID)
			return
		}

		from, err := time.Parse(time.Kitchen, parts[0])
		if err != nil {
			w.logger.Warn("Active windows FROM is wrong, ignoring", "name", w.ID)
			return
		}

		to, err := time.Parse(time.Kitchen, parts[1])
		if err != nil {
			w.logger.Warn("Active windows TO is wrong, ignoring", "name", w.ID)
			return
		}

		w.from = from.Hour()*60 + from.Minute()
		w.to = to.Hour()*60 + to.Minute()
		w.activeWindow = true
	}
}
