package rpg

const (
	RunStateKeyElythium = "elythium"
)

type Run struct {
	State *State
}

func NewRun() *Run {
	return &Run{
		State: NewState(),
	}
}
