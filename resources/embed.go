package resources

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

var (
	//go:embed *
	embeddedResources embed.FS
)

var (
	resources = []LocalResource{
		resourceTilesheets,
		resourceMaps,
		resourceSwatches,
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
		err := fs.WalkDir(embeddedResources, localResource.FileRoot, fsFileHandler(localResource.FileLoader, localResource.FileExtension))
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

func fsFileHandler(handler fileLoader, extension string) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file %s: %w", path, err)
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), "."+extension) {
			return nil
		}

		data, err := embeddedResources.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		filename := filepath.Base(path)
		resourceName := strings.TrimSuffix(filename, filepath.Ext(filename))

		return handler(path, resourceName, data)
	}
}

type postProcessor[T any] func(resourceName string, newResource T) error

func printLoadSummary[T fmt.Stringer](resourceName string, newResource T) error {
	resourceType := reflect.TypeOf(newResource).Name()
	fmt.Printf("[%s] loaded %s: %s\n", resourceType, resourceName, newResource.String())
	return nil
}

func jsonLoader[T any](dest *map[string]T, postProcessors ...postProcessor[T]) fileLoader {
	return func(path, resourceName string, data []byte) error {
		var newResource T
		err := json.Unmarshal(data, &newResource)
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
	_, packageFilePath, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to determine caller")
	}
	folder := filepath.Dir(packageFilePath)
	localPath := filepath.Join(folder, localResource.FileRoot, resourceName+"."+localResource.FileExtension)
	if _, err := os.Stat(localPath); err == nil {
		backupSuffix := time.Now().Format(".backup_2006-01-02_15-04-05")
		backupPath := localPath + backupSuffix
		err := os.Rename(localPath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to create backup %s: %w", backupPath, err)
		}
		err = cleanupBackups(localPath, 3)
		if err != nil {
			fmt.Printf("failed to cleanup backup %s: %w\n", resourceName, err)
		}
	}

	// Write the []byte data to the local file
	err = os.WriteFile(localPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data to file %s: %w", localPath, err)
	}

	fmt.Printf("saved %s:%s to %s\n", localResource.FileRoot, resourceName, localPath)
	return nil
}
