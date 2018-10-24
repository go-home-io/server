package server

import (
	"net/http"
	"sort"

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

	workers := s.state.GetWorkers()
	workers = append(workers, &knownWorker{
		ID:         "master",
		LastSeen:   utils.TimeNow(),
		MaxDevices: 0,
	})

	sort.Slice(workers, func(i, j int) bool {
		return workers[i].ID < workers[j].ID
	})

	// Setting LastSeen property to represent number of seconds from the last event.
	now := utils.TimeNow()
	for _, v := range workers {
		v.LastSeen = now - v.LastSeen
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
