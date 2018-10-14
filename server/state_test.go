package server

import (
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/systems/fanout"
	"github.com/go-home-io/server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Settings mock.
func getFakeSettings(publishCallback func(name string, msg ...interface{}),
	devices []*providers.RawDevice, logCallback func(string)) providers.ISettingsProvider {
	return mocks.FakeNewSettings(publishCallback, false, devices, logCallback)
}

// Service bus patch.
func getSbPatch(published map[string][]string, t *testing.T) func(name string, msg ...interface{}) {
	return func(name string, msg ...interface{}) {
		assert.Equal(t, 1, len(msg), "messages failed")

		r := msg[0].(*bus.DeviceAssignmentMessage)
		cfg := make([]string, 0)
		for _, v := range r.Devices {
			cfg = append(cfg, v.Config)
		}

		published[name] = cfg
	}
}

// Tests whether worker is not picked up if device has selector
// with different worker name.
func TestPickWorkerNotMatchSimple(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "worker-1"},
	}

	device := &providers.RawDevice{
		Selector: &providers.RawDeviceSelector{
			Selectors: map[string]string{"name": "worker-2"},
		},
	}

	results := state.pickWorker(device)
	assert.Equal(t, 0, len(results))
}

// Test whether worker is not picked up if properties are not specified
// and device has simple selectors.
func TestPickWorkerNotMatchNoPropertiesSimple(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{},
	}

	device := &providers.RawDevice{
		Selector: &providers.RawDeviceSelector{
			Selectors: map[string]string{"name": "worker-2"},
		},
	}

	results := state.pickWorker(device)
	assert.Equal(t, 0, len(results))
}

// Tests whether worker is picked up by devices with matching selectors
// and no selectors at all.
func TestPickWorkerSingleMatchSimple(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "worker-1"},
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID:               "2",
		WorkerProperties: map[string]string{},
	}

	device := &providers.RawDevice{
		Selector: &providers.RawDeviceSelector{
			Selectors: map[string]string{"name": "worker-1"},
		},
	}

	results := state.pickWorker(device)
	require.Equal(t, 1, len(results), "wrong count")
	assert.Equal(t, "1", results[0], "wrong worker")
}

// Tests whether error is generated when device has selectors
// with incorrect syntax.
func TestPickWorkerBrokenRegex(t *testing.T) {
	calledTimes := 0
	s := getFakeSettings(nil, nil, func(msg string) {
		if msg == "Device selector misconfiguration" {
			calledTimes++
		}
	})
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID: "1",
		WorkerProperties: map[string]string{
			"location": "san Francisco",
		},
	}

	device := &providers.RawDevice{
		Selector: &providers.RawDeviceSelector{
			Selectors: map[string]string{
				"location": "[!]",
			},
		},
	}

	state.pickWorker(device)
	assert.Equal(t, 1, calledTimes)
}

// Tests whether worker is picked up by matching regexp selectors
// on a device.
func TestPickWorkerSingleMatchRegex(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID: "1",
		WorkerProperties: map[string]string{
			"name":     "worker-1",
			"location": "san Francisco",
		},
	}

	device := &providers.RawDevice{
		Selector: &providers.RawDeviceSelector{
			Selectors: map[string]string{
				"name":     "worker-[0-9]",
				"location": "san*",
			},
		},
	}

	results := state.pickWorker(device)
	assert.Equal(t, 1, len(results))
}

// Tests whether first worker out of two is picked up when all workers matches
// device's selectors.
func TestPickWorkerSingleMatchFromManyRegex(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID: "1",
		WorkerProperties: map[string]string{
			"name":     "worker-1",
			"location": "san Francisco",
		},
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID: "2",
		WorkerProperties: map[string]string{
			"name":     "worker-2",
			"location": "los amos",
		},
	}

	device := &providers.RawDevice{
		Selector: &providers.RawDeviceSelector{
			Selectors: map[string]string{
				"name":     "worker-?",
				"location": "san*",
			},
		},
	}

	results := state.pickWorker(device)
	require.Equal(t, 1, len(results), "count")
	assert.Equal(t, "1", results[0], "worker")
}

// Tests whether worker with non matched selectors is not picked up.
func TestPickWorkerMultipleMatchFromManyRegex(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID: "1",
		WorkerProperties: map[string]string{
			"name":     "worker-1",
			"location": "san Francisco",
		},
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID: "2",
		WorkerProperties: map[string]string{
			"name":     "different-worker",
			"location": "san rafael",
		},
	}

	state.KnownWorkers["3"] = &knownWorker{
		ID: "3",
		WorkerProperties: map[string]string{
			"name":     "worker-3",
			"location": "san-antonio",
		},
	}

	device := &providers.RawDevice{
		Selector: &providers.RawDeviceSelector{
			Selectors: map[string]string{
				"name":     "worker*[0-9]",
				"location": "san[ -]*",
			},
		},
	}

	results := state.pickWorker(device)
	assert.Equal(t, 2, len(results), "count")
	assert.False(t, helpers.SliceContainsString(results, "2"), "wrong worker")
}

