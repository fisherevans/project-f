package adventure

type EntityBehavior interface {
	MovementComplete()
	Update(timeDelta float64)
	Reset()
}

type FaceEntityBehavior struct {
	entity Entity
	facing string
}

func AttachFaceEntityBehavior(entity Entity, facing string) *FaceEntityBehavior {
	b := &FaceEntityBehavior{
		entity: entity,
		facing: facing,
	}
	entity.PushBehavior(b)
	return b
}

func (b *FaceEntityBehavior) SetFacing(facing string) {
	b.facing = facing
}

func (b *FaceEntityBehavior) Reset() {
	b.facing = ""
}

func (b *FaceEntityBehavior) MovementComplete() {

}

func (b *FaceEntityBehavior) Update(timeDelta float64) {
	if b.entity.IsMoving() {
		return
	}
	if b.facing != "" {
		talkingTowardsEnt, exists := b.entity.GetSystem().GetEntity(b.facing)
		if exists {
			faceDir := DirectionTowards(b.entity.GetPreciseLocation(), talkingTowardsEnt.GetPreciseLocation())
			b.entity.SetFacingDirection(faceDir)
		}
	}
}
