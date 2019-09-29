//go:generate statik -src=./../public -f

// Package server contains go-home server.
package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rakyll/statik/fs"
	busPlugin "go-home.io/x/server/plugins/bus"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
	_ "go-home.io/x/server/server/statik" // Importing statik auto-generated files.
	"go-home.io/x/server/systems/api"
	"go-home.io/x/server/systems/bus"
	"go-home.io/x/server/systems/group"
	"go-home.io/x/server/systems/notification"
	"go-home.io/x/server/systems/trigger"
	"go-home.io/x/server/systems/ui"
)

const (
	// Logger system representation.
	logSystem = "server"
)

type knownMasterComponent struct {
	Loaded    bool
	Name      string
	Interface interface{}
}

// GoHomeServer describes master node.
type GoHomeServer struct {
	Settings      providers.ISettingsProvider
	Logger        common.ILoggerProvider
	MessageParser bus.IMasterMessageParserProvider
	incomingChan  chan busPlugin.RawMessage

	state IServerStateProvider

	triggers      []*knownMasterComponent
	extendedAPIs  []*knownMasterComponent
	notifications []*knownMasterComponent
	groups        map[string]providers.IGroupProvider
	locations     []providers.ILocationProvider

	wsSettings websocket.Upgrader
}

// NewServer constructs a new master server.
// nolint: dupl
func NewServer(settings providers.ISettingsProvider) (providers.IServerProvider, error) {
	server := GoHomeServer{
		Logger:        settings.SystemLogger(),
		Settings:      settings,
		MessageParser: bus.NewMasterMessageParser(settings.SystemLogger()),

		incomingChan: make(chan busPlugin.RawMessage, 100),
	}

	server.state = newServerState(settings)

	return &server, nil
}

// Start launches master server.
func (s *GoHomeServer) Start() {
	prepareCidrs()

	s.startTriggers()
	s.startGroups()
	s.startLocations()
	s.startNotifications()

	s.wsSettings = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}
	router := mux.NewRouter()
	s.registerAPI(router)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", s.Settings.MasterSettings().Port),
			handlers.CORS(
				handlers.AllowedOrigins([]string{"*"}),
				handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost}),
				handlers.AllowedHeaders([]string{"Accept-Encoding", "Content-Type", "Connection",
					"Host", "Origin", "User-Agent", "Referer", "Authorization"}),
				handlers.AllowCredentials(),
			)(router))
		if err != nil {
			s.Logger.Fatal("Failed to start server", err, common.LogSystemToken, logSystem)
		}
	}()

	s.Logger.Info(fmt.Sprintf("Started server on port %d", s.Settings.MasterSettings().Port),
		common.LogSystemToken, logSystem)
	go func() {
		sl := s.Settings.MasterSettings().DelayedStart
		if sl > 0 {
			time.Sleep(time.Duration(sl) * time.Second)
		}

		s.busStart()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for range c {
		s.Logger.Info("Received stop command, exiting", common.LogSystemToken, logSystem)
		os.Exit(0)
	}
}

// GetDevice returns known device.
func (s *GoHomeServer) GetDevice(id string) *providers.KnownDevice {
	kd := s.state.GetDevice(id)
	if nil == kd {
		return nil
	}

	return &providers.KnownDevice{
		Commands: kd.Commands,
		Worker:   kd.Worker,
		Type:     kd.Type,
	}
}

// PushMasterDeviceUpdate pushed device to known devices state
func (s *GoHomeServer) PushMasterDeviceUpdate(update *providers.MasterDeviceUpdate) {
	msg := &bus.DeviceUpdateMessage{
		State:      update.State,
		Commands:   update.Commands,
		DeviceID:   update.ID,
		DeviceName: update.Name,
		DeviceType: update.Type,
		WorkerID:   "master",
	}

	s.state.Update(msg)
}

// All API registration.
func (s *GoHomeServer) registerAPI(router *mux.Router) {
	sFS, err := fs.New()
	if err != nil {
		s.Logger.Fatal("Failed to enable statik", err)
	}

	publicRouter := router.PathPrefix("/pub").Subrouter()
	publicRouter.HandleFunc("/ping", s.ping).Methods(http.MethodGet)

	apiRouter := router.PathPrefix(routeAPI).Subrouter()
	apiRouter.HandleFunc("/ws", s.handleWS)
	apiRouter.HandleFunc("/device", s.getDevices).Methods(http.MethodGet)
	apiRouter.HandleFunc(fmt.Sprintf("/state/device/{%s}", urlDeviceID),
		s.getDeviceStateHistory).Methods(http.MethodGet)
	apiRouter.HandleFunc(fmt.Sprintf("/state/trigger/{%s}", urlTriggerID),
		s.getTriggerStateHistory).Methods(http.MethodGet)
	apiRouter.HandleFunc(fmt.Sprintf("/device/{%s}/{%s}", urlDeviceID, urlCommandName),
		s.deviceCommand).Methods(http.MethodPost)
	apiRouter.HandleFunc("/group", s.getGroups).Methods(http.MethodGet)
	apiRouter.HandleFunc("/state", s.getCurrentState).Methods(http.MethodGet)
	apiRouter.HandleFunc("/worker", s.getWorkers).Methods(http.MethodGet)
	apiRouter.HandleFunc("/status", s.getStatus).Methods(http.MethodGet)
	apiRouter.HandleFunc("/logs", s.getLogs).Methods(http.MethodPost)

	apiRouter.Use(s.logMiddleware)
	router.Use(s.authMiddleware)

	router.PathPrefix("/").Handler(http.FileServer(sFS))

	s.startAPI(router, apiRouter)
}

