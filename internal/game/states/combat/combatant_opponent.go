package combat

type Opponent interface {
	Combatant
	GetHealth() *HealthState
}
