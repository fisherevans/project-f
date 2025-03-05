package resources

import (
	_ "embed"
	"encoding/json"
	"fisherevans.com/project/f/assets"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	DefaultTileSize Pixels = 16
	MapTileSize     Pixels = DefaultTileSize
)

var (
	resources = []LocalResource{
		{
			FileRoot:        "maps",
			FileExtension:   "json",
			FileLoader:      unmarshaler(&maps, json.Unmarshal),
			ResourceEncoder: jsonEncoder,
		},
		{
			FileRoot:      "fonts",
			FileExtension: "ttf",
			FileLoader:    loadFont,
		},
		{
			FileRoot:      "sprites",
			FileExtension: "png",
			FileLoader:    loadSpriteResource,
		},
		{
			FileRoot:      "tiled_maps",
			FileExtension: "tmx",
			FileLoader:    loadTiledMap,
		},
	}
)

type fileLoader func(path, resourceName string, data []byte) error

type resourceEncoder func(resource any) ([]byte, error)

type LocalResource struct {
	FileRoot        string
	FileExtension   string
	FileLoader      fileLoader
	PostProcessing  func() error
	ResourceEncoder resourceEncoder
}

func init() {
	for _, localResource := range resources {
		handler := fsFileHandler(localResource)
		err := fs.WalkDir(assets.FS, localResource.FileRoot, handler)
		if err != nil {
			panic(fmt.Sprintf("failed load load %s resources: %v", localResource.FileRoot, err))
		}
		if localResource.PostProcessing != nil {
			err := localResource.PostProcessing()
			if err != nil {
				panic(fmt.Sprintf("failed postprocess %s resources: %v", localResource.FileRoot, err))
			}
		}
	}
}

func fsFileHandler(localResource LocalResource) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file %s: %w", path, err)
		}

		extensionSuffix := "." + localResource.FileExtension

		if d.IsDir() || !strings.HasSuffix(d.Name(), extensionSuffix) {
			return nil
		}

		data, err := assets.FS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		resourceName := strings.TrimSuffix(strings.TrimPrefix(path, localResource.FileRoot+string(filepath.Separator)), extensionSuffix)

		err = localResource.FileLoader(path, resourceName, data)
		if err == nil {
			log.Info().Msgf("loaded %s resource: %s", localResource.FileRoot, resourceName)
		} else {
			log.Error().Msgf("failed to load %s resource: %s: %v", localResource.FileRoot, resourceName, err)
		}

		return err
	}
}

type postProcessor[T any] func(resourceName string, newResource T) error

func unmarshaler[T any](dest *map[string]T, unmarshaler func([]byte, any) error, postProcessors ...postProcessor[T]) fileLoader {
	return func(path, resourceName string, data []byte) error {
		var newResource T
		err := unmarshaler(data, &newResource)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json for %s: %w", path, err)
		}
		for _, postProcessor := range postProcessors {
			if err := postProcessor(resourceName, newResource); err != nil {
				return err
			}
		}
		(*dest)[resourceName] = newResource
		return nil
	}
}

func jsonEncoder(resource any) ([]byte, error) {
	return json.MarshalIndent(resource, "", "  ")
}

type Pixels int

func (p Pixels) Float() float64 {
	return float64(p)
}

func (p Pixels) Int() int {
	return int(p)
}
