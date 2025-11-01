package adventure

const (
	ImpedanceImpassable = 1e100
	ImpedanceLow        = 1e2
	ImpedanceBase       = 1e3
	ImpedanceHigh       = 1e4
	ImpedanceExtreme    = 1e5
)

type PathfindingImpedance interface {
	PathfindingImpedanceWeight(id string) float64
}

type StaticImpedance struct {
	impedance float64
}

func NewStaticImpedance(impedance float64) PathfindingImpedance {
	return &StaticImpedance{impedance: impedance}
}

func NewImpassableImpedance() PathfindingImpedance {
	return NewStaticImpedance(ImpedanceImpassable)
}

func (s *StaticImpedance) PathfindingImpedanceWeight(id string) float64 {
	return s.impedance
}

type MovementAwareImpedance struct {
	entity                         Entity
	idleImpedance, movingImpedance float64
}

func NewMovementAwareImpedance(entity Entity, idleImpedance, movingImpedance float64) PathfindingImpedance {
	return &MovementAwareImpedance{
		entity:          entity,
		idleImpedance:   idleImpedance,
		movingImpedance: movingImpedance,
	}
}

func (m *MovementAwareImpedance) PathfindingImpedanceWeight(id string) float64 {
	if m.entity.IsMoving() {
		return m.movingImpedance
	}
	return m.idleImpedance
}

type ConditionalImpedance struct {
	impedance float64
	doImpede  func(id string) bool
}

func NewConditionalImpedance(impedance float64, doImpede func(id string) bool) PathfindingImpedance {
	return &ConditionalImpedance{impedance: impedance, doImpede: doImpede}
}

func (c *ConditionalImpedance) PathfindingImpedanceWeight(id string) float64 {
	if c.doImpede(id) {
		return c.impedance
	}
	return ImpedanceBase
}
