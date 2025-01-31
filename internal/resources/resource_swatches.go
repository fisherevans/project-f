package resources

import (
	"fmt"
	"github.com/gopxl/pixel/v2"
)

var (
	Swatches = map[string]Swatch{}

	resourceSwatches = LocalResource{
		FileRoot:        "swatches",
		FileExtension:   "json",
		FileLoader:      jsonLoader(&Swatches),
		ResourceEncoder: jsonEncoder,
	}
)

type Swatch struct {
	Samples map[pixel.Button]SwatchSample `json:"samples"`
}

type SwatchSample struct {
	SpriteId SpriteId `json:"sprite_id"`
}

func (s Swatch) Copy() Swatch {
	out := Swatch{
		Samples: map[pixel.Button]SwatchSample{},
	}
	for button, sample := range s.Samples {
		out.Samples[button] = sample
	}
	return out
}

func SaveAllSwatches() {
	for name := range Swatches {
		SaveSwatch(name)
	}
}

func SaveSwatch(resourceName string) {
	err := save(&Swatches, resourceSwatches, resourceName)
	if err != nil {
		panic(fmt.Sprintf("failed save resource %s: %v", resourceName, err))
	}
}
