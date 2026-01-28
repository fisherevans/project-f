package resources

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/assets"
)

const (
	DefaultTileSize Pixels = 16
	MapTileSize     Pixels = DefaultTileSize
)

var (
	initMu              sync.Mutex
	initialized         bool
	deferredInitializers []func()
)

var (
	resources = []LocalResource{
		{
			FileRoot:        "maps",
			FileExtensions:  []string{"json"},
			FileLoader:      unmarshaler(&maps, json.Unmarshal),
			ResourceEncoder: jsonEncoder,
		},
		{
			FileRoot:       "fonts",
			FileExtensions: []string{"ttf"},
			FileLoader:     loadFont,
		},
		{
			FileRoot:       "sprites",
			FileExtensions: []string{"png"},
			FileLoader:     loadSpriteResource,
		},
		{
			FileRoot:       "tiled_maps",
			FileExtensions: []string{"tmx"},
			FileLoader:     loadTiledMap,
		},
		{
			FileRoot:       "audio/sounds",
			FileExtensions: []string{"wav", "mp3", "ogg"},
			FileLoader:     loadSound,
		},
	}
)

type fileLoader func(path, resourceName string, data []byte) error

type resourceEncoder func(resource any) ([]byte, error)

type LocalResource struct {
	FileRoot        string
	FileExtensions  []string
	FileLoader      fileLoader
	PostProcessing  func() error
	ResourceEncoder resourceEncoder
}

// RunOnceInitialized registers a callback to run after resources are initialized.
// If resources are already initialized, the callback runs immediately.
// This allows packages to defer resource-dependent initialization (like creating atlases)
// until after the resource files have been loaded.
func RunOnceInitialized(fn func()) {
	initMu.Lock()
	defer initMu.Unlock()

	if initialized {
		// Already initialized, run immediately
		fn()
	} else {
		// Defer until Initialize() is called
		deferredInitializers = append(deferredInitializers, fn)
	}
}

func Initialize() {
	initMu.Lock()
	if initialized {
		initMu.Unlock()
		log.Warn().Msg("resources.Initialize() called multiple times")
		return
	}
	initMu.Unlock()

	// Load all resource files
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

	// Mark as initialized and run deferred callbacks
	initMu.Lock()
	initialized = true
	callbacks := deferredInitializers
	deferredInitializers = nil // Clear the list
	initMu.Unlock()

	log.Info().Msgf("Running %d deferred initializers", len(callbacks))
	for _, fn := range callbacks {
		fn()
	}
}

func fsFileHandler(localResource LocalResource) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file %s: %w", path, err)
		}

		doLoad, extension := doLoadFile(d, localResource)
		if !doLoad {
			return nil
		}

		if d.Name() != strings.ToLower(d.Name()) {
			log.Warn().Str("path", path).Msgf("skipping file with upper case letters %s", path)
			return nil
		}

		data, err := assets.FS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		resourceName := strings.TrimSuffix(strings.TrimPrefix(path, localResource.FileRoot+"/"), "."+extension)

		err = localResource.FileLoader(path, resourceName, data)
		if err == nil {
			log.Info().Msgf("loaded %s resource: %s", localResource.FileRoot, resourceName)
		} else {
			log.Error().Msgf("failed to load %s resource: %s: %v", localResource.FileRoot, resourceName, err)
		}

		return err
	}
}

func doLoadFile(d fs.DirEntry, resource LocalResource) (bool, string) {
	if d.IsDir() {
		return false, ""
	}
	for _, extension := range resource.FileExtensions {
		lowerExt := strings.ToLower(extension)
		suffix := "." + lowerExt
		if strings.HasSuffix(strings.ToLower(d.Name()), suffix) {
			return true, lowerExt
		}
	}
	return false, ""
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
