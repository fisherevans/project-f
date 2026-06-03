package combat

// Debug introspection for the dev harness. Mirrors the read-only surface the
// adventure state exposes, so an external driver can assert on combat outcomes
// (sync/shield, statuses, queued skills, phase) rather than only screenshotting.

type HealthDebug struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

type CombatantDebug struct {
	Name               string       `json:"name"`
	IsPlayer           bool         `json:"is_player"`
	IsDead             bool         `json:"is_dead"`
	Sync               *HealthDebug `json:"sync,omitempty"`
	Shield             *HealthDebug `json:"shield,omitempty"`
	Health             *HealthDebug `json:"health,omitempty"`
	Statuses           string       `json:"statuses,omitempty"`
	NextSkill          string       `json:"next_skill,omitempty"`
	NextSkillCommitted bool         `json:"next_skill_committed"`
}

type CombatDebugInfo struct {
	Phase    string         `json:"phase"`
	Player   CombatantDebug `json:"player"`
	Opponent CombatantDebug `json:"opponent"`
}

func (p Phase) String() string {
	switch p {
	case PhaseIntro:
		return "intro"
	case PhaseBattle:
		return "battle"
	case PhaseEnd:
		return "end"
	case PhaseReward:
		return "reward"
	case PhaseTerminal:
		return "terminal"
	default:
		return "unknown"
	}
}

func health(h *HealthState) *HealthDebug {
	if h == nil {
		return nil
	}
	return &HealthDebug{Current: h.GetCurrentInt(), Max: h.Max}
}

func nextSkillName(c Combatant) string {
	if id := c.PeekNextSkill(); id != nil {
		return string(*id)
	}
	return ""
}

// DebugInfo returns a serializable snapshot of the current combat state.
func (s *State) DebugInfo() CombatDebugInfo {
	info := CombatDebugInfo{Phase: s.phase.String()}

	if s.Player != nil {
		info.Player = CombatantDebug{
			Name:               s.Player.Name(),
			IsPlayer:           true,
			IsDead:             s.Player.IsDead(),
			Sync:               health(s.Player.GetCurrentSync()),
			Shield:             health(s.Player.GetCurrentShield()),
			Statuses:           s.Player.GetStatuses().String(),
			NextSkill:          nextSkillName(s.Player),
			NextSkillCommitted: s.Player.IsNextSkillCommitted(),
		}
	}
	if s.Opponent != nil {
		info.Opponent = CombatantDebug{
			Name:               s.Opponent.Name(),
			IsPlayer:           false,
			IsDead:             s.Opponent.IsDead(),
			Health:             health(s.Opponent.GetHealth()),
			Statuses:           s.Opponent.GetStatuses().String(),
			NextSkill:          nextSkillName(s.Opponent),
			NextSkillCommitted: s.Opponent.IsNextSkillCommitted(),
		}
	}
	return info
}
