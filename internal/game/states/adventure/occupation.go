package adventure

import "fisherevans.com/project/f/internal/game/input"

type TileState struct {
	EntitiesWithin []EntityId
}

func (s *State) GetTileState(loc MapLocation) *TileState {
	ts, exists := s.tileStates[loc]
	if !exists {
		ts = &TileState{}
		s.tileStates[loc] = ts
	}
	return ts
}

func (s *State) CanEntityLeave(entityIdMoving EntityId, movementDirection input.Direction, fromLocation MapLocation) bool {
	ts := s.GetTileState(fromLocation)
	for _, entityIdWithin := range ts.EntitiesWithin {
		if entityIdWithin == entityIdMoving {
			continue
		}
		entityWithin := s.entities[entityIdWithin]
		if !entityWithin.CanEgress(movementDirection, entityIdMoving) {
			return false
		}
	}
	return true
}

func (s *State) CanEntityEnter(entityIdMoving EntityId, movementDirection input.Direction, desiredLocation MapLocation) bool {
	ts := s.GetTileState(desiredLocation)
	for _, entityIdWithin := range ts.EntitiesWithin {
		if entityIdWithin == entityIdMoving {
			continue
		}
		entityWithin := s.entities[entityIdWithin]
		if !entityWithin.CanIngress(movementDirection.Opposite(), entityIdMoving) {
			return false
		}
	}
	return true
}

func (ts *TileState) AddEntity(id EntityId) bool {
	for _, existingId := range ts.EntitiesWithin {
		if existingId == id {
			return false
		}
	}
	ts.EntitiesWithin = append(ts.EntitiesWithin, id)
	return true
}

func (ts *TileState) RemoveEntity(id EntityId) bool {
	removeIndex := -1
	for index, existingId := range ts.EntitiesWithin {
		if existingId == id {
			removeIndex = index
			break
		}
	}
	if removeIndex == -1 {
		return false
	}
	ts.EntitiesWithin = append(ts.EntitiesWithin[:removeIndex], ts.EntitiesWithin[removeIndex+1:]...)
	return true
}

type Passable interface {
	CanEgress(side input.Direction, id EntityId) bool
	CanIngress(side input.Direction, id EntityId) bool
}

type blockIngressPassable struct {
	doBlockIngress bool
}

func newPassablePreventIngress(doBlockIngress bool) *blockIngressPassable {
	return &blockIngressPassable{
		doBlockIngress: doBlockIngress,
	}
}

func (sp *blockIngressPassable) CanEgress(side input.Direction, id EntityId) bool {
	return true
}

func (sp *blockIngressPassable) CanIngress(side input.Direction, id EntityId) bool {
	return !sp.doBlockIngress
}

type directionalPassable struct {
	blockedSides map[input.Direction]bool
}

func (dp *directionalPassable) CanEgress(side input.Direction, id EntityId) bool {
	if dp.blockedSides == nil {
		return true
	}
	return dp.blockedSides[side]
}

func (dp *directionalPassable) CanIngress(side input.Direction, id EntityId) bool {
	if dp.blockedSides == nil {
		return false
	}
	return dp.blockedSides[side]
}

type bidirectionalPassable struct {
	blockedEgressSides  map[input.Direction]bool
	blockedIngressSides map[input.Direction]bool
}

func (dp *bidirectionalPassable) CanEgress(side input.Direction, id EntityId) bool {
	if dp.blockedEgressSides == nil {
		return false
	}
	return dp.blockedEgressSides[side]
}

func (dp *bidirectionalPassable) CanIngress(side input.Direction, id EntityId) bool {
	if dp.blockedIngressSides == nil {
		return false
	}
	return dp.blockedIngressSides[side]
}
