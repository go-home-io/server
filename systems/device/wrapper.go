package device

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/utils"
)

// IDeviceWrapperProvider interface for any loaded devices.
type IDeviceWrapperProvider interface {
	providers.ILoadedProvider
	ID() string
	Name() string
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
	LoadData         *device.InitDataDevice
	IsRootDevice     bool
	Validator        providers.IValidatorProvider
	UOM              enums.UOM
	processor        IProcessor
	RawConfig        string

	StatusUpdatesChan chan *UpdateEvent
	DiscoveryChan     chan *NewDeviceDiscoveredEvent
}

// Device wrapper implementation.
type deviceWrapper struct {
	sync.Mutex

	Ctor *wrapperConstruct

	internalID  string
	name        string
	State       map[string]interface{}
	Spec        *device.Spec
	CommandsStr []string

	stopped      bool
	stopChan     chan bool
	updateMethod reflect.Value
	commands     map[enums.Command]reflect.Value
	children     []IDeviceWrapperProvider

	isPolling bool
	processor IProcessor
}

// NewDeviceWrapper constructs a new device wrapper.
func NewDeviceWrapper(ctor *wrapperConstruct) IDeviceWrapperProvider {
	w := deviceWrapper{
		Ctor:      ctor,
		isPolling: false,
		State:     make(map[string]interface{}),
		processor: ctor.processor,
		stopChan:  make(chan bool, 5),
		stopped:   false,
		children:  make([]IDeviceWrapperProvider, 0),
	}

	w.Spec = ctor.DeviceInterface.(device.IDevice).GetSpec()
	if nil == w.Spec {
		w.Spec = &device.Spec{
			SupportedProperties: make([]enums.Property, 0),
			SupportedCommands:   make([]enums.Command, 0),
			UpdatePeriod:        0,
		}
	}

	if nil != w.processor {
		w.Spec.SupportedProperties = append(w.Spec.SupportedProperties, w.processor.GetExtraSupportPropertiesSpec()...)
	}

	if !w.setState(ctor.DeviceState) {
		ctor.Logger.Warn("Failed to fetch device state",
			common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
	}

	if w.Spec.UpdatePeriod.Seconds() > 0 {
		w.isPolling = true
		go w.periodicUpdates(w.Spec.UpdatePeriod)
		w.updateMethod = reflect.ValueOf(ctor.DeviceInterface).MethodByName("Update")
		ctor.Logger.Debug(fmt.Sprintf("Polling rate for the device is %d seconds",
			int(w.Spec.UpdatePeriod.Seconds())), common.LogDeviceTypeToken,
			ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
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
func (w *deviceWrapper) ID() string {
	if w.internalID == "" {
		w.internalID = fmt.Sprintf("%s.%s.%s", utils.NormalizeDeviceName(w.Ctor.DeviceConfigName),
			utils.NormalizeDeviceName(w.Ctor.DeviceType.String()),
			utils.NormalizeDeviceName(w.Ctor.DeviceInterface.(device.IDevice).GetName()))
	}
	return w.internalID
}

// Name returns device name.
func (w *deviceWrapper) Name() string {
	if w.name == "" {
		if w.Ctor.IsRootDevice && "" != w.Ctor.DeviceConfigName {
			w.name = w.Ctor.DeviceConfigName
		} else {
			w.name = helpers.GetNameFromID(w.ID())
		}
	}

	return w.name
}

// Unload stops all background activities.
func (w *deviceWrapper) Unload() {
	w.Lock()
	defer w.Unlock()

	if w.stopped {
		return
	}

	// Stopping all three threads.
	w.stopChan <- true
	w.stopChan <- true
	w.stopChan <- true

	w.Ctor.DeviceInterface.(device.IDevice).Unload()
	close(w.Ctor.LoadData.DeviceStateUpdateChan)
	if w.Ctor.IsRootDevice {
		close(w.Ctor.LoadData.DeviceDiscoveredChan)
	}

	w.stopped = true

	for _, v := range w.children {
		v.Unload()
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
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
			common.LogDeviceCommandToken, cmdName.String())
		return
	}

	w.Ctor.Logger.Debug("Invoking device command",
		common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
		common.LogDeviceCommandToken, cmdName.String())

	var results []reflect.Value

	if method.Type().NumIn() > 0 {
		obj, err := json.Marshal(param)
		if err != nil {
			w.Ctor.Logger.Error("Got error while marshalling data for device command", err,
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
				common.LogDeviceCommandToken, cmdName.String())
			return
		}

		objNew := reflect.New(method.Type().In(0)).Interface()
		val := reflect.ValueOf(objNew)

		err = json.Unmarshal(obj, &objNew)
		if err != nil {
			w.Ctor.Logger.Error("Got error while preparing data for device command", err,
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
				common.LogDeviceCommandToken, cmdName.String())
			return
		}

		if !w.Ctor.Validator.Validate(objNew) {
			w.Ctor.Logger.Warn("Received incorrect command params",
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
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
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
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
	msg.DeviceID = w.ID()
	msg.State = w.State
	msg.WorkerID = w.Ctor.WorkerID
	msg.Commands = w.CommandsStr
	msg.DeviceName = w.Name()
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
				common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
				common.LogDeviceCommandToken, v.String())
			continue
		}

		method := reflect.ValueOf(w.Ctor.DeviceInterface).MethodByName(v.GetCommandMethodName())
		if !method.IsValid() {
			ctor.Logger.Warn("Plugin claimed non-implemented command",
				common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
				common.LogDeviceCommandToken, v.String())
			continue
		}

		if method.Type().NumIn() > 1 {
			ctor.Logger.Warn("Plugin declared method with more than one param",
				common.LogDeviceTypeToken, ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID(),
				common.LogDeviceCommandToken, v.String())
			continue
		}

		w.commands[v] = method
		w.CommandsStr = append(w.CommandsStr, v.String())
	}
}

// Updates internal device state which is stored in wrapper.
func (w *deviceWrapper) setState(deviceState interface{}) bool {
	if nil == deviceState ||
		reflect.ValueOf(deviceState).Kind() == reflect.Ptr && reflect.ValueOf(deviceState).IsNil() {
		return false
	}

	allowedProps, ok := enums.AllowedProperties[w.Ctor.DeviceType]
	if !ok {
		w.Ctor.Logger.Warn("Received unknown device type",
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
		return false
	}

	rt, rv := reflect.TypeOf(deviceState), reflect.ValueOf(deviceState)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	for ii := 0; ii < rt.NumField(); ii++ {
		field := rt.Field(ii)
		jsonKey := field.Tag.Get("json")
		if "" == jsonKey {
			continue
		}

		prop, err := enums.PropertyString(jsonKey)
		if err != nil {
			w.Ctor.Logger.Warn("Received unknown device property", common.LogDevicePropertyToken, jsonKey,
				common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
			continue
		}

		if !enums.SliceContainsProperty(w.Spec.SupportedProperties, prop) ||
			!enums.SliceContainsProperty(allowedProps, prop) {
			continue
		}

		val := w.getFieldValueOrNil(rv.Field(ii))
		if val == nil {
			continue
		}

		w.preProcessProperty(jsonKey, prop, val)
	}

	return true
}

// Processes property.
func (w *deviceWrapper) preProcessProperty(jsonKey string, property enums.Property, value interface{}) {
	if nil == w.processor {
		w.State[jsonKey] = value
		return
	}

	if w.processor.IsExtraProperty(property) {
		return
	}

	ok, props := w.processor.IsPropertyGood(property, value)

	if !ok {
		delete(w.State, jsonKey)
		return
	}

	for k, v := range props {
		w.State[k.String()] = v
	}
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

// Calls for updates.
func (w *deviceWrapper) periodicUpdates(duration time.Duration) {
	if !w.isPolling {
		return
	}

	tick := time.Tick(duration)
	for {
		select {
		case <-w.stopChan:
			return
		case <-tick:
			go w.pullUpdate()
		}
	}
}

// Performs data pull from device provider plugin.
func (w *deviceWrapper) pullUpdate() {
	if !w.isPolling {
		return
	}

	w.Ctor.Logger.Debug("Fetching update for the device", common.LogDeviceTypeToken,
		w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
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
			w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
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
		err = &ErrNoDataFromPlugin{}
	}

	if len(state) > 1 && nil != state[1].Interface() {
		err = state[1].Interface().(error)
	}

	if len(state) > 1 && nil == state[0].Interface() && nil == state[1].Interface() {
		err = &ErrNoDataFromPlugin{}
	}

	if err != nil {
		w.Ctor.Logger.Error("Failed to fetch device updates", err,
			common.LogDeviceTypeToken, w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
	} else {
		w.processUpdate(state[0].Interface())
	}
}

// Starts listeners of incoming updates/discovery messages.
func (w *deviceWrapper) startHubListeners() {
	for {
		select {
		case <-w.stopChan:
			return
		case discovery, ok := <-w.Ctor.LoadData.DeviceDiscoveredChan:
			if !ok {
				return
			}
			w.Ctor.Logger.Debug("Received discovery callback for the device", common.LogDeviceTypeToken,
				w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())

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
	for {
		select {
		case <-w.stopChan:
			return
		case update, ok := <-w.Ctor.LoadData.DeviceStateUpdateChan:
			if !ok {
				return
			}
			w.processUpdate(update.State)
		}
	}
}

// Processing update message from provider plugin.
func (w *deviceWrapper) processUpdate(state interface{}) {
	w.Ctor.Logger.Debug("Received update for the device", common.LogDeviceTypeToken,
		w.Ctor.DeviceType.String(), common.LogDeviceNameToken, w.ID())
	w.setState(state)
	w.Ctor.StatusUpdatesChan <- &UpdateEvent{
		ID: w.ID(),
	}
}

// Processing discovery message from hub provider plugin.
func (w *deviceWrapper) processDiscovery(d *device.DiscoveredDevices) {
	w.Ctor.Logger.Info("Discovered a new device", common.LogDeviceTypeToken, d.Type.String())

	subLoadData := &device.InitDataDevice{
		UOM:                   w.Ctor.UOM,
		Logger:                w.Ctor.Logger,
		Secret:                w.Ctor.Secret,
		DeviceDiscoveredChan:  w.Ctor.LoadData.DeviceDiscoveredChan,
		DeviceStateUpdateChan: make(chan *device.StateUpdateData, 10),
	}

	loadedDevice, ok := d.Interface.(device.IDevice)
	if !ok {
		w.Ctor.Logger.Warn("One of the loaded devices is not implementing IDevice interface",
			common.LogDeviceTypeToken, w.Ctor.DeviceConfigName, common.LogDeviceNameToken, w.ID())
		return
	}

	err := loadedDevice.Init(subLoadData)
	if err != nil {
		w.Ctor.Logger.Warn("Failed to execute device.Load method",
			common.LogDeviceTypeToken, w.Ctor.DeviceConfigName, common.LogDeviceNameToken, w.ID())
		return
	}

	ctor := &wrapperConstruct{
		DeviceType:        d.Type,
		DeviceInterface:   d.Interface,
		IsRootDevice:      false,
		DeviceConfigName:  w.Ctor.DeviceConfigName,
		DeviceState:       d.State,
		LoadData:          subLoadData,
		Logger:            w.Ctor.Logger,
		Secret:            w.Ctor.Secret,
		WorkerID:          w.Ctor.WorkerID,
		DiscoveryChan:     w.Ctor.DiscoveryChan,
		StatusUpdatesChan: w.Ctor.StatusUpdatesChan,
		UOM:               w.Ctor.UOM,
		Validator:         w.Ctor.Validator,
		processor:         newDeviceProcessor(d.Type, w.Ctor.RawConfig),
		RawConfig:         w.Ctor.RawConfig,
	}

	wrapper := NewDeviceWrapper(ctor)

	w.Ctor.DiscoveryChan <- &NewDeviceDiscoveredEvent{
		Provider: wrapper,
	}

	subLoadData.DeviceStateUpdateChan <- &device.StateUpdateData{
		State: d.State,
	}

	w.children = append(w.children, wrapper)
}
