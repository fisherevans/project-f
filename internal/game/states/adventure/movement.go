package adventure

type MovementRestriction interface {
	EntryAllowed() bool
	CanDashOver() bool
}

type MovementNotAllowed struct{}

func (m MovementNotAllowed) EntryAllowed() bool {
	return false
}

func (m MovementNotAllowed) CanDashOver() bool {
	return false
}

type MovementJumpTile struct {
}

func (m MovementJumpTile) EntryAllowed() bool {
	return false
}

func (m MovementJumpTile) CanDashOver() bool {
	return true
}
