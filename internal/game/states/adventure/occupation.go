package adventure

import "fisherevans.com/project/f/internal/game/events"

type Occupation struct {
	state             *State
	locationEntities  map[MapLocation]map[string]struct{}
	entitiesLocations map[string]map[MapLocation]struct{}
}

func NewOccupation(state *State) *Occupation {
	return &Occupation{
		state:             state,
		locationEntities:  map[MapLocation]map[string]struct{}{},
		entitiesLocations: map[string]map[MapLocation]struct{}{},
	}
}

func (o *Occupation) ForEachOccupiedLocation(id string, handler func(location MapLocation)) {
	for loc := range o.entitiesLocations[id] {
		handler(loc)
	}
}

func (o *Occupation) OccupiedLocationsList(id string) []MapLocation {
	var out []MapLocation
	for loc := range o.entitiesLocations[id] {
		out = append(out, loc)
	}
	return out
}

func (o *Occupation) ForEachOccupyingEntity(loc MapLocation, handler func(id string)) {
	for id := range o.locationEntities[loc] {
		handler(id)
	}
}

func (o *Occupation) OccupyingEntityList(loc MapLocation) []string {
	var out []string
	for id := range o.locationEntities[loc] {
		out = append(out, id)
	}
	return out
}

func (o *Occupation) Occupy(id string, loc MapLocation) bool {
	if _, exists := o.locationEntities[loc]; !exists {
		o.locationEntities[loc] = map[string]struct{}{}
	}
	o.locationEntities[loc][id] = struct{}{}
	if _, exists := o.entitiesLocations[id]; !exists {
		o.entitiesLocations[id] = map[MapLocation]struct{}{}
	}
	if _, exists := o.entitiesLocations[id][loc]; exists {
		return false
	}
	o.entitiesLocations[id][loc] = struct{}{}
	return true
}

func (o *Occupation) OccupyAndEmit(id string, loc MapLocation, wasTeleported bool) {
	if !o.Occupy(id, loc) {
		return
	}
	o.EmitOccupyEvents(id, loc, wasTeleported)
}

func (o *Occupation) Vacate(id string, loc MapLocation) bool {
	if _, exists := o.entitiesLocations[id]; !exists {
		return false
	}
	delete(o.entitiesLocations[id], loc)
	if len(o.entitiesLocations[id]) == 0 {
		delete(o.entitiesLocations, id)
	}
	if _, exists := o.locationEntities[loc]; !exists {
		return false
	}
	delete(o.locationEntities[loc], id)
	if len(o.locationEntities[loc]) == 0 {
		delete(o.locationEntities, loc)
	}
	return true
}

func (o *Occupation) VacateAndEmit(id string, loc MapLocation, wasTeleported bool) {
	if !o.Vacate(id, loc) {
		return
	}
	o.EmitVacateEvents(id, loc, wasTeleported)
}

func (o *Occupation) EmitVacateEvents(id string, from MapLocation, wasTeleported bool) {
	for _, zoneId := range o.state.zones.ZonesAt(from) {
		o.state.eventDispatcher.Dispatch(events.EventEntityZoneActivity{
			EntityId:      id,
			ZoneId:        zoneId,
			IsEntering:    false,
			WasTeleported: wasTeleported,
		})
	}
}

func (o *Occupation) EmitOccupyEvents(id string, to MapLocation, wasTeleported bool) {
	for _, zoneId := range o.state.zones.ZonesAt(to) {
		o.state.eventDispatcher.Dispatch(events.EventEntityZoneActivity{
			EntityId:      id,
			ZoneId:        zoneId,
			IsEntering:    true,
			WasTeleported: wasTeleported,
		})
	}
}
