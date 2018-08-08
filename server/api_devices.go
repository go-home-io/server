package server

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type knownLocation struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"`
}

// Contains data about known groups.
type knownGroup struct {
	knownLocation
	ID string `json:"id"`
}

// Contains server state required UI to start.
type currentState struct {
	Devices   []*knownDevice   `json:"devices"`
	Groups    []*knownGroup    `json:"groups"`
	Locations []*knownLocation `json:"locations"`
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
		Locations: make([]*knownLocation, 0),
	}

	for _, v := range s.locations {
		l := &knownLocation{
			Devices: v.Devices(),
			Name:    v.ID(),
		}

		response.Locations = append(response.Locations, l)
	}

	respond(writer, response)
}

// Executes device command if it's allowed for the user.
func (s *GoHomeServer) deviceCommand(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	b, _ := ioutil.ReadAll(request.Body)
	respondOkError(writer, s.commandInvokeDeviceCommand(getContextUser(request),
		vars[string(urlDeviceID)], vars[string(urlCommandName)], b))
}
