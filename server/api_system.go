package server

import "net/http"

// Performs quick check whether system is OK.
func (s *GoHomeServer) ping(writer http.ResponseWriter, _ *http.Request) {
	if s.Settings.ServiceBus().Ping() != nil {
		respondError(writer, "Service bus unavailable")
		return
	}
	respondOk(writer)
}
