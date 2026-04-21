package main

import (
	// used in dev builds to expose profiler
	"net/http"
	_ "net/http/pprof"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/devscenes"
	"fisherevans.com/project/f/internal/game/runtime"
	"fisherevans.com/project/f/internal/setup"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

func main() {
	setup.SetupLogging()
	instance := runtime.NewInstance("default", func() any {
		return game.StartupDeviceIntent{}
	}).WithDevScenes(devscenes.Scenes())
	go func() { // expose pprof to diagnose memory usage
		http.ListenAndServe("localhost:6060", nil)
	}()
	opengl.Run(instance.Run)
}
