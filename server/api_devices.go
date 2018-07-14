package server

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// Returns all devices available for the user.
func (s *GoHomeServer) getDevices(writer http.ResponseWriter, request *http.Request) { //nolint: unparam
	respond(writer, s.state.GetAllDevices())
}

// Executes device command if it's allowed for the user.
func (s *GoHomeServer) deviceCommand(writer http.ResponseWriter, request *http.Request) { //nolint: unparam
	vars := mux.Vars(request)
	b, _ := ioutil.ReadAll(request.Body)
	respondOkError(writer, s.commandInvokeDeviceCommand(vars[string(URLDeviceID)], vars[string(URLCommandName)], b))
}
