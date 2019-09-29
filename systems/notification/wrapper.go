// Package notification provides provider around notifications plugins.
package notification

import (
	"fmt"

	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/notification"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
	"go-home.io/x/server/utils"
)

// ConstructNotification defines notifications provider constructor.
type ConstructNotification struct {
	Name      string
	Provider  string
	Loader    providers.IPluginLoaderProvider
	RawConfig []byte
	Logger    common.ILoggerProvider
	Secret    common.ISecretProvider
}

// Notification provider wrapper.
type provider struct {
	id     string
	logger common.ILoggerProvider
	plugin notification.INotification
}

// NewNotificationProvider creates a new notification provider.
func NewNotificationProvider(ctor *ConstructNotification) (providers.INotificationProvider, error) {
	p := &provider{
		id: fmt.Sprintf("%s.%s", utils.NormalizeDeviceName(ctor.Name), systems.SysNotification.String()),
	}

	loggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.Logger,
		Provider:     ctor.Provider,
		System:       systems.SysNotification.String(),
		ExtraFields:  map[string]string{common.LogNameToken: ctor.Name, common.LogIDToken: p.id},
	}

	l := logger.NewPluginLogger(loggerCtor)
	p.logger = l

	pluginLoadRequest := &providers.PluginLoadRequest{
		ExpectedType:   notification.TypeNotification,
		SystemType:     systems.SysNotification,
		PluginProvider: ctor.Provider,
		RawConfig:      ctor.RawConfig,
		InitData: &notification.InitDataNotification{
			Logger: l,
			Secret: ctor.Secret,
		},
	}

	plugin, err := ctor.Loader.LoadPlugin(pluginLoadRequest)
	if err != nil {
		l.Error("Failed to load notification system", err)
		return nil, errors.Wrap(err, "failed to load notification system")
	}

	p.plugin = plugin.(notification.INotification)
	return p, nil
}

// GetID returns provider ID.
func (p *provider) GetID() string {
	return p.id
}

// Message sends the message through a plugin.
func (p *provider) Message(msg string) {
	p.logger.Debug("Sending a notification: " + msg)

	err := p.plugin.Message(msg)
	if err != nil {
		p.logger.Error("Failed to send a notification", err)
	}
}
