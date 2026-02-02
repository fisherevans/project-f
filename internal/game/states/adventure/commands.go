package adventure

import (
	"strings"

	"fisherevans.com/project/f/internal/game"
)

func (s *State) HandleConsoleInput(cmd string) bool {
	args := strings.Split(cmd, " ")
	switch args[0] {
	case "tp":
		s.commandTeleport(args)
		return true
	case "detail":
		s.commandDetail(args)
		return true
	default:
		return false
	}
}

func (s *State) commandTeleport(args []string) {
	if len(args) < 2 {
		game.Console().Write("supply an entity id")
		return
	}
	_, ok := s.entities.GetEntity(args[1])
	if !ok {
		game.Console().Write("entity not found")
		return
	}
	s.ExecuteSystemEffectsInOrder(NewTeleportPlayerEffect().
		WithToEntityId(args[1]))
	game.Console().Write("teleport complete")
}

func (s *State) commandDetail(args []string) {
	p := s.Globals().Player()
	loc := p.GetLocation().Moved(p.GetFacingDirection())
	game.Console().Writef("location: %s", loc)
	for _, eId := range s.entities.occupations.OccupyingEntityList(loc) {
		game.Console().Writef("entity: %s", eId)
		if p, ok := s.entities.presences[eId]; ok {
			game.Console().Writef("- presence: %#v", p)
		}
		if r, ok := s.entities.renderers[eId]; ok {
			game.Console().Writef("- renderer: %#v", r)
		}
		for _, b := range s.entities.behaviors[eId] {
			game.Console().Writef("- behavior: %#v", b)
		}
		if db, ok := s.entities.disabledBehaviors[eId]; ok {
			for by := range db {
				game.Console().Writef("- behavior disabled by: %#v", by)
			}
		}
		for _, s := range s.entities.soundProviders[eId] {
			game.Console().Writef("- sound: %#v", s)
		}
		if db, ok := s.entities.disabledSounds[eId]; ok {
			for by := range db {
				game.Console().Writef("- sound disabled by: %#v", by)
			}
		}
		if md, ok := s.entities.metadata[eId]; ok {
			game.Console().Writef("- metadata: %#v", md)
		}
	}
	for _, z := range s.zones.ZonesAt(loc) {
		game.Console().Writef("zone: %s", z)
	}
}
