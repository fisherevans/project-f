package events

type Location struct {
	X int
	Y int
}

func (l *Location) Validate() error {
	// Location fields are always valid (can be 0)
	return nil
}