// Tests whether warning is generated when no workers is available.
func TestReBalanceNoWorkers(t *testing.T) {
	devices := []*providers.RawDevice{
		{
			StrConfig: "1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
	}

	msgFound := false
	s := getFakeSettings(nil, devices, func(msg string) {
		if msg == "Failed to select a worker for the device" {
			msgFound = true
		}
	})
	state := newServerState(s)

	state.reBalance("")
	assert.True(t, msgFound)
}

// Tests whether warning is generated when worker is at its devices' capacity.
func TestReBalanceTooManyDevices(t *testing.T) {
	max := 3
	devices := make([]*providers.RawDevice, max+2)
	for ii := 0; ii < max+2; ii++ {
		devices[ii] = &providers.RawDevice{
			StrConfig: "1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		}
	}

	msgFound := false
	calledTimes := 0
	s := getFakeSettings(nil, devices, func(msg string) {
		if msg == "Failed to select a worker: too many devices" {
			msgFound = true
			calledTimes++
		}
	})

	state := newServerState(s)
	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{},
		MaxDevices:       max,
	}

	state.reBalance("")
	assert.True(t, msgFound, "not found")
	assert.Equal(t, 2, calledTimes, "calls")
}

// Tests whether re-balance evenly distributes devices among
// available workers.
func TestReBalanceNoSelectors(t *testing.T) {
	devices := []*providers.RawDevice{
		{
			StrConfig: "1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "2",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "3",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "4",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
	}

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{},
		MaxDevices:       999,
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID:               "2",
		WorkerProperties: map[string]string{},
		MaxDevices:       999,
	}

	state.reBalance("")
	assert.Equal(t, 2, len(published), "publish")
	assert.Equal(t, 2, len(published["1"]), "worker 1")
	assert.Equal(t, 2, len(published["2"]), "worker 2")
}

// Tests whether re-balance is prioritizing devices with selectors
// and tries to evenly distribute what's left.
func TestReBalanceWithSelectors(t *testing.T) {
	devices := []*providers.RawDevice{
		{
			StrConfig: "1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "2",
			Selector: &providers.RawDeviceSelector{
				Selectors: map[string]string{"name": "worker-1"},
			},
		},
		{
			StrConfig: "3",
			Selector: &providers.RawDeviceSelector{
				Selectors: map[string]string{"name": "worker-1"},
			},
		},
	}
	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "worker-1"},
		MaxDevices:       999,
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID:               "2",
		WorkerProperties: map[string]string{"name": "worker-2"},
		MaxDevices:       999,
	}

	state.reBalance("")
	assert.Equal(t, 2, len(published), "publish")
	assert.Equal(t, 2, len(published["1"]), "worker 1 calls")
	assert.False(t, helpers.SliceContainsString(published["1"], "1"), "worker 1 assignment")
	assert.Equal(t, 1, len(published["2"]), "worker 2 calls")
	assert.True(t, helpers.SliceContainsString(published["2"], "1"), "worker 2 assignment")
}

// Tests whether regular ping message is not triggering re-balance.
func TestPingMessageNoReBalance(t *testing.T) {
	devices := make([]*providers.RawDevice, 0)

	published := make(map[string][]string)
	logInvoked := false
	s := getFakeSettings(getSbPatch(published, t), devices, func(s string) {
		if s == "Received discovery from a known worker" {
			logInvoked = true
		}
	})
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "1"},
		MaxDevices:       999,
	}

	discovery := &bus.DiscoveryMessage{
		MaxDevices:   999,
		NodeID:       "1",
		IsFirstStart: false,
	}
	state.Discovery(discovery)
	assert.True(t, logInvoked, "log")
	assert.Equal(t, 1, len(published), "calls")
}

// Tests whether after reboot worker will receive it's device
// assignment back.
func TestWorkerRestartNoReBalance(t *testing.T) {
	devices := make([]*providers.RawDevice, 0)

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "1"},
		MaxDevices:       999,
		Devices: []*bus.DeviceAssignment{
			{
				Config: "1",
			},
		},
	}

	discovery := &bus.DiscoveryMessage{
		MaxDevices:   999,
		NodeID:       "1",
		IsFirstStart: true,
	}
	state.Discovery(discovery)
	assert.Equal(t, 1, len(published), "calls")
	assert.Equal(t, 1, len(published["1"]), "worker devices")
	assert.Equal(t, "1", published["1"][0], "device ID")
}

