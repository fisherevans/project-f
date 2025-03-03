package resources

import (
	_ "embed"
	"encoding/json"
	"fisherevans.com/project/f/assets"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"
)

var (
	resources = []LocalResource{
		resourceMaps,
		resourceFonts,
		resourceSprites,
		resourceFrames,
		resourceTiledMaps,
	}
)

type fileLoader func(path, resourceName string, tags []string, data []byte) error

type resourceEncoder func(resource any) ([]byte, error)

type LocalResource struct {
	FileRoot        string
	FileExtension   string
	RequiredTags    []string
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
	loadAtlas()
	processFrames()
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
		resourceNameParts := strings.Split(resourceName, ".")
		tags := resourceNameParts[1:]

		if len(localResource.RequiredTags) > 0 {
			isTagged := false
			for _, tag := range localResource.RequiredTags {
				if slices.Contains(tags, tag) {
					isTagged = true
					break
				}
			}
			if !isTagged {
				return nil
			}
		}

		err = localResource.FileLoader(path, resourceNameParts[0], tags, data)
		if err == nil {
			log.Info().Msgf("loaded %s resource: %s", localResource.FileRoot, resourceName)
		} else {
			log.Error().Msgf("failed to load %s resource: %s: %v", localResource.FileRoot, resourceName, err)
		}

		return err
	}
}

type postProcessor[T any] func(resourceName string, newResource T) error

func printLoadSummary[T fmt.Stringer](resourceName string, newResource T) error {
	resourceType := reflect.TypeOf(newResource).Name()
	log.Info().Msgf("[%s] loaded %s: %s", resourceType, resourceName, newResource.String())
	return nil
}

func unmarshaler[T any](dest *map[string]T, unmarshaler func([]byte, any) error, postProcessors ...postProcessor[T]) fileLoader {
	return func(path, resourceName string, tags []string, data []byte) error {
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

func save[T any](source *map[string]T, localResource LocalResource, resourceName string) error {
	// Encode the data
	resource, exists := (*source)[resourceName]
	if !exists {
		return fmt.Errorf("resource %s not found", resourceName)
	}
	data, err := localResource.ResourceEncoder(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal %s/%s: %w", localResource.FileRoot, resourceName, err)
	}

	// Check if the local file exists
	localPath := filepath.Join(assets.LocalFolderPath(), localResource.FileRoot, resourceName+"."+localResource.FileExtension)
	if _, err := os.Stat(localPath); err == nil {
		backupSuffix := time.Now().Format(".backup_2006-01-02_15-04-05")
		backupPath := localPath + backupSuffix
		err := os.Rename(localPath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to create backup %s: %w", backupPath, err)
		}
		err = cleanupBackups(localPath, 3)
		if err != nil {
			log.Error().Msgf("failed to cleanup backup %s: %w", resourceName, err)
		}
	}

	// Write the []byte data to the local file
	err = os.WriteFile(localPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data to file %s: %w", localPath, err)
	}

	log.Info().Msgf("saved %s:%s to %s", localResource.FileRoot, resourceName, localPath)
	return nil
}
