package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/utils"
)

// IDeviceWrapperProvider interface for any loaded devices.
type IDeviceWrapperProvider interface {
	GetID() string
	Unload()
	InvokeCommand(enums.Command, map[string]interface{})
	GetUpdateMessage() *bus.DeviceUpdateMessage
}

// UpdateEvent is a type used for updates sent by a device.
type UpdateEvent struct {
	ID string
}

// NewDeviceDiscoveredEvent is a type used for discovering a new device.
type NewDeviceDiscoveredEvent struct {
	Provider IDeviceWrapperProvider
}

// Data required for a new wrapper.
type wrapperConstruct struct {
	DeviceType       enums.DeviceType
	DeviceConfigName string
	DeviceInterface  interface{}
	DeviceState      interface{}
	Logger           common.ILoggerProvider
	Secret           common.ISecretProvider
	WorkerID         string
	Cron             providers.ICronProvider
	LoadData         *device.InitDataDevice
	IsRootDevice     bool
	Validator        providers.IValidatorProvider

	StatusUpdatesChan chan *UpdateEvent
	DiscoveryChan     chan *NewDeviceDiscoveredEvent
}

// Device wrapper implementation.
type deviceWrapper struct {
	sync.Mutex

	Ctor *wrapperConstruct

	ID          string
	State       map[string]interface{}
	Spec        *device.Spec
	CommandsStr []string

	jobID        int
	updateMethod reflect.Value
	commands     map[enums.Command]reflect.Value

	isPolling bool
}