// Tests device re-sending to a worker if it changes property.
// In addition, already running worker shouldn't receive anything
// since there're no updates to his devices.
func TestWorkerPropertiesChangesReBalance(t *testing.T) {
	defer leaktest.Check(t)()

	devices := []*providers.RawDevice{
		{
			StrConfig: "d1",
			Name:      "d1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "d2",
			Name:      "d2",
			Selector: &providers.RawDeviceSelector{
				Selectors: map[string]string{"name": "1"},
			},
		},
		{
			StrConfig: "d3",
			Name:      "d3",
			Selector: &providers.RawDeviceSelector{
				Selectors: map[string]string{"name": "1", "location": "1"},
			},
		},
	}

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "1", "location": "old"},
		MaxDevices:       999,
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d2",
			},
		},
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID:               "2",
		WorkerProperties: map[string]string{"name": "2"},
		MaxDevices:       999,
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d1",
				Name:   "d1",
			},
		},
	}

	discovery := &bus.DiscoveryMessage{
		MaxDevices:   999,
		NodeID:       "1",
		IsFirstStart: true,
		Properties:   map[string]string{"location": "1"},
	}

	state.Discovery(discovery)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, len(published), "calls")
	assert.Equal(t, 2, len(published["1"]), "calls to worker 1")
	assert.True(t, helpers.SliceContainsString(published["1"], "d2"), "device 2")
	assert.True(t, helpers.SliceContainsString(published["1"], "d3"), "device 3")
}

// Test that new worker discovery triggers re-balance if current worker handles
// more that one device.
func TestNewWorkerDiscoveryReBalance(t *testing.T) {
	defer leaktest.Check(t)()

	devices := []*providers.RawDevice{
		{
			StrConfig: "d1",
			Name:      "d1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "d2",
			Name:      "d2",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
	}

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{},
		MaxDevices:       999,
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d2",
				Name:   "d2",
			},
			{
				Config: "d1",
				Name:   "d1",
			},
		},
	}

	discovery := &bus.DiscoveryMessage{
		MaxDevices:   999,
		NodeID:       "2",
		IsFirstStart: true,
		Properties:   map[string]string{},
	}

	state.Discovery(discovery)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 2, len(published))
}

// Test that new worker discovery triggers re-balance if current worker handles
// more that one device.
func TestNewWorkerDiscoveryNoReBalanceWithSelectors(t *testing.T) {
	defer leaktest.Check(t)()

	devices := []*providers.RawDevice{
		{
			StrConfig: "d1",
			Name:      "d1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{"name": "1"}},
		},
		{
			StrConfig: "d2",
			Name:      "d2",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{"name": "1"}},
		},
	}

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "1"},
		MaxDevices:       999,
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d2",
				Name:   "d2",
			},
			{
				Config: "d1",
				Name:   "d1",
			},
		},
	}

	discovery := &bus.DiscoveryMessage{
		MaxDevices:   999,
		NodeID:       "2",
		IsFirstStart: true,
		Properties:   map[string]string{},
	}

	state.Discovery(discovery)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 0, len(published))
}

// Tests that adding a new property results in changing state.
func TestComparePropertiesNewOne(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "1"},
	}

	discovery := &bus.DiscoveryMessage{
		MaxDevices:   999,
		NodeID:       "1",
		IsFirstStart: true,
		Properties:   map[string]string{"location": "1"},
	}

	assert.False(t, state.compareProperties(discovery))
}

// Tests that adding a new property results in changing state.
func TestComparePropertiesUpdatedOne(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:               "1",
		WorkerProperties: map[string]string{"name": "1", "location": "old"},
	}

	discovery := &bus.DiscoveryMessage{
		MaxDevices:   999,
		NodeID:       "1",
		IsFirstStart: true,
		Properties:   map[string]string{"location": "1"},
	}

	assert.False(t, state.compareProperties(discovery))
}

// Tests whether stale check honors last seen.
func TestStaleWorkersAllAreFine(t *testing.T) {
	devices := []*providers.RawDevice{
		{
			StrConfig: "d1",
			Name:      "d1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "d2",
			Name:      "d2",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
	}

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:         "1",
		MaxDevices: 999,
		LastSeen:   utils.TimeNow(),
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d1",
			},
		},
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID:         "2",
		MaxDevices: 999,
		LastSeen:   utils.TimeNow(),
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d2",
			},
		},
	}

	state.checkStaleWorkers()
	assert.Equal(t, 0, len(published))
}

// Tests whether re-balance is triggered when one worker is stale.
func TestStaleWorkersOneIsStale(t *testing.T) {
	devices := []*providers.RawDevice{
		{
			StrConfig: "d1",
			Name:      "d1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
		{
			StrConfig: "d2",
			Name:      "d2",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
	}

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:         "1",
		MaxDevices: 999,
		LastSeen:   utils.TimeNow() - int64(time.Duration(2)*time.Hour),
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d1",
			},
		},
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID:         "2",
		MaxDevices: 999,
		LastSeen:   utils.TimeNow(),
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d2",
			},
		},
	}

	state.checkStaleWorkers()
	require.Equal(t, 1, len(published), "count")
	assert.Equal(t, 2, len(published["2"]), "calls")
}

