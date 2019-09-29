package notification

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/notification"
)

type fakePlugin struct {
	called int
}

func (f *fakePlugin) Init(*notification.InitDataNotification) error {
	return nil
}

func (f *fakePlugin) Message(string) error {
	f.called ++
	return errors.New("test")
}

// Tests error during the plugin load.
func TestErrorLoading(t *testing.T) {
	ctor := &ConstructNotification{
		Name:      "",
		Provider:  "",
		Loader:    mocks.FakeNewPluginLoader(nil),
		RawConfig: nil,
		Logger:    mocks.FakeNewLogger(nil),
		Secret:    nil,
	}

	_, err := NewNotificationProvider(ctor)
	assert.Error(t, err)
}

// Tests correct ID.
func TestID(t *testing.T) {
	ctor := &ConstructNotification{
		Name:      "test id",
		Provider:  "",
		Loader:    mocks.FakeNewPluginLoader(&fakePlugin{}),
		RawConfig: nil,
		Logger:    mocks.FakeNewLogger(nil),
		Secret:    nil,
	}

	p, _ := NewNotificationProvider(ctor)
	assert.NotNil(t, p, "failed to create provider")

	assert.Equal(t, "test_id.notification", p.GetID(), "wrong ID")
}

// Tests message sending.
func TestMessage(t *testing.T) {
	f := &fakePlugin{}

	ctor := &ConstructNotification{
		Name:      "test id",
		Provider:  "",
		Loader:    mocks.FakeNewPluginLoader(f),
		RawConfig: nil,
		Logger:    mocks.FakeNewLogger(nil),
		Secret:    nil,
	}

	p, _ := NewNotificationProvider(ctor)
	assert.NotNil(t, p, "failed to create provider")

	p.Message("test")
	assert.Equal(t, 1, f.called, "wrong invoke count")
}