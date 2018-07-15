package server

import (
	"testing"
	"github.com/go-home-io/server/providers"
	"regexp"
)

func compileRegexp(r string) *regexp.Regexp {
	reg, _ := regexp.Compile(r)
	return reg
}

// Tests whether baked rules are interpreted correctly to allow operations.
func TestAllowed(t *testing.T) {
	devices := []*knownDevice{
		{ID: "device.12"},
		{ID: "device-2"},
		{ID: "device_123"},
		{ID: "static device"},
		{ID: "hub.127.0.0.1"},
	}

	user := &providers.AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					History: true,
					Command: true,
					Resources: [] *regexp.Regexp{
						compileRegexp("static device"),
						compileRegexp("dev\\S"),
					},
				},
			},
			providers.SecSystemAll: {
				{
					Get:     true,
					History: true,
					Command: true,
					Resources: [] *regexp.Regexp{
						compileRegexp("hub\\S"),
					},
				},
			},
		},
	}

	for _, v := range devices {
		if !v.Get(user) {
			t.Log("Failed Get on " + v.ID)
			t.Fail()
		}

		if !v.Command(user) {
			t.Log("Failed command on " + v.ID)
			t.Fail()
		}

		if !v.History(user) {
			t.Log("Failed history on " + v.ID)
			t.Fail()
		}
	}
}

// Tests whether baked rules are interpreted correctly to forbid operations.
func TestForbidden(t *testing.T) {
	devices := []*knownDevice{
		{ID: "device.12"},
		{ID: "device-2"},
		{ID: "device_123"},
		{ID: "static device"},
		{ID: "hub.127.0.0.1"},
	}

	user := &providers.AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					History: false,
					Command: true,
					Resources: [] *regexp.Regexp{
						compileRegexp("static[a-zA-Z]"),
						compileRegexp("dev\\s"),
					},
				},
			},
			providers.SecSystemAll: {
				{
					Get:     false,
					History: true,
					Command: false,
					Resources: [] *regexp.Regexp{
						compileRegexp("hub\\S"),
					},
				},
			},
		},
	}

	for _, v := range devices {
		if v.Get(user) {
			t.Log("Failed Get on " + v.ID)
			t.Fail()
		}

		if v.Command(user) {
			t.Log("Failed command on " + v.ID)
			t.Fail()
		}

		if v.History(user) {
			t.Log("Failed history on " + v.ID)
			t.Fail()
		}
	}
}

// Tests whether baked rules and verbs are interpreted correctly to allow operations.
func TestInternalAllowed(t *testing.T) {
	devices := []*knownDevice{
		{ID: "device.12"},
		{ID: "device-2"},
		{ID: "device_123"},
		{ID: "static device"},
		{ID: "hub.127.0.0.1"},
	}

	rules := []*providers.BakedRule{
		{
			Get:     false,
			History: true,
			Command: false,
			Resources: [] *regexp.Regexp{
				compileRegexp("static device"),
				compileRegexp("dev\\S"),
			},
		},
		{
			Get:     true,
			History: false,
			Command: true,
			Resources: [] *regexp.Regexp{
				compileRegexp("hub\\S"),
			},
		},
	}

	for _, v := range devices {
		if !v.isAllowed(rules, providers.SecVerbAll) {
			t.Log("Failed on " + v.ID)
			t.Fail()
		}
	}
}

// Tests whether baked rules and verbs are interpreted correctly to forbid operations.
func TestInternalForbidden(t *testing.T) {
	devices := []*knownDevice{
		{ID: "device.12"},
		{ID: "device-2"},
		{ID: "device_123"},
		{ID: "static device"},
		{ID: "hub.127.0.0.1"},
	}

	rules := []*providers.BakedRule{
		{
			Get:     false,
			History: false,
			Command: false,
			Resources: [] *regexp.Regexp{
				compileRegexp("static device"),
				compileRegexp("dev\\S"),
			},
		},
		{
			Get:     false,
			History: false,
			Command: false,
			Resources: [] *regexp.Regexp{
				compileRegexp("hub\\S"),
			},
		},
	}

	for _, v := range devices {
		if v.isAllowed(rules, providers.SecVerbAll) {
			t.Log("Failed on " + v.ID)
			t.Fail()
		}
	}
}
