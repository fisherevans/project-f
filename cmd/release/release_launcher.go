package main

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/runtime"
	"fisherevans.com/project/f/internal/setup"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

func main() {
	redirectLogsToFile("./game_data/logs.txt")
	noColor := true
	setup.SetupLoggingWithOptions(setup.LoggingOptions{NoColor: &noColor})
	opengl.Run(runtime.NewInstance("default", func() any {
		return game.StartupDeviceIntent{}
	}).Run)
}
