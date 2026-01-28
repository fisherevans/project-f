package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gopxl/pixel/v2/backends/opengl"

	_ "fisherevans.com/project/f/cmd/setup"
	"fisherevans.com/project/f/internal/game/runtime"
)

func main() {
	instance := runtime.NewInstance()

	go func() { // expose pprof to diagnose memory usage
		http.ListenAndServe("localhost:6060", nil)
	}()

	opengl.Run(instance.Run)
}