// Tests whether re-balance is triggered when one worker is stale
// and selectors are honored.
// No re-balance should be called for the second worker.
func TestStaleWorkersOneIsStaleNoWorkersForDevice(t *testing.T) {
	devices := []*providers.RawDevice{
		{
			StrConfig: "d1",
			Name:      "d1",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{"name": "1"}},
		},
		{
			StrConfig: "d2",
			Name:      "d2",
			Selector:  &providers.RawDeviceSelector{Selectors: map[string]string{}},
		},
	}

	published := make(map[string][]string)
	s := getFakeSettings(getSbPatch(published, t), devices, nil)
	state := newServerState(s)

	state.KnownWorkers["1"] = &knownWorker{
		ID:         "1",
		MaxDevices: 999,
		LastSeen:   utils.TimeNow() - int64(time.Duration(2)*time.Hour),
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d1",
			},
		},
	}

	state.KnownWorkers["2"] = &knownWorker{
		ID:         "2",
		MaxDevices: 999,
		LastSeen:   utils.TimeNow(),
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d2",
				Name:   "d2",
			},
		},
	}

	state.checkStaleWorkers()
	assert.Equal(t, 0, len(published))
}

// Tests that state properly returns all known devices.
func TestALllDevicesAreReturned(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"1": {ID: "1"},
		"2": {ID: "2"},
	}

	assert.Equal(t, 2, len(state.GetAllDevices()))
}

// Tests that state properly returns known device.
func TestGetKnownDevice(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"1": {ID: "1"},
		"2": {ID: "2"},
	}

	dev := state.GetDevice("1")
	require.NotNil(t, dev)
	assert.Equal(t, "1", dev.ID)
}

// Tests that state properly returns nothing.
func TestGetUnKnownDevice(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownDevices = map[string]*knownDevice{
		"1": {ID: "1"},
		"2": {ID: "2"},
	}

	dev := state.GetDevice("132")
	assert.Nil(t, dev)
}

// Test that state returns nothing while querying unknown device.
func TestQueryUnknownWorker(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	assert.False(t, state.isWorkerHasSameDevicesAlready("workerId", nil))
}

// Test that state returns nothing while querying unknown device.
func TestQueryUnknownDevice(t *testing.T) {
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.KnownWorkers["1"] = &knownWorker{
		ID:         "1",
		MaxDevices: 999,
		LastSeen:   utils.TimeNow() - int64(time.Duration(2)*time.Hour),
		Devices: []*bus.DeviceAssignment{
			{
				Config: "d1",
				Name:   "d1",
			},
			{
				Config: "d2",
				Name:   "d2",
			},
		},
	}

	have := state.isWorkerHasSameDevicesAlready("1", []*bus.DeviceAssignment{
		{
			Config: "d1",
			Name:   "d1",
		},
		{
			Config: "d2",
			Name:   "d3",
		},
	})

	assert.False(t, have)
}

// Tests various updates.
func TestUpdatesFanOut(t *testing.T) {
	fo := fanout.NewFanOut()
	s := getFakeSettings(nil, nil, nil)
	state := newServerState(s)
	state.fanOut = fo
	_, updates := fo.SubscribeDeviceUpdates()

	var msg *common.MsgDeviceUpdate

	go func() {
		for m := range updates {
			msg = m
		}
	}()

	busMsg := &bus.DeviceUpdateMessage{
		DeviceID: "test",
		State: map[string]interface{}{
			"brightness": 50,
		},
	}

	state.Update(busMsg)
	time.Sleep(1 * time.Second)
	assert.True(t, msg.FirstSeen, "first seen")

	msg = nil
	state.Update(busMsg)
	time.Sleep(1 * time.Second)
	assert.Nil(t, msg, "same message")

	msg = nil
	busMsg.State["brightness"] = 60
	state.Update(busMsg)
	time.Sleep(1 * time.Second)
	assert.NotNil(t, msg, "brightness message")

	msg = nil
	busMsg.State["brightness1"] = 60
	state.Update(busMsg)
	time.Sleep(1 * time.Second)
	assert.Nil(t, msg, "wrong message")

	msg = nil
	busMsg.State["brightness"] = 65
	busMsg.State["on"] = "wrong_bool"
	state.Update(busMsg)
	time.Sleep(1 * time.Second)
	require.NotNil(t, msg, "wrong data message")
	assert.Equal(t, 1, len(msg.State), "wrong data message state")
}
