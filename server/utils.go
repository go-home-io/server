package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Plain HTTP_200 API response.
func respondOk(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, `{ "status": "OK" }`)
}

// Generic API respond.
func respond(writer http.ResponseWriter, data interface{}) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	d, err := json.Marshal(data)
	if err != nil {
		return
	}

	writer.Write(d)
}

// Validates whether error is not null and responds different status
// depending on it.
func respondOkError(writer http.ResponseWriter, err error) {
	if err != nil {
		respondError(writer, err.Error())
	} else {
		respondOk(writer)
	}
}

// Plain HTTP_500 API response.
func respondError(writer http.ResponseWriter, err string) {
	writer.WriteHeader(http.StatusInternalServerError)
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, fmt.Sprintf(`{ "status": "ERROR", "problem": "%s"}`, err))
}

// Logger middleware for the API.
func (s *GoHomeServer) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Debug("REST invocation", "url", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
