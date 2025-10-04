package rpg

type StatusType string

const (
	// kinetic
	// todo fortified
	StatusFortified StatusType = "fortified"
	// thermal
	StatusBurning StatusType = "burning"
	// corrosive
	StatusPoisoned StatusType = "poisoned"
	// voltaic
	StatusIonized StatusType = "ionized"
	// mutagenic
	StatusMending StatusType = "mending"
)

func (s StatusType) Label() string {
	switch s {
	case StatusFortified:
		return "Fortified"
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

type StatusLevel int

const (
	StatusLevel0 StatusLevel = 0
	StatusLevel1 StatusLevel = 1
	StatusLevel2 StatusLevel = 2
	StatusLevel3 StatusLevel = 3
)
