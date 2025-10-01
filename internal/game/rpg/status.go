package rpg

type StatusType string

const (
	// kinetic
	// todo fortified
	// thermal
	StatusBurning StatusType = "burning"
	// corrosive
	StatusPoisoned StatusType = "poisoned"
	// voltaic
	// todo ionized - makes next damage higher
	// mutagenic
	// todo mending - heals sync
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
