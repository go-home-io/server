package server

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"go-home.io/x/server/plugins/device/enums"
)

// Contains data about known locations.
type knownLocation struct {
	Name    string   `json:"name"`
	Icon    string   `json:"icon"`
	Devices []string `json:"devices"`
}

// Contains data about known groups.
type knownGroup struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"`
	ID      string   `json:"id"`
}

// Contains server state required UI to start.
type currentState struct {
	Devices   []*knownDevice   `json:"devices"`
	Groups    []*knownGroup    `json:"groups"`
	Locations []*knownLocation `json:"locations"`
	UOM       enums.UOM        `json:"uom"`
}

// Known devices, received from workers.
type knownDevice struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Worker     string                 `json:"worker"`
	Type       enums.DeviceType       `json:"type"`
	State      map[string]interface{} `json:"state"`
	LastSeen   int64                  `json:"last_seen"`
	Commands   []string               `json:"commands"`
	IsReadOnly bool                   `json:"read_only"`
}

// Returns all devices available for the user.
func (s *GoHomeServer) getDevices(writer http.ResponseWriter, request *http.Request) {
	respond(writer, s.commandGetAllDevices(getContextUser(request)))
}

// Returns all groups available for the user.
func (s *GoHomeServer) getGroups(writer http.ResponseWriter, request *http.Request) {
	respond(writer, s.commandGetAllGroups(getContextUser(request)))
}

// Returns server state required for UI to start.
func (s *GoHomeServer) getCurrentState(writer http.ResponseWriter, request *http.Request) {
	usr := getContextUser(request)
	response := &currentState{
		Devices:   s.commandGetAllDevices(usr),
		Groups:    s.commandGetAllGroups(usr),
		Locations: s.commandGetAllLocations(usr),
		UOM:       s.Settings.MasterSettings().UOM,
	}

	respond(writer, response)
}

// Executes device command if it's allowed for the user.
func (s *GoHomeServer) deviceCommand(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		respondError(writer, "Failed to read body")
	}
	respondOkError(writer, s.commandInvokeDeviceCommand(getContextUser(request),
		vars[string(urlDeviceID)], vars[string(urlCommandName)], b))
}

// Gets device state history.
func (s *GoHomeServer) getStateHistory(writer http.ResponseWriter, request *http.Request) {
	user := getContextUser(request)
	vars := mux.Vars(request)
	kd := s.state.GetDevice(vars[string(urlDeviceID)])

	if !user.DeviceHistory(kd.ID) {
		respondForbidden(writer)
		return
	}

	respond(writer, s.Settings.Storage().History(kd.ID))
}
