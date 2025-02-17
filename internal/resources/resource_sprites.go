package resources

import (
	"bytes"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"image"
	"strings"
)

var (
	Sprites         = map[string]*SpriteReference{}
	NonAtlasSprites = map[string]*SpriteReference{}

	resourceSprites = LocalResource{
		FileRoot:      "sprites",
		FileExtension: "png",
		FileLoader:    loadSprite,
	}
)

var nonAtlasPrefixes = []string{
	"background",
}

func loadSprite(path string, name string, data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error().Msgf("Failed to decode image %s: %v", path, err)
		return nil // Continue with the next file
	}
	atlasSprite := true
	for _, prefix := range nonAtlasPrefixes {
		prefix = prefix + "_"
		if strings.HasPrefix(name, prefix) {
			atlasSprite = false
			//name = strings.TrimPrefix(name, prefix)
			break
		}
	}
	if atlasSprite {
		addAtlasImage(&img, func(picture pixel.Picture, rect pixel.Rect) {
			Sprites[name] = &SpriteReference{
				Source: picture,
				Bounds: rect,
				Sprite: pixel.NewSprite(picture, rect),
			}
		})
		return nil
	}
	picData := pixel.PictureDataFromImage(img)
	NonAtlasSprites[name] = &SpriteReference{
		Source: picData,
		Bounds: picData.Bounds(),
		Sprite: pixel.NewSprite(picData, picData.Bounds()),
	}
	return nil
}

type SpriteReference struct {
	Source pixel.Picture
	Bounds pixel.Rect
	Sprite *pixel.Sprite
}
