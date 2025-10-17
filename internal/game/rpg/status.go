package rpg

type StatusType string

const (
	// kinetic
	StatusWarded StatusType = "warded"
	// thermal
	StatusBurning StatusType = "burning"
	// corrosive
	StatusPoisoned StatusType = "poisoned"
	// voltaic
	StatusIonized StatusType = "ionized"
	// mutagenic
	StatusMending StatusType = "mending"
)

func (s StatusType) PastTense() string {
	switch s {
	case StatusWarded:
		return "Warded"
	case StatusBurning:
		return "Burned"
	case StatusPoisoned:
		return "Poisoned"
	case StatusIonized:
		return "Ionized"
	case StatusMending:
		return "Mending"
	}
	return ""
}

func (s StatusType) CurrentTense() string {
	switch s {
	case StatusWarded:
		return "Warded"
	case StatusBurning:
		return "Burning"
	case StatusPoisoned:
		return "Poisoned"
	case StatusIonized:
		return "Ionized"
	case StatusMending:
		return "Mending"
	}
	return ""
}

func (s StatusType) Description() string {
	switch s {
	case StatusWarded:
		return "Reduces incoming damage"
	case StatusBurning:
		return "Deals damage over time, increasing with exposure"
	case StatusPoisoned:
		return "Deals percentage-based damage over time"
	case StatusIonized:
		return "Amplifies damage received"
	case StatusMending:
		return "Restores health over time"
	}
	return ""
}

type StatusLevel int

const (
	StatusLevel0 StatusLevel = 0
	StatusLevel1 StatusLevel = 1
	StatusLevel2 StatusLevel = 2
	StatusLevel3 StatusLevel = 3
)
