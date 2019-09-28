// Package trigger contains wrappers around system triggers.
package trigger

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
	pluginTrigger "go-home.io/x/server/plugins/trigger"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
	"go-home.io/x/server/utils"
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
	timezone  *time.Location

	fanOut providers.IInternalFanOutProvider

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
	Name      string
	RawConfig []byte
	FanOut    providers.IInternalFanOutProvider
	Server    providers.IServerProvider
	Timezone  *time.Location
}

// NewTrigger creates a new trigger.
func NewTrigger(ctor *ConstructTrigger) (providers.ITriggerProvider, error) {
	logCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.Logger,
		Provider:     ctor.Provider,
		System:       systems.SysAPI.String(),
		ExtraFields:  map[string]string{common.LogNameToken: ctor.Name, common.LogIDToken: getID(ctor.Name)},
	}
	log := logger.NewPluginLogger(logCtor)

	cfg := &trigger{}
	err := yaml.Unmarshal(ctor.RawConfig, cfg)
	if err != nil {
		log.Error("Failed to unmarshal trigger config", err)
		return nil, errors.Wrap(err, "yaml un-marshal failed")
	}

	w := &wrapper{
		logger:    log,
		name:      ctor.Name,
		validator: ctor.Validator,
		server:    ctor.Server,
		ID:        getID(ctor.Name),
		timezone:  ctor.Timezone,
		fanOut:    ctor.FanOut,
	}
	err = w.loadActions(cfg.Actions)
	if err != nil {
		log.Error("Failed to init trigger provider", err)
		return nil, errors.Wrap(err, "load action failed")
	}

	w.loadActiveWindow(cfg.ActiveHrs)

	callback := make(chan interface{}, 5)

	initData := &pluginTrigger.InitDataTrigger{
		FanOut:    ctor.FanOut,
		Logger:    log,
		Triggered: callback,
		Secret:    ctor.Secret,
		Timezone:  ctor.Timezone,
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
		log.Error("Failed to load trigger provider", err)
		return nil, errors.Wrap(err, "plugin load failed")
	}

	w.trigger = plugin.(pluginTrigger.ITrigger)
	w.triggerChan = callback

	go w.processTriggers()

	return w, nil
}

// GetID returns trigger id.
// Format is name.trigger.
func (w *wrapper) GetID() string {
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
			w.logger.Warn("Unknown trigger system", common.LogSystemToken, s.(string))
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
		return &ErrNoActions{}
	}

	return nil
}

// Loads single device action.
func (w *wrapper) loadDeviceAction(data map[string]interface{}) {
	tmp, err := yaml.Marshal(data)
	if err != nil {
		w.logger.Error("Failed to parse device action: marshal error", err)
		return
	}

	action := &triggerActionDevice{}
	err = yaml.Unmarshal(tmp, action)
	if err != nil {
		w.logger.Error("Failed to parse device action: un-marshal error", err)
		return
	}

	if !w.validator.Validate(action) {
		err := &ErrInvalidActionConfig{}
		w.logger.Error("Failed to validate action", err)
		return
	}

	action.prepEntity, err = glob.Compile(action.Entity)
	if err != nil {
		w.logger.Error("Failed to compile regexp", err)
	}

	action.cmd, err = enums.CommandString(action.Command)
	if err != nil {
		w.logger.Error("Failed to validate action properties: unknown command", err,
			common.LogDeviceCommandToken, action.Command)
		return
	}

	action.Args, err = helpers.CommandPropertyFixYaml(action.Args, action.cmd)
	if err != nil {
		w.logger.Error("Failed to validate action properties", err)
		return
	}

	action.prepArgs = make(map[string]interface{})
	if nil != action.Args {
		tmp, err = yaml.Marshal(action.Args)
		if err != nil {
			w.logger.Error("Failed to parse device action: marshal error", err)
			return
		}

		err = yaml.Unmarshal(tmp, action.prepArgs)
		if err != nil {
			w.logger.Error("Failed to parse device action: un-marshal error", err)
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
		w.logger.Debug("Triggered but outside of active window")
		return
	}

	w.fanOut.ChannelInTriggerUpdates() <- w.ID

	for _, v := range w.deviceActions {
		w.logger.Info("Invoking trigger device action",
			"target_id", v.Entity, common.LogDeviceCommandToken, v.Command)
		w.server.InternalCommandInvokeDeviceCommand(v.prepEntity, v.cmd, v.prepArgs)
	}
}

// Determines whether local time is within operation hours.
func (w *wrapper) isInActiveTimeWindow() bool {
	if !w.activeWindow {
		return true
	}

	nowT := time.Now().In(w.timezone)
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
			w.logger.Warn("Active windows is wrong, ignoring")
			return
		}

		from, err := time.Parse(time.Kitchen, parts[0])
		if err != nil {
			w.logger.Warn("Active windows FROM is wrong, ignoring")
			return
		}

		to, err := time.Parse(time.Kitchen, parts[1])
		if err != nil {
			w.logger.Warn("Active windows TO is wrong, ignoring")
			return
		}

		w.from = from.Hour()*60 + from.Minute()
		w.to = to.Hour()*60 + to.Minute()
		w.activeWindow = true
	}
}

// Returns trigger ID.
func getID(name string) string {
	return fmt.Sprintf("%s.%s", utils.NormalizeDeviceName(name), systems.SysTrigger.String())
}
