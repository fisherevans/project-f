package adventure

type EntityBehavior interface {
	MovementComplete(dispatcher Dispatcher)
	Update(timeDelta float64, dispatcher Dispatcher)
	Reset()
}
