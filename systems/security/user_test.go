package security

import (
	"testing"

	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/providers"
)

func compileRegexp(r string) glob.Glob {
	reg, _ := glob.Compile(r)
	return reg
}

func checkAllAllowed(t *testing.T, user providers.IAuthenticatedUser) {
	assert.True(t, user.DeviceGet("device1"), "get %s", "get")
	assert.True(t, user.DeviceCommand("dev2"), "command %s", "command")
	assert.True(t, user.DeviceHistory("другой_девайс"), "history %s", "history")
	assert.True(t, user.Workers(), "workers")
	assert.True(t, user.Entities(), "entities")
	assert.True(t, user.Entities(), "logs")
}

// Tests whether baked rules are interpreted correctly to allow operations.
func TestAllowed(t *testing.T) {
	devices := []string{"device.12",
		"device-2",
		"device_123",
		"static device",
		"hub.127.0.0.1",
	}

	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					History: true,
					Command: true,
					Resources: []glob.Glob{
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
					Resources: []glob.Glob{
						compileRegexp("hub*"),
					},
				},
			},
		},
	}

	for _, v := range devices {
		assert.True(t, user.DeviceGet(v), "get %s", v)
		assert.True(t, user.DeviceCommand(v), "command %s", v)
		assert.True(t, user.DeviceHistory(v), "history %s", v)
	}
}

// Tests whether baked rules are interpreted correctly to forbid operations.
func TestForbidden(t *testing.T) {
	devices := []string{
		"device.12",
		"device-2",
		"device_123",
		"static device",
		"hub.127.0.0.1",
	}

	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemDevice: {
				{
					Get:     true,
					History: false,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("static[! ]*"),
						compileRegexp("dev?"),
					},
				},
			},
			providers.SecSystemAll: {
				{
					Get:     false,
					History: false,
					Command: false,
					Resources: []glob.Glob{
						compileRegexp("hub*"),
					},
				},
			},
		},
	}

	for _, v := range devices {
		assert.False(t, user.DeviceGet(v), "get %s", v)
		assert.False(t, user.DeviceCommand(v), "command %s", v)
		assert.False(t, user.DeviceHistory(v), "history %s", v)
	}
}

// Tests whether baked rules and verbs are interpreted correctly to allow operations.
func TestInternalAllowed(t *testing.T) {
	devices := []string{
		"device.12",
		"device-2",
		"device_123",
		"static device",
		"hub.127.0.0.1",
	}

	user := &AuthenticatedUser{}
	rules := []*providers.BakedRule{
		{
			Get:     false,
			History: true,
			Command: false,
			Resources: []glob.Glob{
				compileRegexp("static device"),
				compileRegexp("dev*"),
			},
		},
		{
			Get:     true,
			History: false,
			Command: true,
			Resources: []glob.Glob{
				compileRegexp("hub*"),
			},
		},
	}

	for _, v := range devices {
		assert.True(t, user.isAllowed(rules, providers.SecVerbAll, v), v)
	}
}

// Tests whether baked rules and verbs are interpreted correctly to forbid operations.
func TestInternalForbidden(t *testing.T) {
	devices := []string{
		"device.12",
		"device-2",
		"device_123",
		"static device",
		"hub.127.0.0.1",
	}

	user := &AuthenticatedUser{}
	rules := []*providers.BakedRule{
		{
			Get:     false,
			History: false,
			Command: false,
			Resources: []glob.Glob{
				compileRegexp("static device"),
				compileRegexp("dev\\S"),
			},
		},
		{
			Get:     false,
			History: false,
			Command: false,
			Resources: []glob.Glob{
				compileRegexp("hub\\S"),
			},
		},
	}

	for _, v := range devices {
		assert.False(t, user.isAllowed(rules, providers.SecVerbAll, v), v)
	}
}

// Tests that workers are allowed.
func TestWorkersAllow(t *testing.T) {
	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemCore: {
				{
					Get:     true,
					History: false,
					Command: false,
					Resources: []glob.Glob{
						compileRegexp("worker"),
					},
				},
			},
		},
	}

	assert.True(t, user.Workers())
}

// Tests that workers are forbidden.
func TestWorkersForbidden(t *testing.T) {
	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemCore: {
				{
					Get:     false,
					History: true,
					Command: false,
					Resources: []glob.Glob{
						compileRegexp("w"),
					},
				},
			},
		},
	}

	assert.False(t, user.Workers())
}

// Tests that workers are allowed.
func TestEntitiesAllow(t *testing.T) {
	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemCore: {
				{
					Get:     true,
					History: false,
					Command: false,
					Resources: []glob.Glob{
						compileRegexp("status"),
					},
				},
			},
		},
	}

	assert.True(t, user.Entities())
}

// Tests that workers are forbidden.
func TestEntitiesForbidden(t *testing.T) {
	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemCore: {
				{
					Get:     false,
					History: true,
					Command: false,
					Resources: []glob.Glob{
						compileRegexp("w"),
					},
				},
			},
		},
	}

	assert.False(t, user.Entities())
}

// Tests that everything is allowed.
func TestAllowedAll(t *testing.T) {
	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemAll: {
				{
					Get:     true,
					History: true,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("*"),
					},
				},
			},
		},
	}

	checkAllAllowed(t, user)
}

// Tests that logs access if forbidden.
func TestLogsForbidden(t *testing.T) {
	user := &AuthenticatedUser{
		Rules: map[providers.SecSystem][]*providers.BakedRule{
			providers.SecSystemCore: {
				{
					Get:     true,
					History: true,
					Command: true,
					Resources: []glob.Glob{
						compileRegexp("w"),
					},
				},
			},
		},
	}

	assert.False(t, user.Logs())
}
