package resources

var (
	tilesheetAnimations = map[tilesheetAnimationKey]*SpriteTilesheetAnimation{}
)

type tilesheetAnimationKey struct {
	tilesheet string
	name      string
}

type SpriteTilesheetAnimation struct {
	HSequence       *SpriteTilesheetAnimationHSequence `yaml:"h_sequence"`
	VSequence       *SpriteTilesheetAnimationVSequence `yaml:"v_sequence"`
	Tiles           []SpriteTilesheetAnimationTile     `yaml:"tiles"`
	FramesPerSecond float64                            `yaml:"framesPerSecond"`
	JitterPercent   float64                            `yaml:"jitterPercent"`
	Randomize       *bool                              `yaml:"randomize"`
	PingPong        *bool                              `yaml:"pingPong"`
	Repeat          *bool                              `yaml:"repeat"`
	Reverse         *bool                              `yaml:"reverse"`
}

type SpriteTilesheetAnimationHSequence struct {
	Row          int       `yaml:"row"`
	FromColumn   int       `yaml:"fromColumn"`
	ToColumn     int       `yaml:"toColumn"`
	FrameWeights []float64 `yaml:"frameWeights"`
}

type SpriteTilesheetAnimationVSequence struct {
	Column       int       `yaml:"column"`
	FromRow      int       `yaml:"fromRow"`
	ToRow        int       `yaml:"toRow"`
	FrameWeights []float64 `yaml:"frameWeights"`
}

type SpriteTilesheetAnimationTile struct {
	Row    int     `yaml:"row"`
	Column int     `yaml:"column"`
	Weight float64 `yaml:"weight"`
}

func GetTilesheetAnimation(tilesheet string, name string) *SpriteTilesheetAnimation {
	return tilesheetAnimations[tilesheetAnimationKey{tilesheet, name}]
}
