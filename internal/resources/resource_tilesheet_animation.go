package resources

var (
	tilesheetAnimations = map[tilesheetAnimationKey]*SpriteTilesheetAnimation{}
)

type tilesheetAnimationKey struct {
	tilesheet string
	name      string
}

type SpriteTilesheetAnimation struct {
	Sequence        *SpriteTilesheetAnimationSequence `yaml:"sequence"`
	Tiles           []SpriteTilesheetAnimationTile    `yaml:"tiles"`
	FramesPerSecond float64                           `yaml:"framesPerSecond"`
	Randomize       bool                              `yaml:"randomize"`
	PingPong        bool                              `yaml:"pingPong"`
}

type SpriteTilesheetAnimationSequence struct {
	Row          int       `yaml:"row"`
	ColumnFrom   int       `yaml:"columnFrom"`
	ColumnTo     int       `yaml:"columnTo"`
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
