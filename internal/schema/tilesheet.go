package schema

import "image"

type SpriteTilesheet struct {
    TileWidth  Pixels `yaml:"tileWidth" json:"tileWidth"`
    TileHeight Pixels `yaml:"tileHeight" json:"tileHeight"`

    Columns int `yaml:"-" json:"columns,omitempty"`
    Rows    int `yaml:"-" json:"rows,omitempty"`
}

func (t *SpriteTilesheet) Init(img image.Image) {
    t.Columns = img.Bounds().Dx() / t.TileWidth.Int()
    t.Rows = img.Bounds().Dy() / t.TileHeight.Int()
}

type TilesheetCoordinates struct {
    Row    int `yaml:"row" json:"row"`
    Column int `yaml:"column" json:"column"`
}
