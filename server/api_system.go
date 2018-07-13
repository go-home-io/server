package server

import "net/http"

func (s *GoHomeServer) ping(writer http.ResponseWriter, request *http.Request) { //nolint: unparam
	if s.Settings.ServiceBus().Ping() != nil {
		respondError(writer, "Service bus unavailable")
		return
	}
	respondOk(writer)
}
