package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-home-io/server/plugins/api"
	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"github.com/go-home-io/server/systems/logger"
	"github.com/go-home-io/server/utils"
	"github.com/gobwas/glob"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// API wrapper provider
type provider struct {
	plugin         api.IExtendedAPI
	name           string
	inChannelName  string
	outChannelName string

	serviceBus providers.IBusProvider
	logger     common.ILoggerProvider
	server     providers.IServerProvider

	inQueue     chan bus.RawMessage
	pluginQueue chan []byte
}

// ConstructAPI has data to create a new API provider.
type ConstructAPI struct {
	Logger             common.ILoggerProvider
	Loader             providers.IPluginLoaderProvider
	Secret             common.ISecretProvider
	Validator          providers.IValidatorProvider
	Provider           string
	Name               string
	RawConfig          []byte
	FanOut             providers.IInternalFanOutProvider
	Server             providers.IServerProvider
	InternalRootRouter *mux.Router
	ExternalAPIRouter  *mux.Router
	IsServer           bool
	ServiceBus         providers.IBusProvider
}

// NewExtendedAPIProvider creates a new API provider.
func NewExtendedAPIProvider(ctor *ConstructAPI) (providers.IExtendedAPIProvider, error) {
	logCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.Logger,
		Provider:     ctor.Provider,
		System:       systems.SysAPI.String(),
		ExtraFields:  map[string]string{common.LogNameToken: ctor.Name},
	}
	log := logger.NewPluginLogger(logCtor)

	w := &provider{
		name:       ctor.Name,
		logger:     log,
		serviceBus: ctor.ServiceBus,
		server:     ctor.Server,
		inQueue:    make(chan bus.RawMessage, 5),
	}

	srv := fmt.Sprintf(bus.ChExtendedAPIFormat, utils.NormalizeDeviceName(ctor.Name), "Srv")
	wkr := fmt.Sprintf(bus.ChExtendedAPIFormat, utils.NormalizeDeviceName(ctor.Name), "Wkr")

	initData := &api.InitDataAPI{
		Logger:       log,
		Secret:       ctor.Secret,
		IsMaster:     ctor.IsServer,
		Communicator: w,
	}

	if ctor.IsServer {
		initData.FanOut = ctor.FanOut
		initData.ExternalAPIRouter = ctor.ExternalAPIRouter
		initData.InternalRootRouter = ctor.InternalRootRouter
		w.inChannelName = srv
		w.outChannelName = wkr
	} else {
		w.inChannelName = wkr
		w.outChannelName = srv
	}

	request := &providers.PluginLoadRequest{
		PluginProvider: ctor.Provider,
		RawConfig:      ctor.RawConfig,
		SystemType:     systems.SysAPI,
		ExpectedType:   api.TypeExtendedAPI,
		InitData:       initData,
	}

	plugin, err := ctor.Loader.LoadPlugin(request)
	if err != nil {
		log.Error("Failed to load api provider", err)
		return nil, errors.Wrap(err, "plugin load failed")
	}

	log.Info("Successfully registered extended API")

	w.plugin = plugin.(api.IExtendedAPI)

	log.Info("Extended API requested URL", "urls", strings.Join(w.plugin.Routes(), " "))
	go w.busCycle()

	return w, nil
}

// ID returns internal API identifier.
func (p *provider) ID() string {
	return p.name
}

// Routes returns list of exposed APIs if any.
func (p *provider) Routes() []string {
	return p.plugin.Routes()
}

// Unload helps to unload plugin. Called only on worker.
func (p *provider) Unload() {
	if nil != p.pluginQueue {
		p.serviceBus.Unsubscribe(p.inChannelName)
	}

	close(p.inQueue)
	p.plugin.Unload()
}

// Subscribe provides ability to subscribe for a new
// API plugin channel.
func (p *provider) Subscribe(queue chan []byte) error {
	p.pluginQueue = queue
	return p.serviceBus.SubscribeStr(p.inChannelName, p.inQueue)
}

// Publish provides ability to send a new message to the
// API plugin channel.
func (p *provider) Publish(messages ...api.IExtendedAPIMessage) {
	msgs := make([]interface{}, 0)
	for _, v := range messages {
		v.SetSendTime(utils.TimeNow())
		msgs = append(msgs, v)
	}
	p.serviceBus.PublishStr(p.outChannelName, msgs...)
}

// InvokeDeviceCommand invokes command. Called from server only.
func (p *provider) InvokeDeviceCommand(deviceRegexp glob.Glob, cmd enums.Command, data map[string]interface{}) {
	if nil == p.server {
		p.logger.Warn("API provider tried to invoke device command from worker")
		return
	}

	p.server.InternalCommandInvokeDeviceCommand(deviceRegexp, cmd, data)
}

// Internal service bus processing cycle.
func (p *provider) busCycle() {
	for msg := range p.inQueue {
		pluginMsg := &api.ExtendedAPIMessage{}
		err := json.Unmarshal(msg.Body, pluginMsg)
		if err != nil {
			p.logger.Error("Received corrupted API message", err)
			continue
		}

		if utils.TimeNow()-pluginMsg.SendTime > bus.MsgTTLSeconds {
			p.logger.Debug("Received API message is too old")
			continue
		}

		if nil != p.pluginQueue {
			p.pluginQueue <- msg.Body
		}
	}
}
