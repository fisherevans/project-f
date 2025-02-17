package main

import (
	_ "fisherevans.com/project/f/cmd/setup"
	"fisherevans.com/project/f/internal/game/runtime"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() { // expose pprof to diagnose memory usage
		http.ListenAndServe("localhost:6060", nil)
	}()
	opengl.Run(runtime.Run)
}
