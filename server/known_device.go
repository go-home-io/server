package server

import (
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
)

// Known devices, received from workers.
type knownDevice struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Worker     string                 `json:"worker"`
	Type       enums.DeviceType       `json:"type"`
	State      map[string]interface{} `json:"state"`
	LastSeen   int64                  `json:"last_seen"`
	Commands   []string               `json:"commands"`
	IsReadOnly bool                   `json:"read_only"`
}

// Get validates whether user can see this device.
func (d *knownDevice) Get(usr *providers.AuthenticatedUser) bool {
	return d.verifyDevice(usr, providers.SecVerbGet)
}

// Command validates whether user can control device.
func (d *knownDevice) Command(usr *providers.AuthenticatedUser) bool {
	return d.verifyDevice(usr, providers.SecVerbCommand)
}

// History validates whether user can access device's history.
func (d *knownDevice) History(usr *providers.AuthenticatedUser) bool {
	return d.verifyDevice(usr, providers.SecVerbHistory)
}

// Verifies device access.
func (d *knownDevice) verifyDevice(usr *providers.AuthenticatedUser, verb providers.SecVerb) bool {
	for k, v := range usr.Rules {
		switch k {
		case providers.SecSystemAll, providers.SecSystemDevice:
			if d.isAllowed(v, verb) {
				return true
			}
		}
	}

	return false
}

// Checks whether device operation is allowed according to the received rules.
// nolint: gocyclo
func (d *knownDevice) isAllowed(rules []*providers.BakedRule, verb providers.SecVerb) bool {
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
			if !v.Get {
				continue
			}
		default:
			if !v.Get && !v.Command && !v.History {
				continue
			}
		}

		for _, r := range v.Resources {
			if r.Match(d.ID) {
				return true
			}
		}
	}

	return false
}
