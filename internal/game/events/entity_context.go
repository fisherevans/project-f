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
