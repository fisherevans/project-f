package resources

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"fisherevans.com/project/f/internal/game/audio"
)

func loadSound(path string, resourceName string, data []byte) error {
	reader := io.NopCloser(bytes.NewReader(data))
	err := audio.GetSystem().Preload(resourceName, filepath.Ext(path), reader)
	if err != nil {
		return fmt.Errorf("audio.GetSystem().Preload: %w", err)
	}
	return nil
}
