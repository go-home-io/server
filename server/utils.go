package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/go-home-io/server/providers"
)

// Plain HTTP_200 API response.
func respondOk(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, `{ "status": "OK" }`) // nolint: gosec
}

// Generic API respond.
func respond(writer http.ResponseWriter, data interface{}) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	d, err := json.Marshal(data)
	if err != nil {
		return
	}

	writer.Write(d) // nolint: gosec
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

// Return HTTP_UNAUTH status.
func respondUnAuth(writer http.ResponseWriter) {
	http.Error(writer, "Forbidden", http.StatusUnauthorized)
}

// Plain HTTP_500 API response.
func respondError(writer http.ResponseWriter, err string) {
	writer.WriteHeader(http.StatusInternalServerError)
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, fmt.Sprintf(`{ "status": "ERROR", "problem": "%s"}`, err)) // nolint: gosec
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
		if !strings.HasPrefix(r.RequestURI, routeAPI) {
			if !isRequestInternal(r) {
				respondUnAuth(w)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

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

// Private CIDRs.
var cidrs []*net.IPNet

// Preparing private CIDRs.
func prepareCidrs() {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, err := net.ParseCIDR(maxCidrBlock)
		if err != nil {
			continue
		}
		cidrs[i] = cidr
	}
}

// Checks whether request came from internal subnet.
func isRequestInternal(r *http.Request) bool {
	realIP := r.Header.Get("X-Real-Ip")
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if realIP == "" && forwardedFor == "" {
		var remoteIP string
		var err error
		if strings.ContainsRune(r.RemoteAddr, ':') {
			remoteIP, _, err = net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return false
			}
		} else {
			remoteIP = r.RemoteAddr
		}

		return isAddressPrivate(remoteIP)
	}

	found := false
	for _, address := range strings.Split(forwardedFor, ",") {
		address = strings.TrimSpace(address)
		if !isAddressPrivate(address) {
			return false
		}
		found = true
	}

	if "" == realIP && found {
		return true
	}

	return isAddressPrivate(realIP)
}

// Checks whether IP address belongs to private subnet.
func isAddressPrivate(address string) bool {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			return true
		}
	}

	return false
}
