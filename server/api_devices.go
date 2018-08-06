package server

import (
	"io/ioutil"
	"net/http"

	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/gorilla/mux"
)

// Contains data about known groups.
type knownGroup struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Devices []string `json:"devices"`
}

// Returns all devices available for the user.
func (s *GoHomeServer) getDevices(writer http.ResponseWriter, request *http.Request) {
	respond(writer, s.commandGetAllDevices(getContextUser(request)))
}

// Returns all groups available for the user.
func (s *GoHomeServer) getGroups(writer http.ResponseWriter, request *http.Request) {
	devices := s.commandGetAllDevices(getContextUser(request))
	response := make([]*knownGroup, 0)
	for _, v := range devices {
		if v.Type != enums.DevGroup {
			continue
		}

		g, ok := s.groups[v.ID]
		if !ok {
			continue
		}

		group := &knownGroup{
			ID:      v.ID,
			Name:    v.Name,
			Devices: g.Devices(),
		}

		response = append(response, group)
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
