package adventure

type EntityBehavior interface {
	MovementComplete()
	Update(timeDelta float64)
	Reset()
}
