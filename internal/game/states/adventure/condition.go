package adventure

type ConditionCheck func(*State, float64) bool

type Condition struct {
	ConditionId string
	Check       ConditionCheck
}

type Conditions struct {
	s          *State
	conditions []Condition
}

func NewConditions(s *State) *Conditions {
	return &Conditions{
		s: s,
	}
}

func (c *Conditions) AddCondition(conditionId string, check ConditionCheck) {
	c.conditions = append(c.conditions, Condition{
		ConditionId: conditionId,
		Check:       check,
	})
}

func (c *Conditions) Update(timeDelta float64) {
	for id := 0; id < len(c.conditions); id++ {
		condition := c.conditions[id]
		if condition.Check == nil || condition.Check(c.s, timeDelta) {
			c.conditions[id] = c.conditions[len(c.conditions)-1]
			c.conditions = c.conditions[:len(c.conditions)-1]
			c.s.planExecutor.MarkConditionComplete(condition.ConditionId)
			id--
		}
	}
}
