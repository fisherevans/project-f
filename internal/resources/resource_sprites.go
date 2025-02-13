package resources

import (
	"bytes"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"image"
)

var (
	Sprites = map[string]*SpriteReference{}

	resourceSprites = LocalResource{
		FileRoot:      "sprites",
		FileExtension: "png",
		FileLoader:    loadSprite,
	}
)

func loadSprite(path string, name string, data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error().Msgf("Failed to decode image %s: %v", path, err)
		return nil // Continue with the next file
	}
	addAtlasImage(&img, func(picture pixel.Picture, rect pixel.Rect) {
		Sprites[name] = &SpriteReference{
			Source: picture,
			Bounds: rect,
			Sprite: pixel.NewSprite(picture, rect),
		}
	})
	return nil
}

type SpriteReference struct {
	Source pixel.Picture
	Bounds pixel.Rect
	Sprite *pixel.Sprite
}