// NewDeviceWrapper constructs a new device wrapper.
func NewDeviceWrapper(ctor *wrapperConstruct) IDeviceWrapperProvider {
	w := deviceWrapper{
		Ctor:      ctor,
		isPolling: false,
	}

	w.Spec = ctor.DeviceInterface.(device.IDevice).GetSpec()
	if nil == w.Spec {
		w.Spec = &device.Spec{
			SupportedProperties: make([]enums.Property, 0),
			SupportedCommands:   make([]enums.Command, 0),
			UpdatePeriod:        0,
		}
	}

	if !w.setState(ctor.DeviceState) {
		ctor.Logger.Warn("Failed to fetch device state",
			common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
	}

	interval := int(w.Spec.UpdatePeriod / time.Second)
	if interval > 0 {
		w.isPolling = true
		if interval < 10 {
			interval = 10
		}

		w.updateMethod = reflect.ValueOf(ctor.DeviceInterface).MethodByName("Update")
		var err error
		w.jobID, err = ctor.Cron.AddFunc(fmt.Sprintf("@every %ds", interval), w.pullUpdate)
		if err != nil {
			ctor.Logger.Warn("Failed to schedule device updates",
				common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
		}

		ctor.Logger.Debug(fmt.Sprintf("Polling rate for the device is %d seconds", interval),
			common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
	}

	w.validateDeviceSpec(ctor)

	if ctor.IsRootDevice {
		go w.startHubListeners()
	} else {
		go w.startDeviceListeners()
	}

	return &w
}

// GetID returns unique device ID.
// ID is normalized and contains config name, provider name and ID returned from actual device.
func (w *deviceWrapper) GetID() string {
	if w.ID == "" {
		w.ID = fmt.Sprintf("%s.%s.%s", utils.NormalizeDeviceName(w.Ctor.DeviceConfigName),
			utils.NormalizeDeviceName(w.Ctor.DeviceType.String()),
			utils.NormalizeDeviceName(w.Ctor.DeviceInterface.(device.IDevice).GetName()))
	}
	return w.ID
}

// Unload stops all background activities.
func (w *deviceWrapper) Unload() {
	w.Ctor.DeviceInterface.(device.IDevice).Unload()
	if 0 != w.jobID {
		w.Ctor.Cron.RemoveFunc(w.jobID)
	}

	close(w.Ctor.LoadData.DeviceStateUpdateChan)

	if w.Ctor.IsRootDevice {
		close(w.Ctor.LoadData.DeviceDiscoveredChan)
	}
}

// InvokeCommand performs a call to the device provider.
// This method validates whether device actually reported this operation as supported.
func (w *deviceWrapper) InvokeCommand(cmdName enums.Command, param map[string]interface{}) {
	w.Lock()
	defer w.Unlock()

	method, ok := w.commands[cmdName]
	if !ok {
		w.Ctor.Logger.Warn("Device doesn't support this command",
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
			common.LogDeviceCommandToken, cmdName.String())
		return
	}

	w.Ctor.Logger.Debug("Invoking device command",
		common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
		common.LogDeviceCommandToken, cmdName.String())

	var results []reflect.Value

	if method.Type().NumIn() > 0 {
		obj, err := json.Marshal(param)
		if err != nil {
			w.Ctor.Logger.Error("Got error while marshalling data for device command", err,
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
				common.LogDeviceCommandToken, cmdName.String())
			return
		}

		objNew := reflect.New(method.Type().In(0)).Interface()
		val := reflect.ValueOf(objNew)

		err = json.Unmarshal(obj, &objNew)
		if err != nil {
			w.Ctor.Logger.Error("Got error while preparing data for device command", err,
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
				common.LogDeviceCommandToken, cmdName.String())
			return
		}

		if !w.Ctor.Validator.Validate(objNew) {
			w.Ctor.Logger.Warn("Received incorrect command params",
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
				common.LogDeviceCommandToken, cmdName.String())
			return
		}
		if reflect.ValueOf(objNew).Kind() != method.Type().In(0).Kind() {
			val = val.Elem()
		}

		results = method.Call([]reflect.Value{val})
	} else {
		results = method.Call(nil)
	}

	if len(results) > 0 && results[0].Interface() != nil {
		w.Ctor.Logger.Error("Got error while invoking device command", results[0].Interface().(error),
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
			common.LogDeviceCommandToken, cmdName.String())

		return
	}
	if w.Spec.PostCommandDeferUpdate > 0 {
		time.Sleep(w.Spec.PostCommandDeferUpdate)
	}

	w.pullUpdate()
}

// GetUpdateMessage constructs device update message.
func (w *deviceWrapper) GetUpdateMessage() *bus.DeviceUpdateMessage {
	msg := bus.NewDeviceUpdateMessage()
	msg.DeviceType = w.Ctor.DeviceType
	msg.DeviceID = w.ID
	msg.State = w.State
	msg.WorkerID = w.Ctor.WorkerID
	msg.Commands = w.CommandsStr
	return msg
}

// Validates specification, returned by device and prepares
// supported commands.
func (w *deviceWrapper) validateDeviceSpec(ctor *wrapperConstruct) {
	w.CommandsStr = make([]string, 0)
	w.commands = make(map[enums.Command]reflect.Value)
	for _, v := range w.Spec.SupportedCommands {
		if !v.IsCommandAllowed(ctor.DeviceType) {
			ctor.Logger.Warn("Plugin claimed restricted command",
				common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
				common.LogDeviceCommandToken, v.String())
			continue
		}

		method := reflect.ValueOf(w.Ctor.DeviceInterface).MethodByName(v.GetCommandMethodName())
		if !method.IsValid() {
			ctor.Logger.Warn("Plugin claimed non-implemented command",
				common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
				common.LogDeviceCommandToken, v.String())
			continue
		}

		if method.Type().NumIn() > 1 {
			ctor.Logger.Warn("Plugin declared method with more than one param",
				common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID,
				common.LogDeviceCommandToken, v.String())
			continue
		}

		w.commands[v] = method
		w.CommandsStr = append(w.CommandsStr, v.String())
	}
}

// Updates internal device state which is stored in wrapper.
func (w *deviceWrapper) setState(deviceState interface{}) bool {
	if nil == deviceState || reflect.ValueOf(deviceState).Kind() == reflect.Ptr && reflect.ValueOf(deviceState).IsNil() {
		return false
	}

	allowedProps, ok := enums.AllowedProperties[w.Ctor.DeviceType]
	if !ok {
		w.Ctor.Logger.Warn("Received unknown device type",
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
		return false
	}

	rt, rv := reflect.TypeOf(deviceState), reflect.ValueOf(deviceState)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	w.State = make(map[string]interface{}, rt.NumField())
	for ii := 0; ii < rt.NumField(); ii++ {
		field := rt.Field(ii)
		jsonKey := field.Tag.Get("json")
		if "" == jsonKey {
			continue
		}

		prop, err := enums.PropertyString(jsonKey)
		if err != nil {
			w.Ctor.Logger.Warn("Received unknown device property", common.LogDevicePropertyToken, jsonKey,
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
			continue
		}

		if !enums.SliceContainsProperty(w.Spec.SupportedProperties, prop) {
			continue
		}

		if !enums.SliceContainsProperty(allowedProps, prop) {
			continue
		}

		val := w.getFieldValueOrNil(rv.Field(ii))
		if val != nil {
			w.State[jsonKey] = val
		}
	}

	return true
}

// Returns actual value or nil.
func (w *deviceWrapper) getFieldValueOrNil(valField reflect.Value) interface{} {
	val := valField.Interface()
	switch valField.Kind() {
	case reflect.Slice, reflect.Chan, reflect.Map, reflect.Array, reflect.String:
		if 0 == valField.Len() {
			return nil
		}
	default:
		if nil == val {
			return nil
		}
	}

	return val
}

// Performs data pull from device provider plugin.
func (w *deviceWrapper) pullUpdate() {
	if !w.isPolling {
		return
	}

	w.Ctor.Logger.Debug("Fetching update for the device", common.LogDeviceTypeToken,
		w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
	switch w.Ctor.DeviceType {
	case enums.DevHub:
		w.pullHubUpdate()
	default:
		w.pullDeviceUpdate()
	}
}

// Performs data pull from hub.
// Hub could have discovered new devices.
func (w *deviceWrapper) pullHubUpdate() {
	hubState, err := w.Ctor.DeviceInterface.(device.IHub).Update()
	if err != nil {
		w.Ctor.Logger.Error("Failed to fetch hub updates", err, common.LogDeviceTypeToken,
			w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
		return
	}
	w.processUpdate(hubState)

	for _, d := range hubState.Devices {
		w.processDiscovery(d)
	}
}

// Performs data pull from device.
func (w *deviceWrapper) pullDeviceUpdate() {
	state := w.updateMethod.Call(nil)
	var err error
	if 0 == len(state) {
		err = errors.New("plugin didn't return any data")
	}

	if len(state) > 1 && nil != state[1].Interface() {
		err = state[1].Interface().(error)
	}

	if len(state) > 1 && nil == state[0].Interface() && nil == state[1].Interface() {
		err = errors.New("plugin didn't return any data")
	}

	if err != nil {
		w.Ctor.Logger.Error("Failed to fetch device updates", err,
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)
	} else {
		w.processUpdate(state[0].Interface())
	}
}

// Starts listeners of incoming updates/discovery messages.
func (w *deviceWrapper) startHubListeners() {
	for {
		select {
		case discovery, ok := <-w.Ctor.LoadData.DeviceDiscoveredChan:
			if !ok {
				return
			}
			w.Ctor.Logger.Debug("Received discovery callback for the device", common.LogDeviceTypeToken,
				w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)

			w.processDiscovery(discovery)
		case update, ok := <-w.Ctor.LoadData.DeviceStateUpdateChan:
			if !ok {
				return
			}
			w.processUpdate(update.State)
		}
	}
}

// Starts listeners of incoming messages.
// Only hub listens for discovery.
func (w *deviceWrapper) startDeviceListeners() {
	for update := range w.Ctor.LoadData.DeviceStateUpdateChan {
		w.processUpdate(update.State)
	}
}

// Processing update message from provider plugin.
func (w *deviceWrapper) processUpdate(state interface{}) {
	w.Ctor.Logger.Debug("Received update for the device", common.LogDeviceTypeToken,
		w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)

	w.setState(state)

	w.Ctor.StatusUpdatesChan <- &UpdateEvent{
		ID: w.ID,
	}
}

// Processing discovery message from hub provider plugin.
func (w *deviceWrapper) processDiscovery(d *device.DiscoveredDevices) {
	w.Ctor.Logger.Info("Discovered a new device", common.LogDeviceTypeToken,
		w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID)

	subLoadData := &device.InitDataDevice{
		Logger:                w.Ctor.Logger,
		Secret:                w.Ctor.Secret,
		DeviceDiscoveredChan:  w.Ctor.LoadData.DeviceDiscoveredChan,
		DeviceStateUpdateChan: make(chan *device.StateUpdateData, 10),
	}

	loadedDevice, ok := d.Interface.(device.IDevice)
	if !ok {
		w.Ctor.Logger.Warn("One of the loaded devices is not implementing IDevice interface",
			common.LogDeviceTypeToken, w.Ctor.DeviceConfigName, common.LogDeviceNameToken, w.ID)
		return
	}

	err := loadedDevice.Init(subLoadData)
	if err != nil {
		w.Ctor.Logger.Warn("Failed to execute device.Load method",
			common.LogDeviceTypeToken, w.Ctor.DeviceConfigName, common.LogDeviceNameToken, w.ID)
		return
	}

	ctor := &wrapperConstruct{
		DeviceType:        d.Type,
		DeviceInterface:   d.Interface,
		IsRootDevice:      false,
		Cron:              w.Ctor.Cron,
		DeviceConfigName:  w.Ctor.DeviceConfigName,
		DeviceState:       d.State,
		LoadData:          w.Ctor.LoadData,
		Logger:            w.Ctor.Logger,
		Secret:            w.Ctor.Secret,
		WorkerID:          w.Ctor.WorkerID,
		DiscoveryChan:     w.Ctor.DiscoveryChan,
		StatusUpdatesChan: w.Ctor.StatusUpdatesChan,
	}

	wrapper := NewDeviceWrapper(ctor)

	w.Ctor.DiscoveryChan <- &NewDeviceDiscoveredEvent{
		Provider: wrapper,
	}
}
