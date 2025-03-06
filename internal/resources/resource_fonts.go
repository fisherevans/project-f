package resources

import (
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/gopxl/pixel/v2/ext/text"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/font"
)

var (
	fontRegistry = map[string]*truetype.Font{}
)

func loadFont(path string, resourceName string, data []byte) error {
	ttfFont, err := truetype.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse font from %s: %w", path, err)
	}
	fontRegistry[resourceName] = ttfFont
	return nil
}

const (
	FontNameDogica      = "dogica"
	FontNameDogicaBold  = "dogica_bold"
	FontNameMunro       = "munro"
	FontNameMunroNarrow = "munro_narrow"
	FontNameMunroMicro  = "munro_small"
	FontNameM5x7        = "m5x7"
	FontNameM3x6        = "m3x6"
	FontNameAddStandard = "addstandard"
	FontName3x5         = "3-by-5-pixel-font"
	FontNameFF          = "ffont"
)

type FontMetadata struct {
	RenderSize   int
	LetterHeight int
	LineSpacing  int
	TailHeight   int
}

func (f FontMetadata) GetLetterHeight() int {
	return f.LetterHeight
}

func (f FontMetadata) GetTailHeight() int {
	return f.TailHeight
}

func (f FontMetadata) GetLineSpacing() int {
	return f.LineSpacing
}

func (f FontMetadata) GetFullLineHeight() int {
	return f.LetterHeight + f.TailHeight
}

var fontMetadata = map[string]FontMetadata{
	FontNameDogica: {
		RenderSize: 8,
		// TODO
	},
	FontNameDogicaBold: {
		RenderSize: 8,
		// TODO
	},
	FontNameMunro: {
		RenderSize: 10,
		// TODO
	},
	FontNameMunroNarrow: {
		RenderSize: 10,
		// TODO
	},
	FontNameMunroMicro: {
		RenderSize: 10,
		// TODO
	},
	FontNameM5x7: {
		RenderSize:   16,
		LetterHeight: 7,
		LineSpacing:  3,
		TailHeight:   2,
	},
	FontNameM3x6: {
		RenderSize:   16,
		LetterHeight: 6,
		LineSpacing:  3,
		TailHeight:   2,
	},
	FontNameAddStandard: {
		RenderSize:   9,
		LetterHeight: 7,
		LineSpacing:  0,
		TailHeight:   2,
	},
	FontName3x5: {
		RenderSize:   8,
		LetterHeight: 5,
		LineSpacing:  0,
		TailHeight:   0,
	},
	FontNameFF: {
		RenderSize:   6,
		LetterHeight: 5,
		LineSpacing:  2,
		TailHeight:   0,
	},
}

type FontInstance struct {
	Name     string
	Metadata FontMetadata
	Atlas    *text.Atlas
}

func CreateFont(fontName string) FontInstance {
	ttf, exists := fontRegistry[fontName]
	if !exists {
		log.Fatal().Msgf("font not found in registry: %s", fontName)
	}
	meta, exist := fontMetadata[fontName]
	if !exist {
		log.Fatal().Msgf("font metadata not found in registry: %s", fontName)
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              float64(meta.RenderSize),
		DPI:               72,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: 1,
	})
	return FontInstance{
		Name:     fontName,
		Metadata: meta,
		Atlas:    text.NewAtlas(face, text.ASCII),
	}
}
