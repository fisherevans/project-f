package schema

type SpriteTilesheetAnimation struct {
    HSequence       *SpriteTilesheetAnimationHSequence `yaml:"h_sequence,omitempty" json:"h_sequence,omitempty"`
    VSequence       *SpriteTilesheetAnimationVSequence `yaml:"v_sequence,omitempty" json:"v_sequence,omitempty"`
    Tiles           []SpriteTilesheetAnimationTile     `yaml:"tiles,omitempty" json:"tiles,omitempty"`
    FramesPerSecond float64                            `yaml:"framesPerSecond" json:"framesPerSecond"`
    JitterPercent   float64                            `yaml:"jitterPercent,omitempty" json:"jitterPercent,omitempty"`
    Randomize       *bool                              `yaml:"randomize,omitempty" json:"randomize,omitempty"`
    PingPong        *bool                              `yaml:"pingPong,omitempty" json:"pingPong,omitempty"`
    Repeat          *bool                              `yaml:"repeat,omitempty" json:"repeat,omitempty"`
    Reverse         *bool                              `yaml:"reverse,omitempty" json:"reverse,omitempty"`
}

type SpriteTilesheetAnimationHSequence struct {
    Row          int       `yaml:"row" json:"row"`
    FromColumn   int       `yaml:"fromColumn,omitempty" json:"fromColumn,omitempty"`
    ToColumn     int       `yaml:"toColumn,omitempty" json:"toColumn,omitempty"`
    FrameWeights []float64 `yaml:"frameWeights,omitempty" json:"frameWeights,omitempty"`
}

type SpriteTilesheetAnimationVSequence struct {
    Column       int       `yaml:"column" json:"column"`
    FromRow      int       `yaml:"fromRow,omitempty" json:"fromRow,omitempty"`
    ToRow        int       `yaml:"toRow,omitempty" json:"toRow,omitempty"`
    FrameWeights []float64 `yaml:"frameWeights,omitempty" json:"frameWeights,omitempty"`
}

type SpriteTilesheetAnimationTile struct {
    Row    int     `yaml:"row" json:"row"`
    Column int     `yaml:"column" json:"column"`
    Weight float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}
