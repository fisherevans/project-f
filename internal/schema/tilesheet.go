package schema

import "image"

// SpriteTilesheet describes a grid-based sprite sheet.
type SpriteTilesheet struct {
    TileWidth  Pixels `yaml:"tileWidth"`
    TileHeight Pixels `yaml:"tileHeight"`

    // Set when loaded from an image.
    Columns int `yaml:"-"`
    Rows    int `yaml:"-"`
}

func (t *SpriteTilesheet) Init(img image.Image) {
    t.Columns = img.Bounds().Dx() / t.TileWidth.Int()
    t.Rows = img.Bounds().Dy() / t.TileHeight.Int()
}

// TilesheetCoordinates identifies a named sprite cell in a tilesheet.
type TilesheetCoordinates struct {
    Row    int `yaml:"row"`
    Column int `yaml:"column"`
}