// Starting bus communications.
func (s *GoHomeServer) busStart() {
	err := s.Settings.ServiceBus().Subscribe(busPlugin.ChDiscovery, s.incomingChan)
	if err != nil {
		s.Logger.Fatal("Failed to subscribe to discovery channel", err, common.LogSystemToken, logSystem)
	}

	err = s.Settings.ServiceBus().Subscribe(busPlugin.ChDeviceUpdates, s.incomingChan)

	if err != nil {
		s.Logger.Fatal("Failed to subscribe to updates channel", err, common.LogSystemToken, logSystem)
	}

	s.Logger.Debug("Successfully subscribed to bus channels", common.LogSystemToken, logSystem)
	s.busCycle()
}

// Internal bus cycle.
func (s *GoHomeServer) busCycle() {
	for {
		select {
		case msg := <-s.incomingChan:
			go s.MessageParser.ProcessIncomingMessage(&msg)
		case dis := <-s.MessageParser.GetDiscoveryMessageChan():
			s.state.Discovery(dis)
		case dup := <-s.MessageParser.GetDeviceUpdateMessageChan():
			s.state.Update(dup)
		case load := <-s.MessageParser.GetEntityLoadStatueMessageChan():
			s.state.EntityLoad(load)
		}
	}
}

// Starts triggers.
func (s *GoHomeServer) startTriggers() {
	s.triggers = make([]*knownMasterComponent, 0)
	for _, v := range s.Settings.Triggers() {
		ctor := &trigger.ConstructTrigger{
			Logger:    s.Settings.PluginLogger(),
			Provider:  v.Provider,
			Name:      v.Name,
			RawConfig: v.RawConfig,
			Loader:    s.Settings.PluginLoader(),
			FanOut:    s.Settings.FanOut(),
			Secret:    s.Settings.Secrets(),
			Validator: s.Settings.Validator(),
			Storage:   s.Settings.Storage(),
			Server:    s,
			Timezone:  s.Settings.Timezone(),
		}
		tr, err := trigger.NewTrigger(ctor)
		comp := &knownMasterComponent{
			Name:      v.Name,
			Interface: tr,
			Loaded:    true,
		}

		if err != nil {
			comp.Loaded = false
		}

		s.triggers = append(s.triggers, comp)
	}
}

// Starts APIs.
func (s *GoHomeServer) startAPI(root *mux.Router, external *mux.Router) {
	s.extendedAPIs = make([]*knownMasterComponent, 0)
	for _, v := range s.Settings.ExtendedAPIs() {
		ctor := &api.ConstructAPI{
			Provider:           v.Provider,
			RawConfig:          v.RawConfig,
			Server:             s,
			Validator:          s.Settings.Validator(),
			Logger:             s.Settings.PluginLogger(),
			Secret:             s.Settings.Secrets(),
			Loader:             s.Settings.PluginLoader(),
			FanOut:             s.Settings.FanOut(),
			InternalRootRouter: root,
			ExternalAPIRouter:  external,
			IsServer:           true,
			ServiceBus:         s.Settings.ServiceBus(),
			Name:               v.Name,
		}

		a, err := api.NewExtendedAPIProvider(ctor)
		comp := &knownMasterComponent{
			Name:      v.Name,
			Interface: a,
			Loaded:    true,
		}

		if err != nil {
			comp.Loaded = false
		}

		s.extendedAPIs = append(s.extendedAPIs, comp)
	}
}

// Starts groups.
func (s *GoHomeServer) startGroups() {
	s.groups = make(map[string]providers.IGroupProvider)

	for _, v := range s.Settings.Groups() {
		ctor := &group.ConstructGroup{
			RawConfig: v.RawConfig,
			Settings:  s.Settings,
			Server:    s,
		}

		g, err := group.NewGroupProvider(ctor)
		if err != nil {
			continue
		}

		s.groups[g.ID()] = g
	}
}

// Starts locations.
func (s *GoHomeServer) startLocations() {
	s.locations = make([]providers.ILocationProvider, 0)

	for _, v := range s.Settings.MasterSettings().Locations {
		ctor := &ui.ConstructLocation{
			RawConfig: v.RawConfig,
			Logger:    s.Settings.SystemLogger(),
			FanOut:    s.Settings.FanOut(),
		}

		l, err := ui.NewLocationProvider(ctor)
		if err != nil {
			continue
		}

		s.locations = append(s.locations, l)
	}
}

// Starts notification systems.
func (s *GoHomeServer) startNotifications() {
	s.notifications = make([]*knownMasterComponent, 0)

	for _, v := range s.Settings.Notifications() {
		ctor := &notification.ConstructNotification{
			Name:      v.Name,
			Provider:  v.Provider,
			Loader:    s.Settings.PluginLoader(),
			RawConfig: v.RawConfig,
			Logger:    s.Settings.SystemLogger(),
			Secret:    s.Settings.Secrets(),
		}

		n, err := notification.NewNotificationProvider(ctor)
		comp := &knownMasterComponent{
			Loaded:    true,
			Name:      v.Name,
			Interface: n,
		}

		if err != nil {
			comp.Loaded = false
		}

		s.notifications = append(s.notifications, comp)
	}
}
