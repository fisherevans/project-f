package resources

import (
	"bytes"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"image"
	"path/filepath"
	"strings"
)

var (
	sprites         = map[string]*SpriteReference{}
	nonAtlasSprites = map[string]*SpriteReference{}

	resourceSprites = LocalResource{
		FileRoot:      "sprites",
		FileExtension: "png",
		FileLoader:    loadSprite,
	}
)

func GetSprite(name string) *SpriteReference {
	sprite := sprites[name]
	if sprite == nil {
		log.Error().Msgf("missing sprite: %s", name)
	}
	return sprite
}

func GetNonAtlasSprite(name string) *SpriteReference {
	sprite := nonAtlasSprites[name]
	if sprite == nil {
		log.Error().Msgf("missing non-atlas sprite: %s", name)
	}
	return sprite
}

var nonAtlasPrefixes = []string{
	"background",
}

func loadSprite(path string, name string, tags []string, data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error().Msgf("Failed to decode image %s: %v", path, err)
		return nil // Continue with the next file
	}
	atlasSprite := true
	nameParts := strings.Split(name, string(filepath.Separator))
	localName := nameParts[len(nameParts)-1]
	for _, prefix := range nonAtlasPrefixes {
		prefix = prefix + "_"
		if strings.HasPrefix(localName, prefix) {
			atlasSprite = false
			break
		}
	}
	if !atlasSprite {
		picData := pixel.PictureDataFromImage(img)
		nonAtlasSprites[name] = &SpriteReference{
			Source: picData,
			Bounds: picData.Bounds(),
			Sprite: pixel.NewSprite(picData, picData.Bounds()),
		}
		return nil
	}

	registerFunc := func(picture pixel.Picture, rect pixel.Rect) {
		sprites[name] = &SpriteReference{
			Source: picture,
			Bounds: rect,
			Sprite: pixel.NewSprite(picture, rect),
		}
	}

	for _, tag := range tags {
		if !strings.HasPrefix(tag, "tilesheet") {
			continue
		}
		_, tileW, tileH := parseTilesheetName(tag)
		registerFunc = func(picture pixel.Picture, rect pixel.Rect) {
			sliceTilesheet(picture, rect, name, tileW, tileH)
		}
	}

	addAtlasImage(&img, registerFunc)
	return nil
}

type SpriteReference struct {
	Source pixel.Picture
	Bounds pixel.Rect
	Sprite *pixel.Sprite
}

func (s *SpriteReference) HalfDimensions() pixel.Vec {
	return pixel.V(s.Bounds.W()/2.0, s.Bounds.H()/2.0)
}

func (s *SpriteReference) MoveVecBottomLeft() pixel.Vec {
	return pixel.V(s.Bounds.W()/2.0, s.Bounds.H()/2.0)
}
