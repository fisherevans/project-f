package rpg

type StatusType string

const (
	StatusBurning  StatusType = "burning"
	StatusPoisoned StatusType = "poisoned"
)

func (s StatusType) PastTense() string {
	switch s {
	case StatusBurning:
		return "Burned"
	case StatusPoisoned:
		return "Poisoned"
	}
	return ""
}
