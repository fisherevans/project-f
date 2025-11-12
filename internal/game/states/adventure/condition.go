package adventure

type ConditionCheck func(*State, float64) bool

func CameraWithinDistanceToTarget(distanceWithin float64) ConditionCheck {
	return func(s *State, _ float64) bool {
		return s.camera.CurrentLocation().Sub(s.camera.TargetLocation(s)).Len() <= distanceWithin
	}
}

type Condition struct {
	ConditionId  string
	CompletionId string
	Check        ConditionCheck
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

func (c *Conditions) AddCondition(conditionId string, completionId string, check ConditionCheck) {
	c.conditions = append(c.conditions, Condition{
		ConditionId:  conditionId,
		CompletionId: completionId,
		Check:        check,
	})
}

func (c *Conditions) Update(timeDelta float64) {
	for id := 0; id < len(c.conditions); id++ {
		condition := c.conditions[id]
		if condition.Check == nil || condition.Check(c.s, timeDelta) {
			c.conditions[id] = c.conditions[len(c.conditions)-1]
			c.conditions = c.conditions[:len(c.conditions)-1]
			c.s.planExecutor.MarkComplete(condition.CompletionId)
			id--
		}
	}
}
