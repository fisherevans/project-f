package adventure

type EntityBehavior interface {
	MovementComplete(dispatcher Dispatcher)
	Update(timeDelta float64, position *EntityPosition, dispatcher Dispatcher)
}
