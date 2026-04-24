package schema

import "image"

type SpriteMetadata struct {
    Frame          *SpriteFrame                         `yaml:"frame,omitempty" json:"frame,omitempty"`
    Tilesheet      *SpriteTilesheet                     `yaml:"tilesheet,omitempty" json:"tilesheet,omitempty"`
    Animations     map[string]*SpriteTilesheetAnimation `yaml:"animations,omitempty" json:"animations,omitempty"`
    Sprites        map[string]*TilesheetCoordinates     `yaml:"sprites" json:"sprites,omitempty"`
    NonAtlasSprite bool                                 `yaml:"nonAtlasSprite,omitempty" json:"nonAtlasSprite,omitempty"`
}

func (m SpriteMetadata) Init(img image.Image) {
    if m.Frame != nil {
        m.Frame.Init(img)
    }
    if m.Tilesheet != nil {
        m.Tilesheet.Init(img)
    }
}
