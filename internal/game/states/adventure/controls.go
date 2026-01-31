package adventure

import "fisherevans.com/project/f/internal/resources"

func (s *State) EvaluateControlToggles(def bool, toggles []resources.ControlToggle) bool {
	player := s.globals.Player()
	location := player.GetLocation()
	zoneSet := s.zones.ZonesAtSet(location)
	for _, toggle := range toggles {
		if toggle.ZoneId != "" && zoneSet.Contains(toggle.ZoneId) {
			return !def
		}
		if toggle.GlobalKey != "" && s.globals.Get(toggle.GlobalKey).AsBool(false) {
			return !def
		}
	}
	return def
}
