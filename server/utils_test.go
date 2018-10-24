package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/providers"
)

// Tests log middleware.
func TestLogMiddleware(t *testing.T) {
	in := []struct {
		url string
	}{
		{
			url: "/api/v1/test",
		},
		{
			url: "/pub/ping",
		},
	}

	nextCalled := false
	logCalled := false
	s := &GoHomeServer{
		Logger: mocks.FakeNewLogger(func(s string) {
			logCalled = true
		}),
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nextCalled = true
	})
	ts := httptest.NewServer(s.logMiddleware(handler))

	for _, v := range in {
		nextCalled = false
		logCalled = false
		var u bytes.Buffer
		u.WriteString(string(ts.URL))
		u.WriteString(v.url)

		_, err := http.Get(u.String())
		assert.NoError(t, err, "error %s", v.url)
		assert.True(t, nextCalled, "next %s", v.url)
		assert.True(t, logCalled, "log %s", v.url)
	}
}

// Tests authorization middleware.
func TestAuthMiddleware(t *testing.T) {
	prepareCidrs()
	in := []struct {
		url          string
		nextExpected bool
		security     providers.ISecurityProvider
		headers      map[string]string
	}{
		{
			url:          "/api/v2/test/1",
			security:     mocks.FakeNewSecurityProvider(true),
			headers:      map[string]string{"X-Real-Ip": "512.0.0.0"},
			nextExpected: false,
		},
		{
			url:          "/api/v2/test/2",
			security:     mocks.FakeNewSecurityProvider(true),
			headers:      map[string]string{"X-Forwarded-For": "245.0.0.0"},
			nextExpected: false,
		},
		{
			url:          "/api/v2/test/3",
			security:     mocks.FakeNewSecurityProvider(true),
			headers:      map[string]string{"X-Forwarded-For": "10.0.0.0", "X-Real-IP": "245.0.0.0"},
			nextExpected: false,
		},
		{
			url:          "/api/v2/test/4",
			security:     mocks.FakeNewSecurityProvider(true),
			headers:      map[string]string{"X-Forwarded-For": "10.0.0.1"},
			nextExpected: true,
		},
		{
			url:          "/api/v2/test/5",
			security:     mocks.FakeNewSecurityProvider(true),
			headers:      map[string]string{},
			nextExpected: true,
		},
		{
			url:          "/api/v1/test/6",
			security:     mocks.FakeNewSecurityProvider(true),
			headers:      map[string]string{},
			nextExpected: true,
		},
		{
			url:          "/api/v1/test/7",
			security:     mocks.FakeNewSecurityProvider(false),
			headers:      map[string]string{},
			nextExpected: false,
		},
	}

	nextCalled := false
	s := &GoHomeServer{
		Logger: mocks.FakeNewLogger(nil),
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nextCalled = true
	})

	ts := httptest.NewServer(s.authMiddleware(handler))
	for _, v := range in {
		s.Settings = mocks.FakeNewSettingsWithUserStorage(v.security)

		nextCalled = false
		var u bytes.Buffer
		u.WriteString(string(ts.URL))
		u.WriteString(v.url)

		client := &http.Client{}
		req, _ := http.NewRequest("GET", u.String(), nil)

		for k, h := range v.headers {
			req.Header.Add(k, h)
		}

		_, err := client.Do(req)
		assert.NoError(t, err, "error %s", v.url)
		assert.Equal(t, v.nextExpected, nextCalled, "call %s", v.url)
	}
}
