package adventure

const (
	ImpedanceImpassable = 1e100
	ImpedanceLow        = 10.
	ImpedanceNormal     = 20.
	ImpedanceHigh       = 40.
	ImpedanceExtreme    = 100.
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
	return ImpedanceNormal
}
