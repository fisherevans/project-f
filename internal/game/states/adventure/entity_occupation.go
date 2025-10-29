package adventure

import "fisherevans.com/project/f/internal/game/events"

type Occupation struct {
	state *State

	entitiesWithinLocations    map[MapLocation]map[string]struct{}
	locationsEntitiesAreWithin map[string]map[MapLocation]struct{}

	entityPrimaryLocation map[string]MapLocation
}

func NewOccupation(state *State) *Occupation {
	return &Occupation{
		state:                      state,
		entitiesWithinLocations:    map[MapLocation]map[string]struct{}{},
		locationsEntitiesAreWithin: map[string]map[MapLocation]struct{}{},
		entityPrimaryLocation:      map[string]MapLocation{},
	}
}

func (o *Occupation) ForEachOccupiedLocation(entityId string, handler func(location MapLocation)) {
	for loc := range o.locationsEntitiesAreWithin[entityId] {
		handler(loc)
	}
}

func (o *Occupation) OccupiedLocationsList(entityId string) []MapLocation {
	var out []MapLocation
	for loc := range o.locationsEntitiesAreWithin[entityId] {
		out = append(out, loc)
	}
	return out
}

func (o *Occupation) ForEachOccupyingEntity(loc MapLocation, handler func(entityId string)) {
	for entityId := range o.entitiesWithinLocations[loc] {
		handler(entityId)
	}
}

func (o *Occupation) OccupyingEntityList(loc MapLocation) []string {
	var out []string
	for id := range o.entitiesWithinLocations[loc] {
		out = append(out, id)
	}
	return out
}

func (o *Occupation) Occupy(id string, loc MapLocation) bool {
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

func (o *Occupation) Vacate(id string, loc MapLocation) bool {
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

func (o *Occupation) GetPrimaryLocation(id string) MapLocation {
	return o.entityPrimaryLocation[id]
}

func (o *Occupation) SetPrimaryLocation(id string, newLocation MapLocation, wasTeleported bool) {
	o.Occupy(id, newLocation)
	priorZones := o.state.zones.ZonesAtSet(o.GetPrimaryLocation(id))
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

	o.entityPrimaryLocation[id] = newLocation
}
