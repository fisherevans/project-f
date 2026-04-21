//go:build js && wasm

package main

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/devscenes"
	"fisherevans.com/project/f/internal/game/runtime"
	"fisherevans.com/project/f/internal/setup"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

func main() {
	opengl.Run(func() {
		noColor := true
		setup.SetupLoggingWithOptions(setup.LoggingOptions{NoColor: &noColor})
		setup.LogMetadata()
		runtime.NewInstance("default", func() any {
			return game.StartupDeviceIntent{}
		}).WithDevScenes(devscenes.Scenes()).Run()
	})
}
