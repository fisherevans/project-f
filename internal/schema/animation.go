package schema

type SpriteTilesheetAnimation struct {
    HSequence       *SpriteTilesheetAnimationHSequence `yaml:"h_sequence" json:"h_sequence,omitempty"`
    VSequence       *SpriteTilesheetAnimationVSequence `yaml:"v_sequence" json:"v_sequence,omitempty"`
    Tiles           []SpriteTilesheetAnimationTile     `yaml:"tiles" json:"tiles,omitempty"`
    FramesPerSecond float64                            `yaml:"framesPerSecond" json:"framesPerSecond"`
    JitterPercent   float64                            `yaml:"jitterPercent" json:"jitterPercent,omitempty"`
    Randomize       *bool                              `yaml:"randomize" json:"randomize,omitempty"`
    PingPong        *bool                              `yaml:"pingPong" json:"pingPong,omitempty"`
    Repeat          *bool                              `yaml:"repeat" json:"repeat,omitempty"`
    Reverse         *bool                              `yaml:"reverse" json:"reverse,omitempty"`
}

type SpriteTilesheetAnimationHSequence struct {
    Row          int       `yaml:"row" json:"row"`
    FromColumn   int       `yaml:"fromColumn" json:"fromColumn,omitempty"`
    ToColumn     int       `yaml:"toColumn" json:"toColumn,omitempty"`
    FrameWeights []float64 `yaml:"frameWeights" json:"frameWeights,omitempty"`
}

type SpriteTilesheetAnimationVSequence struct {
    Column       int       `yaml:"column" json:"column"`
    FromRow      int       `yaml:"fromRow" json:"fromRow,omitempty"`
    ToRow        int       `yaml:"toRow" json:"toRow,omitempty"`
    FrameWeights []float64 `yaml:"frameWeights" json:"frameWeights,omitempty"`
}

type SpriteTilesheetAnimationTile struct {
    Row    int     `yaml:"row" json:"row"`
    Column int     `yaml:"column" json:"column"`
    Weight float64 `yaml:"weight" json:"weight,omitempty"`
}
