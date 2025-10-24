package adventure

// EntityBehavior defines how an entity behaves each frame
type EntityBehavior interface {
	Update(s *State, timeDelta float64)
	OnMovementComplete(s *State, movement *MovementState) bool
}

// BehaviorController manages all entity behaviors
type BehaviorController struct {
	behaviors map[EntityId]EntityBehavior
}

func NewBehaviorController() *BehaviorController {
	return &BehaviorController{
		behaviors: make(map[EntityId]EntityBehavior),
	}
}

func (bc *BehaviorController) Register(entityId EntityId, behavior EntityBehavior) {
	bc.behaviors[entityId] = behavior
}

func (bc *BehaviorController) Unregister(entityId EntityId) {
	delete(bc.behaviors, entityId)
}

func (bc *BehaviorController) Get(entityId EntityId) EntityBehavior {
	return bc.behaviors[entityId]
}

func (bc *BehaviorController) Update(s *State, timeDelta float64) {
	for _, behavior := range bc.behaviors {
		behavior.Update(s, timeDelta)
	}
}
