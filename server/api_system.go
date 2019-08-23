package server

import (
	"encoding/json"
	"net/http"
	"sort"

	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/utils"
)

// Performs quick check whether system is OK.
func (s *GoHomeServer) ping(writer http.ResponseWriter, _ *http.Request) {
	if s.Settings.ServiceBus().Ping() != nil {
		respondError(writer, "Service bus unavailable")
		return
	}
	respondOk(writer)
}

// Responds with known workers.
func (s *GoHomeServer) getWorkers(writer http.ResponseWriter, request *http.Request) {
	user := getContextUser(request)
	if !user.Workers() {
		respondForbidden(writer)
		return
	}

	now := utils.TimeNow()
	workers := s.state.GetWorkers()
	workers = append(workers, &knownWorker{
		ID:          "master",
		LastSeenSec: now,
		MaxDevices:  0,
	})

	sort.Slice(workers, func(i, j int) bool {
		return workers[i].ID < workers[j].ID
	})

	// Setting LastSeen property to represent number of seconds from the last event.

	for _, v := range workers {
		v.LastSeenSec = now - v.LastSeen
	}

	respond(writer, workers)
}

// Responds with entities status.
func (s *GoHomeServer) getStatus(writer http.ResponseWriter, request *http.Request) {
	user := getContextUser(request)
	if !user.Entities() {
		respondForbidden(writer)
		return
	}

	entities := s.state.GetEntities()
	entities = append(entities, addMasterComponents(s.triggers, systems.SysTrigger)...)
	entities = append(entities, addMasterComponents(s.extendedAPIs, systems.SysAPI)...)

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Name < entities[j].Name
	})
	respond(writer, entities)
}

// Queries logs.
func (s *GoHomeServer) getLogs(writer http.ResponseWriter, request *http.Request) {
	user := getContextUser(request)
	if !user.Logs() || !s.Logger.GetSpecs().IsHistorySupported {
		respondForbidden(writer)
		return
	}

	dec := json.NewDecoder(request.Body)
	req := &common.LogHistoryRequest{}
	err := dec.Decode(req)

	if err != nil {
		s.Logger.Error("Failed to decode Logs Request", err, common.LogSystemToken, logSystem,
			common.LogUserNameToken, user.Name())
		respondError(writer, "Wrong request")
		return
	}

	respond(writer, s.Logger.Query(req))
}

// Processes known master components.
func addMasterComponents(components []*knownMasterComponent, componentType systems.SystemType) []*knownEntity {
	result := make([]*knownEntity, 0)
	for _, v := range components {
		e := &knownEntity{
			Name:   v.Name,
			Type:   componentType,
			Worker: "master",
		}

		if v.Loaded {
			e.Status = entityLoaded
		} else {
			e.Status = entityLoadFailed
		}

		result = append(result, e)
	}

	return result
}
