package assets

import (
	"embed"
	"path/filepath"
	"runtime"
)

var (
	//go:embed *
	FS embed.FS
)

func LocalFolderPath() string {
	_, packageFilePath, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine caller")
	}
	return filepath.Dir(packageFilePath)
}
