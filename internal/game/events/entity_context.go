package events

type EntityPosition struct {
	X        int
	Y        int
	IsMoving bool
}

type EntityContext interface {
	Id() string
	Mode() string
	Position() *EntityPosition
}

func NewEphemeralEntityContext(id string) EntityContext {
	return ephemeralEntityContext{
		id: id,
	}
}

type ephemeralEntityContext struct {
	id string
}

func (e ephemeralEntityContext) Id() string {
	return e.id
}

func (e ephemeralEntityContext) Mode() string {
	return ""
}

func (e ephemeralEntityContext) Position() *EntityPosition {
	return nil
}
