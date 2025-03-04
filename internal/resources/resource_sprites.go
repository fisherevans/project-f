package resources

import (
	"bytes"
	"errors"
	"fisherevans.com/project/f/assets"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"image"
	"io/fs"
	"path/filepath"
	"strings"
)

var (
	spriteResources = map[string]spriteResource{}
)

type spriteResource struct {
	data     image.Image
	metadata SpriteMetadata
}

type SpriteMetadata struct {
	Frame          *SpriteFrame     `yaml:"frame,omitempty"`
	Tilesheet      *SpriteTilesheet `yaml:"tilesheet,omitempty"`
	NonAtlasSprite bool             `yaml:"nonAtlasSprite,omitempty"`
}

func (m SpriteMetadata) init(img image.Image) {
	if m.Frame != nil {
		m.Frame.init(img)
	}
	if m.Tilesheet != nil {
		m.Tilesheet.init(img)
	}
}

func LoadSprite(name string) *pixel.Sprite {
	sprite, exists := spriteResources[name]
	if !exists {
		log.Error().Msgf("missing sprite: %s", name)
	}
	pd := pixel.PictureDataFromImage(sprite.data)
	return pixel.NewSprite(pd, pd.Bounds())
}

func loadSpriteResource(path string, name string, data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error().Msgf("Failed to decode image %s: %v", path, err)
		return nil // Continue with the next file
	}

	metadataPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".yaml"
	metadataData, err := assets.FS.ReadFile(metadataPath)
	var metadata SpriteMetadata
	if !errors.Is(err, fs.ErrNotExist) {
		if err != nil {
			log.Fatal().Msgf("Failed to read metadata %s: %v", metadataPath, err)
		}
		err = yaml.Unmarshal(metadataData, &metadata)
		if err != nil {
			log.Fatal().Msgf("Failed to decode metadata %s: %v", metadataPath, err)
		}
	}
	metadata.init(img)

	spriteResources[name] = spriteResource{
		data:     img,
		metadata: metadata,
	}
	if metadata.Tilesheet != nil {
		tilesheets[name] = metadata.Tilesheet
	}
	if metadata.Frame != nil {
		frames[name] = metadata.Frame
	}

	return nil
}
