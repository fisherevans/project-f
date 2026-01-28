package main

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/runtime"
	"fisherevans.com/project/f/internal/setup"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

func main() {
	redirectLogsToFile("./game_data/logs.txt")
	setup.SetupLogging()
	opengl.Run(runtime.NewInstance("default", game.StartupDeviceIntent{}).Run)
}
