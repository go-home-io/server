package server

import (
	"testing"
	"github.com/go-home-io/server/providers"
	"github.com/gobwas/glob"
)

func compileRegexp(r string) glob.Glob {
	reg, _ := glob.Compile(r)
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
					Resources: [] glob.Glob{
						compileRegexp("static device"),
						compileRegexp("dev*"),
					},
				},
			},
			providers.SecSystemAll: {
				{
					Get:     true,
					History: true,
					Command: true,
					Resources: [] glob.Glob{
						compileRegexp("hub*"),
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
					Resources: [] glob.Glob{
						compileRegexp("static[! ]*"),
						compileRegexp("dev?"),
					},
				},
			},
			providers.SecSystemAll: {
				{
					Get:     false,
					History: true,
					Command: false,
					Resources: [] glob.Glob{
						compileRegexp("hub*"),
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
			Resources: [] glob.Glob{
				compileRegexp("static device"),
				compileRegexp("dev*"),
			},
		},
		{
			Get:     true,
			History: false,
			Command: true,
			Resources: [] glob.Glob{
				compileRegexp("hub*"),
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
			Resources: [] glob.Glob{
				compileRegexp("static device"),
				compileRegexp("dev\\S"),
			},
		},
		{
			Get:     false,
			History: false,
			Command: false,
			Resources: [] glob.Glob{
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
