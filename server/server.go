// Package server contains go-home server.
package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	busPlugin "github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/gorilla/mux"
)

const (
	// Logger system representation.
	logSystem = "server"
)

// GoHomeServer describes master node.
type GoHomeServer struct {
	Settings      providers.ISettingsProvider
	Logger        common.ILoggerProvider
	MessageParser bus.IMasterMessageParserProvider

	incomingChan chan busPlugin.RawMessage

	state IServerStateProvider
}

// NewServer constructs a new master server.
// nolint: dupl
func NewServer(settings providers.ISettingsProvider) (*GoHomeServer, error) {
	server := GoHomeServer{
		Logger:        settings.SystemLogger(),
		Settings:      settings,
		MessageParser: bus.NewServerMessageParser(settings.SystemLogger()),

		incomingChan: make(chan busPlugin.RawMessage, 100),
		state:        newServerState(settings),
	}

	return &server, nil
}

// Start launches master server.
func (s *GoHomeServer) Start() {

	router := mux.NewRouter()
	s.registerAPI(router)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", s.Settings.MasterSettings().Port), router)
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

// All API registration.
func (s *GoHomeServer) registerAPI(router *mux.Router) {
	publicRouter := router.PathPrefix("/pub").Subrouter()
	publicRouter.HandleFunc("/ping", s.ping).Methods(http.MethodGet)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/device", s.getDevices).Methods(http.MethodGet)
	apiRouter.HandleFunc(fmt.Sprintf("/device/{%s}/{%s}", urlDeviceID, urlCommandName),
		s.deviceCommand).Methods(http.MethodPost)
	apiRouter.Use(s.authMiddleware)
	apiRouter.Use(s.logMiddleware)
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
		}
	}
}
