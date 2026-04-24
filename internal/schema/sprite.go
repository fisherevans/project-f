package schema

import "image"

// SpriteMetadata is the top-level structure of a sprite YAML sidecar.
type SpriteMetadata struct {
    Frame          *SpriteFrame                         `yaml:"frame,omitempty"`
    Tilesheet      *SpriteTilesheet                     `yaml:"tilesheet,omitempty"`
    Animations     map[string]*SpriteTilesheetAnimation `yaml:"animations,omitempty"`
    Sprites        map[string]*TilesheetCoordinates     `yaml:"sprites"`
    NonAtlasSprite bool                                 `yaml:"nonAtlasSprite,omitempty"`
}

func (m SpriteMetadata) Init(img image.Image) {
    if m.Frame != nil {
        m.Frame.Init(img)
    }
    if m.Tilesheet != nil {
        m.Tilesheet.Init(img)
    }
}
