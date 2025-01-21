package main

import (
	"fisherevans.com/project/f/game/runtime"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

func main() {
	opengl.Run(runtime.Run)
}
