package adventure

type MovementRestriction interface {
	EntryAllowed() bool
}

type MovementNotAllowed struct{}

func (m MovementNotAllowed) EntryAllowed() bool {
	return false
}
