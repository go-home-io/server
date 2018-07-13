package server

import (
	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/utils"
	"testing"
)

func getFakeSettings(publishCallback func(name string, msg ...interface{}), devices []providers.RawDevice, logCallback func(string)) providers.ISettingsProvider {
	return mocks.FakeNewSettings(publishCallback, false, devices, logCallback)
}

func getSbPatch(published map[string][]string, t *testing.T) func(name string, msg ...interface{}) {
	return func(name string, msg ...interface{}) {
		if 1 != len(msg) {
			t.Errorf("Got %d messages", len(msg))
			t.Fail()
		}

		r := msg[0].(*bus.DeviceAssignmentMessage)
		cfg := make([]string, 0)
		for _, v := range r.Devices {
			cfg = append(cfg, v.Config)
		}

		published[name] = cfg
	}
}

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
	if 0 != len(results) {
		t.Fail()
	}
}

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
	if 0 != len(results) {
		t.Fail()
	}
}

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
	if 1 != len(results) {
		t.Fail()
	}

	if "1" != results[0] {
		t.Fail()
	}
}

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
				"location": "((",
			},
		},
	}

	_ = state.pickWorker(device)
	if 1 != calledTimes {
		t.Fail()
	}
}

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
				"name":     "worker-\\d",
				"location": "san.([a-zA-Z]+)",
			},
		},
	}

	results := state.pickWorker(device)
	if 1 != len(results) {
		t.Fail()
	}
}

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
				"name":     "worker-\\d",
				"location": "san.([a-zA-Z]+)",
			},
		},
	}

	results := state.pickWorker(device)
	if 1 != len(results) {
		t.Errorf("Actual selected %d", len(results))
		t.Fail()
	}

	if results[0] != "1" {
		t.Fail()
	}
}

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
				"name":     "worker-\\d",
				"location": "san.([a-zA-Z]+)",
			},
		},
	}

	results := state.pickWorker(device)
	if 2 != len(results) {
		t.Errorf("Actual selected %d", len(results))
		t.Fail()
	}

	if utils.SliceContainsString(results, "2") {
		t.Fail()
	}
}

func TestReBalanceNoWorkers(t *testing.T) {
	devices := []providers.RawDevice{
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

	state.reBalance()
	if !msgFound {
		t.Fail()
	}
}

func TestReBalanceTooManyDevices(t *testing.T) {
	max := 3
	devices := make([]providers.RawDevice, max+2)
	for ii := 0; ii < max+2; ii++ {
		devices[ii] = providers.RawDevice{
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

	state.reBalance()
	if !msgFound || 2 != calledTimes {
		t.Fail()
	}
}

func TestReBalanceNoSelectors(t *testing.T) {
	devices := []providers.RawDevice{
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

	state.reBalance()

	if 2 != len(published) {
		t.Errorf("Published has incorrect data")
		t.Fail()
	}

	if 2 != len(published["1"]) {
		t.Errorf("First worker has incorrect assigments, got %d devices", len(published["1"]))
		t.Fail()
	}

	if 2 != len(published["2"]) {
		t.Errorf("Second worker has incorrect assigments, got %d devices", len(published["2"]))
		t.Fail()
	}
}

func TestReBalanceWithSelectors(t *testing.T) {
	devices := []providers.RawDevice{
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

	state.reBalance()

	if 2 != len(published) {
		t.Errorf("Published has incorrect data")
		t.Fail()
	}

	if 2 != len(published["1"]) || utils.SliceContainsString(published["1"], "1") {
		t.Errorf("First worker has incorrect assigments, got %d devices", len(published["1"]))
		t.Fail()
	}

	if 1 != len(published["2"]) || !utils.SliceContainsString(published["2"], "1") {
		t.Errorf("Second worker has incorrect assigments, got %d devices", len(published["2"]))
		t.Fail()
	}
}
