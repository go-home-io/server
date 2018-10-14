// Package ui contains UI-related configuration.
package ui

import (
	"sync"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/go-home-io/server/providers"
	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// location implements ILocationProvider
type location struct {
	sync.Mutex

	name      string
	icon      string
	devices   []string
	unmatched []string

	devicesExp  []glob.Glob
	updatesChan chan *common.MsgDeviceUpdate
}

// Devices returns list of devices assigned to a location.
func (l *location) Devices() []string {
	l.Lock()
	defer l.Unlock()

	return l.devices
}

// ID returns location name.
func (l *location) ID() string {
	return l.name
}

// Icon returns user-defined icon.
func (l *location) Icon() string {
	return l.icon
}

// Groups settings.
type locationSettings struct {
	Name    string   `yaml:"name"`
	Icon    string   `yaml:"icon"`
	Devices []string `yaml:"devices"`
}

// ConstructLocation has data required for creating a new location.
type ConstructLocation struct {
	RawConfig []byte
	FanOut    common.IFanOutProvider
	Logger    common.ILoggerProvider
}

// NewLocationProvider creates a new location.
func NewLocationProvider(ctor *ConstructLocation) (providers.ILocationProvider, error) {
	settings := &locationSettings{}
	err := yaml.Unmarshal(ctor.RawConfig, settings)
	if err != nil {
		ctor.Logger.Error("Failed to init location", err)
		return nil, errors.Wrap(err, "yaml un-marshal failed")
	}

	provider := &location{
		devices:    make([]string, 0),
		unmatched:  make([]string, 0),
		name:       settings.Name,
		devicesExp: make([]glob.Glob, 0),
		icon:       settings.Icon,
	}

	for _, v := range settings.Devices {
		g, err := glob.Compile(v)
		if err != nil {
			ctor.Logger.Error("Failed to compile location regexp, skipping", err)
			continue
		}

		provider.devicesExp = append(provider.devicesExp, g)
	}

	_, provider.updatesChan = ctor.FanOut.SubscribeDeviceUpdates()
	go provider.deviceUpdates()
	return provider, nil
}

// Subscribes for devices updates.
func (l *location) deviceUpdates() {
	for msg := range l.updatesChan {
		go l.processDeviceUpdates(msg)
	}
}

// Processes devices update messages.
func (l *location) processDeviceUpdates(msg *common.MsgDeviceUpdate) {
	l.Lock()
	defer l.Unlock()

	if !msg.FirstSeen || helpers.SliceContainsString(l.unmatched, msg.ID) ||
		helpers.SliceContainsString(l.devices, msg.ID) {
		return
	}

	for _, v := range l.devicesExp {
		if v.Match(msg.ID) {
			l.devices = append(l.devices, msg.ID)
			return
		}
	}

	l.unmatched = append(l.unmatched, msg.ID)
}
