package security

import "go-home.io/x/server/providers"

// AuthenticatedUser has data with authenticated user, returned by user store.
type AuthenticatedUser struct {
	Username string
	Rules    map[providers.SecSystem][]*providers.BakedRule
}

// Name returns the user name.
func (u *AuthenticatedUser) Name() string {
	return u.Username
}

// TriggerGet get verifies whether user is allowed to get a trigger.
func (u *AuthenticatedUser) TriggerGet(triggerID string) bool {
	return u.verifyEntity(providers.SecSystemTrigger, providers.SecVerbGet, triggerID)
}

// TriggerHistory verifies whether user is allowed to query a device history.
func (u *AuthenticatedUser) TriggerHistory(triggerID string) bool {
	return u.verifyEntity(providers.SecSystemTrigger, providers.SecVerbHistory, triggerID)
}

// DeviceGet verifies whether user is allowed to get a device.
func (u *AuthenticatedUser) DeviceGet(deviceID string) bool {
	return u.verifyEntity(providers.SecSystemDevice, providers.SecVerbGet, deviceID)
}

// DeviceCommand verifies whether user is allowed to issue a command to a device.
func (u *AuthenticatedUser) DeviceCommand(deviceID string) bool {
	return u.verifyEntity(providers.SecSystemDevice, providers.SecVerbCommand, deviceID)
}

// DeviceHistory verifies whether user is allowed to query a device history.
func (u *AuthenticatedUser) DeviceHistory(deviceID string) bool {
	return u.verifyEntity(providers.SecSystemDevice, providers.SecVerbHistory, deviceID)
}

// Workers verifies whether user is allowed to get workers.
func (u *AuthenticatedUser) Workers() bool {
	return u.verifyEntity(providers.SecSystemCore, providers.SecVerbGet, "worker")
}

// Entities verifies whether user is allowed to get config entities.
func (u *AuthenticatedUser) Entities() bool {
	return u.verifyEntity(providers.SecSystemCore, providers.SecVerbGet, "status")
}

// Logs verifies whether user is allowed to query logs.
func (u *AuthenticatedUser) Logs() bool {
	return u.verifyEntity(providers.SecSystemCore, providers.SecVerbGet, "logs")
}

// Verifies access.
func (u *AuthenticatedUser) verifyEntity(system providers.SecSystem, verb providers.SecVerb, entityID string) bool {
	for k, v := range u.Rules {
		if k == providers.SecSystemAll || k == system {
			if u.isAllowed(v, verb, entityID) {
				return true
			}
		}
	}

	return false
}

// Checks whether entity operation is allowed according to the received rules.
// nolint: gocyclo
func (u *AuthenticatedUser) isAllowed(rules []*providers.BakedRule, verb providers.SecVerb, entityID string) bool {
	for _, v := range rules {
		switch verb {
		case providers.SecVerbGet:
			if !v.Get {
				continue
			}
		case providers.SecVerbCommand:
			if !v.Command {
				continue
			}
		case providers.SecVerbHistory:
			if !v.History {
				continue
			}
		default:
			if !v.Get && !v.Command && !v.History {
				continue
			}
		}

		for _, r := range v.Resources {
			if r.Match(entityID) {
				return true
			}
		}
	}

	return false
}
