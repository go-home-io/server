// Package group contains groups provider.
package group

import (
	"fmt"
	"sync"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
	"go-home.io/x/server/utils"
	"gopkg.in/yaml.v2"
)

// Implements groups provider.
type provider struct {
	sync.Mutex

	internalID string

	devicesExp  []glob.Glob
	updatesChan chan *common.MsgDeviceUpdate
	logger      common.ILoggerProvider
	server      providers.IServerProvider

	devices   []*groupDevice
	unmatched []string

	Name     string
	State    map[string]interface{}
	Commands []string
}

// Groups settings.
type settings struct {
	Name    string   `yaml:"name"`
	Devices []string `yaml:"devices"`
}

// Single group device.
type groupDevice struct {
	ID       string
	State    map[enums.Property]interface{}
	Commands []string
	IDExp    glob.Glob
}

// ConstructGroup has data required for instantiating a new group.
type ConstructGroup struct {
	RawConfig []byte
	Settings  providers.ISettingsProvider
	Server    providers.IServerProvider
}

// NewGroupProvider creates a new group provider.
func NewGroupProvider(ctor *ConstructGroup) (providers.IGroupProvider, error) {

	settings := &settings{}
	err := yaml.Unmarshal(ctor.RawConfig, settings)
	if err != nil {
		ctor.Settings.SystemLogger().Error("Failed to load group", err)
		return nil, errors.Wrap(err, "yaml un-marshal failed")
	}

	logCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.Settings.PluginLogger(),
		Provider:     systems.SysDevice.String(),
		System:       "group",
		ExtraFields: map[string]string{
			common.LogNameToken: settings.Name,
			common.LogIDToken:   getID(settings.Name),
		},
	}
	log := logger.NewPluginLogger(logCtor)

	provider := &provider{
		logger:     log,
		devicesExp: make([]glob.Glob, 0),
		internalID: getID(settings.Name),
		Name:       settings.Name,
		devices:    make([]*groupDevice, 0),
		unmatched:  make([]string, 0),
		Commands:   make([]string, 0),
		State:      make(map[string]interface{}),
		server:     ctor.Server,
	}

	for _, v := range settings.Devices {
		exp, err := glob.Compile(v)
		if err != nil {
			provider.logger.Error("Failed to compile group regexp, skipping", err)
			continue
		}

		provider.devicesExp = append(provider.devicesExp, exp)
	}

	fanOut := ctor.Settings.FanOut()
	_, provider.updatesChan = fanOut.SubscribeDeviceUpdates()
	go provider.deviceUpdates()

	return provider, nil
}

// ID returns group internal ID.
func (p *provider) ID() string {
	return p.internalID
}

// Devices returns list of devices assigned to the group.
func (p *provider) Devices() []string {
	p.Lock()
	defer p.Unlock()

	ids := make([]string, 0)
	for _, v := range p.devices {
		ids = append(ids, v.ID)
	}

	return ids
}

// InvokeCommand invokes commands on all group's devices.
func (p *provider) InvokeCommand(cmd enums.Command, props map[string]interface{}) {
	p.Lock()
	defer p.Unlock()

	for _, v := range p.devices {
		p.server.InternalCommandInvokeDeviceCommand(v.IDExp, cmd, props)
	}
}

// Subscribes for devices updates.
func (p *provider) deviceUpdates() {
	for msg := range p.updatesChan {
		go p.processDeviceUpdates(msg)
	}
}

// Processed devices updates.
func (p *provider) processDeviceUpdates(msg *common.MsgDeviceUpdate) {
	p.Lock()
	defer p.Unlock()

	if msg.ID == p.internalID {
		return
	}

	if helpers.SliceContainsString(p.unmatched, msg.ID) {
		return
	}

	kd := p.server.GetDevice(msg.ID)
	if nil == kd || kd.Type == enums.DevGroup {
		return
	}

	found := false
	for _, v := range p.devices {
		if v.ID == msg.ID {
			copyState(v.State, msg.State)
			v.Commands = kd.Commands
			found = true
			break
		}
	}

	if !found {
		matched := false
		for _, v := range p.devicesExp {
			if v.Match(msg.ID) {
				matched = true
				break
			}
		}

		if !matched {
			p.unmatched = append(p.unmatched, msg.ID)
			return
		}

		exp, err := glob.Compile(msg.ID)
		if err != nil {
			return
		}

		p.devices = append(p.devices, &groupDevice{
			ID:       msg.ID,
			Commands: kd.Commands,
			State:    msg.State,
			IDExp:    exp,
		})
	}

	p.updateGroupState()
	p.updateGroupCommands()

	p.server.PushMasterDeviceUpdate(&providers.MasterDeviceUpdate{
		Type:     enums.DevGroup,
		Name:     p.Name,
		ID:       p.internalID,
		Commands: p.Commands,
		State:    p.State,
	})
}

// Updates group state.
func (p *provider) updateGroupState() {
	if 0 == len(p.devices) {
		return
	}

	p.State = make(map[string]interface{})
	for k, s := range p.devices[0].State {
		found := true

		for _, v := range p.devices {
			_, ok := v.State[k]
			if !ok {
				found = false
				break
			}
		}

		if !found {
			continue
		}

		p.State[k.String()] = s
	}
}

// Updates available commands.
func (p *provider) updateGroupCommands() {
	p.Commands = make([]string, 0)

	for _, c := range p.devices[0].Commands {
		found := true
		for _, v := range p.devices {
			if !helpers.SliceContainsString(v.Commands, c) {
				found = false
				break
			}
		}

		if found {
			p.Commands = append(p.Commands, c)
		}
	}
}

// Converts ID.
func getID(name string) string {
	return fmt.Sprintf("group.%s", utils.NormalizeDeviceName(name))
}

// Copies state.
func copyState(from, to map[enums.Property]interface{}) {
	for k, v := range to {
		from[k] = v
	}
}
