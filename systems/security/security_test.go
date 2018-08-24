package security

import (
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/providers"
)

// Tests that provider falls back to default FS implementation.
func TestFallbackToDefaultProvider(t *testing.T) {
	found := false
	ctor := &ConstructSecurityProvider{
		Logger: mocks.FakeNewLogger(func(s string) {
			if s == "Loading default user storage" {
				found = true
			}
		}),
	}

	NewSecurityProvider(ctor)
	if !found {
		t.Fail()
	}
}

func getFakeProvider(usr string) providers.ISecurityProvider {
	ctor := &ConstructSecurityProvider{
		Logger:       mocks.FakeNewLogger(nil),
		UserProvider: "test",
		Loader:       mocks.FakeNewPluginLoader(mocks.FakeNewUserStorage(usr)),
		Roles: []*providers.SecRole{
			{
				Name: "1",
				Rules: []providers.SecRoleRule{
					{
						System: providers.SecSystemAll.String(),
						Verbs: []providers.SecVerb{providers.SecVerbAll,
							providers.SecVerbHistory, providers.SecVerbGet},
						Resources: []string{"res"},
					},
				},
				Users: []string{"usr[!0-9]*"},
			},
			{
				Name: "2",
				Rules: []providers.SecRoleRule{
					{
						System: providers.SecSystemDevice.String(),
						Verbs: []providers.SecVerb{providers.SecVerbCommand,
							providers.SecVerbHistory, providers.SecVerbGet},
						Resources: []string{"res*"},
					},
				},
				Users: []string{"usr?"},
			},
			{
				Name: "3",
				Rules: []providers.SecRoleRule{
					{
						System:    "wrong",
						Verbs:     []providers.SecVerb{providers.SecVerbAll},
						Resources: []string{"res1"},
					},
				},
				Users: []string{"user"},
			},
		},
	}

	return NewSecurityProvider(ctor)
}

// Tests fallback to default FS provider.
func TestWrongProvider(t *testing.T) {
	found := false
	ctor := &ConstructSecurityProvider{
		Logger: mocks.FakeNewLogger(func(s string) {
			if s == "Failed to load user storage, defaulting to basic" {
				found = true
			}
		}),
		UserProvider: "wrong",
		Loader:       mocks.FakeNewPluginLoader(nil),
	}

	NewSecurityProvider(ctor)
	if !found {
		t.Fail()
	}
}

// Tests possible errors with roles.
func TestWrongRoles(t *testing.T) {
	wrongResRegex := false
	emptyResource := false
	wrongUserRegex := false
	emptyUsers := false
	emptyRules := false
	ctor := &ConstructSecurityProvider{
		Logger: mocks.FakeNewLogger(func(s string) {
			switch s {
			case "Failed to compile role's resource regexp":
				wrongResRegex = true
			case "Skipping role since resources are empty":
				emptyResource = true
			case "Failed to compile role's user regexp":
				wrongUserRegex = true
			case "Skipping role since users are empty":
				emptyUsers = true
			case "Skipping role since rules are empty":
				emptyRules = true
			}
		}),
		UserProvider: "wrong",
		Loader:       mocks.FakeNewPluginLoader(nil),
		Roles: []*providers.SecRole{
			{
				Name: "1",
				Rules: []providers.SecRoleRule{
					{
						System:    providers.SecSystemAll.String(),
						Verbs:     []providers.SecVerb{providers.SecVerbAll},
						Resources: []string{"[!]"},
					},
				},
				Users: []string{"usr"},
			},
			{
				Name: "2",
				Rules: []providers.SecRoleRule{
					{
						System:    providers.SecSystemAll.String(),
						Verbs:     []providers.SecVerb{providers.SecVerbAll},
						Resources: []string{"res\\s"},
					},
				},
				Users: []string{"(("},
			},
			{
				Name: "3",
				Rules: []providers.SecRoleRule{
					{
						System:    "wrong",
						Verbs:     []providers.SecVerb{providers.SecVerbAll},
						Resources: []string{"res*"},
					},
				},
				Users: []string{"[!]"},
			},
		},
	}

	NewSecurityProvider(ctor)
	if !wrongResRegex || !emptyResource || !wrongUserRegex || !emptyUsers || !emptyRules {
		t.Fail()
	}
}

// Tests correct user validation.
func TestCorrectUsers(t *testing.T) {
	prov := getFakeProvider("usr1")
	usr, err := prov.GetUser(nil)
	if err != nil || 1 != len(usr.Rules) {
		t.Fail()
	}
}

// Tests that incorrect user won't pass validation.
func TestIncorrectUsers(t *testing.T) {
	prov := getFakeProvider("user1")
	usr, err := prov.GetUser(nil)
	if err != nil || 0 != len(usr.Rules) {
		t.Fail()
	}
}

// Tests user not found scenario.
func TestUserNotFound(t *testing.T) {
	prov := getFakeProvider("")
	_, err := prov.GetUser(nil)
	if err == nil {
		t.Fail()
	}
}
