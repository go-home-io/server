// Package worker contains go-home worker nodes logic.
package worker

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	busPlugin "github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
)

const (
	// Default logger system.
	logSystem = "worker"
)

// GoHomeWorker node definition.
type GoHomeWorker struct {
	Settings      providers.ISettingsProvider
	Logger        common.ILoggerProvider
	MessageParser bus.IWorkerMessageParserProvider

	workerChan chan busPlugin.RawMessage

	state IWorkerStateProvider
}

// NewWorker constructs a go-home worker.
// settings holds details from parsed yaml and all necessary helper-providers.
// nolint: dupl
func NewWorker(settings providers.ISettingsProvider) (*GoHomeWorker, error) {
	worker := GoHomeWorker{
		Logger:        settings.SystemLogger(),
		Settings:      settings,
		MessageParser: bus.NewWorkerMessageParser(settings.SystemLogger()),

		workerChan: make(chan busPlugin.RawMessage, 20),

		state: newWorkerState(settings),
	}

	return &worker, nil
}

// Start a go-home worker.
//noinspection GoUnhandledErrorResult
func (w *GoHomeWorker) Start() {
	w.busStart()

	w.sendDiscovery(true)
	w.Settings.Cron().AddFunc("@every 1m", func() {
		w.sendDiscovery(false)
	})

	w.Logger.Info("Successfully started go-home worker",
		"max_devices", strconv.Itoa(w.Settings.WorkerSettings().MaxDevices))

	w.busCycle()
}

// Starting service-bus listeners.
func (w *GoHomeWorker) busStart() {
	err := w.Settings.ServiceBus().SubscribeToWorker(w.Settings.NodeID(), w.workerChan)
	if err != nil {
		w.Logger.Fatal("Failed to subscribe to worker channel", err, common.LogSystemToken, logSystem)
	}

	w.Logger.Debug("Successfully subscribed to worker channels", common.LogSystemToken, logSystem)
}

// Sending discovery message to the go-home server.
func (w *GoHomeWorker) sendDiscovery(isFirstStart bool) {
	w.Logger.Debug("Sending discovery message", common.LogSystemToken, logSystem)
	w.Settings.ServiceBus().Publish(busPlugin.ChDiscovery, bus.NewDiscoveryMessage(w.Settings.NodeID(), isFirstStart,
		w.Settings.WorkerSettings().Properties, w.Settings.WorkerSettings().MaxDevices))
}

// Processing incoming service-bus messages.
func (w *GoHomeWorker) busCycle() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case msg := <-w.workerChan:
			go w.MessageParser.ProcessIncomingMessage(&msg)
		case assign := <-w.MessageParser.GetDeviceAssignmentMessageChan():
			w.state.DevicesAssignmentMessage(assign)
		case cmd := <-w.MessageParser.GetDeviceCommandMessageChan():
			go w.state.DevicesCommandMessage(cmd)
		case <-c:
			w.Logger.Info("Received stop command, exiting", common.LogSystemToken, logSystem)
			os.Exit(0)
		}
	}
}
