package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-home-io/server/providers"
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

// Return HTTP_FORBIDDEN status.
func respondUnAuth(writer http.ResponseWriter) {
	http.Error(writer, "Forbidden", http.StatusForbidden)
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

// Authz middleware.
func (s *GoHomeServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.Settings.Security().GetUser(r.Header)
		if err != nil {
			s.Logger.Warn("Unauthorized access attempt", "url", r.RequestURI)
			respondUnAuth(w)

			return
		}

		ctx := context.WithValue(r.Context(), ctxtUserName, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Gets current user out of context.
func getContextUser(request *http.Request) *providers.AuthenticatedUser {
	return request.Context().Value(ctxtUserName).(*providers.AuthenticatedUser)
}
