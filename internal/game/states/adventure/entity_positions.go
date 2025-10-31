package adventure

import (
	"math"

	"fisherevans.com/project/f/internal/game/events"
)

type Positions struct {
	state *State

	entityPositions map[string]MapLocation

	entitiesWithinLocations    map[MapLocation]map[string]struct{}
	locationsEntitiesAreWithin map[string]map[MapLocation]struct{}
}

func NewPositions(state *State) *Positions {
	return &Positions{
		state:                      state,
		entitiesWithinLocations:    map[MapLocation]map[string]struct{}{},
		locationsEntitiesAreWithin: map[string]map[MapLocation]struct{}{},
		entityPositions:            map[string]MapLocation{},
	}
}

func (o *Positions) RemoveEntity(id string) {
	var toVacate []MapLocation
	for loc := range o.locationsEntitiesAreWithin[id] {
		toVacate = append(toVacate, loc)
	}
	for _, loc := range toVacate {
		o.Vacate(id, loc)
		o.SetPosition(id, MapLocation{
			X: math.MinInt,
			Y: math.MinInt,
		}, true)
	}
	delete(o.entityPositions, id)
	delete(o.locationsEntitiesAreWithin, id)
}

func (o *Positions) ForEachOccupiedLocation(entityId string, handler func(location MapLocation)) {
	for loc := range o.locationsEntitiesAreWithin[entityId] {
		handler(loc)
	}
}

func (o *Positions) OccupiedLocationsList(entityId string) []MapLocation {
	var out []MapLocation
	for loc := range o.locationsEntitiesAreWithin[entityId] {
		out = append(out, loc)
	}
	return out
}

func (o *Positions) ForEachOccupyingEntity(loc MapLocation, handler func(entityId string)) {
	for entityId := range o.entitiesWithinLocations[loc] {
		handler(entityId)
	}
}

func (o *Positions) OccupyingEntityList(loc MapLocation) []string {
	var out []string
	for id := range o.entitiesWithinLocations[loc] {
		out = append(out, id)
	}
	return out
}

func (o *Positions) Occupy(id string, loc MapLocation) bool {
	if _, exists := o.entitiesWithinLocations[loc]; !exists {
		o.entitiesWithinLocations[loc] = map[string]struct{}{}
	}
	o.entitiesWithinLocations[loc][id] = struct{}{}
	if _, exists := o.locationsEntitiesAreWithin[id]; !exists {
		o.locationsEntitiesAreWithin[id] = map[MapLocation]struct{}{}
	}
	if _, exists := o.locationsEntitiesAreWithin[id][loc]; exists {
		return false
	}
	o.locationsEntitiesAreWithin[id][loc] = struct{}{}
	return true
}

func (o *Positions) Vacate(id string, loc MapLocation) bool {
	if _, exists := o.locationsEntitiesAreWithin[id]; !exists {
		return false
	}
	delete(o.locationsEntitiesAreWithin[id], loc)
	if len(o.locationsEntitiesAreWithin[id]) == 0 {
		delete(o.locationsEntitiesAreWithin, id)
	}
	if _, exists := o.entitiesWithinLocations[loc]; !exists {
		return false
	}
	delete(o.entitiesWithinLocations[loc], id)
	if len(o.entitiesWithinLocations[loc]) == 0 {
		delete(o.entitiesWithinLocations, loc)
	}
	return true
}

func (o *Positions) GetPosition(id string) MapLocation {
	return o.entityPositions[id]
}

func (o *Positions) SetPosition(id string, newLocation MapLocation, wasTeleported bool) {
	o.Occupy(id, newLocation)
	priorZones := o.state.zones.ZonesAtSet(o.GetPosition(id))
	nextZones := o.state.zones.ZonesAtSet(newLocation)

	emitZoneEvent := func(zoneId string, isEntering bool) {
		o.state.eventDispatcher.Dispatch(events.EventEntityZoneActivity{
			EntityId:      id,
			ZoneId:        zoneId,
			IsEntering:    isEntering,
			WasTeleported: wasTeleported,
		})
	}

	// Emit exit events for zones no longer occupied
	for zoneId := range priorZones {
		if _, stillInZone := nextZones[zoneId]; !stillInZone {
			emitZoneEvent(zoneId, false)
		}
	}

	// Emit enter events for newly occupied zones
	for zoneId := range nextZones {
		if _, wasInZone := priorZones[zoneId]; !wasInZone {
			emitZoneEvent(zoneId, true)
		}
	}

	o.entityPositions[id] = newLocation
}
