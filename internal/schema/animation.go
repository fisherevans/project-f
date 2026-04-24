package schema

// SpriteTilesheetAnimation defines playback for a named animation in a tilesheet sidecar.
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

// SpriteTilesheetAnimationHSequence selects consecutive columns in one row.
type SpriteTilesheetAnimationHSequence struct {
    Row          int       `yaml:"row"`
    FromColumn   int       `yaml:"fromColumn"`
    ToColumn     int       `yaml:"toColumn"`
    FrameWeights []float64 `yaml:"frameWeights"`
}

// SpriteTilesheetAnimationVSequence selects consecutive rows in one column.
type SpriteTilesheetAnimationVSequence struct {
    Column       int       `yaml:"column"`
    FromRow      int       `yaml:"fromRow"`
    ToRow        int       `yaml:"toRow"`
    FrameWeights []float64 `yaml:"frameWeights"`
}

// SpriteTilesheetAnimationTile is an arbitrary frame reference.
type SpriteTilesheetAnimationTile struct {
    Row    int     `yaml:"row"`
    Column int     `yaml:"column"`
    Weight float64 `yaml:"weight"`
}
