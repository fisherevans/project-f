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
	Fonts        = DefinedFonts{}

	resourceFonts = LocalResource{
		FileRoot:       "fonts",
		FileExtension:  "ttf",
		FileLoader:     loadFont,
		PostProcessing: loadDefinedFonts,
	}
)

func loadFont(path string, resourceName string, tags []string, data []byte) error {
	ttfFont, err := truetype.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse font from %s: %w", path, err)
	}
	fontRegistry[resourceName] = ttfFont
	return nil
}

const (
	fontNameDogica      = "dogica"
	fontNameDogicaBold  = "dogica_bold"
	fontNameMunro       = "munro"
	fontNameMunroNarrow = "munro_narrow"
	fontNameMunroMicro  = "munro_small"
	FontNameM5x7        = "m5x7"
	FontNameM3x6        = "m3x6"
	FontNameAddStandard = "addstandard"
	FontName3x5         = "3-by-5-pixel-font"

	fontSizeM5x7 = 16
	fontSizeM3x6 = 16

	fontSizeDogicaSmall = 8
	fontSizeDogicaLarge = 16

	fontSizeMunroRegularSmall = 10
	fontSizeMunroRegularLarge = 20

	fontSizeMunroMicroSmall = 10
	fontSizeMunroMicroLarge = 20

	fontSizeAddStandard = 9

	fontSize3x5 = 8
)

type DefinedFonts struct {
	M6   *DefinedFont
	M6x2 *DefinedFont

	M7   *DefinedFont
	M7x2 *DefinedFont

	DogicaRegularSizeSmall *DefinedFont
	DogicaRegularSizeLarge *DefinedFont

	DogicaBoldSizeSmall *DefinedFont
	DogicaBoldSizeLarge *DefinedFont

	MunroRegularSizeSmall *DefinedFont
	MunroRegularSizeLarge *DefinedFont

	MunroNarrowSizeSmall *DefinedFont
	MunroNarrowSizeLarge *DefinedFont

	MunroMicroSizeSmall *DefinedFont
	MunroMicroSizeLarge *DefinedFont

	AddStandard *DefinedFont

	Pixel3x5 *DefinedFont
}

func loadDefinedFonts() error {
	Fonts.M6 = NewDefinedFont(FontNameM3x6, fontSizeM3x6)
	Fonts.M6x2 = NewDefinedFont(FontNameM3x6, fontSizeM3x6*2)

	Fonts.M7 = NewDefinedFont(FontNameM5x7, fontSizeM5x7)
	Fonts.M7x2 = NewDefinedFont(FontNameM5x7, fontSizeM5x7*2)

	Fonts.DogicaRegularSizeSmall = NewDefinedFont(fontNameDogica, fontSizeDogicaSmall)
	Fonts.DogicaRegularSizeLarge = NewDefinedFont(fontNameDogica, fontSizeDogicaSmall)

	Fonts.DogicaBoldSizeSmall = NewDefinedFont(fontNameDogicaBold, fontSizeDogicaLarge)
	Fonts.DogicaBoldSizeLarge = NewDefinedFont(fontNameDogicaBold, fontSizeDogicaLarge)

	Fonts.MunroRegularSizeSmall = NewDefinedFont(fontNameMunro, fontSizeMunroRegularSmall)
	Fonts.MunroRegularSizeLarge = NewDefinedFont(fontNameMunro, fontSizeMunroRegularLarge)

	Fonts.MunroNarrowSizeSmall = NewDefinedFont(fontNameMunroNarrow, fontSizeMunroRegularSmall)
	Fonts.MunroNarrowSizeLarge = NewDefinedFont(fontNameMunroNarrow, fontSizeMunroRegularLarge)

	Fonts.MunroMicroSizeSmall = NewDefinedFont(fontNameMunroMicro, fontSizeMunroMicroSmall)
	Fonts.MunroMicroSizeLarge = NewDefinedFont(fontNameMunroMicro, fontSizeMunroMicroLarge)

	Fonts.AddStandard = NewDefinedFont(FontNameAddStandard, fontSizeAddStandard)

	Fonts.Pixel3x5 = NewDefinedFont(FontName3x5, fontSize3x5)
	return nil
}

type DefinedFont struct {
	TTF   *truetype.Font
	Face  font.Face
	Atlas *text.Atlas
	Size  int
}

func NewDefinedFont(resourceName string, size int) *DefinedFont {
	ttf, exists := fontRegistry[resourceName]
	if !exists {
		log.Fatal().Msgf("font not found in registry: %s", resourceName)
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              float64(size),
		DPI:               72,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: 1,
	})
	atlas := text.NewAtlas(face, text.ASCII)
	return &DefinedFont{
		TTF:   ttf,
		Face:  face,
		Atlas: atlas,
		Size:  size,
	}
}
