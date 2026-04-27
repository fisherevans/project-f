package main

import (
    // used in dev builds to expose profiler
    "net/http"
    _ "net/http/pprof"

    "fisherevans.com/project/f/cmd/development/debugapi"
    "fisherevans.com/project/f/internal/game"
    "fisherevans.com/project/f/internal/game/devscenes"
    "fisherevans.com/project/f/internal/game/runtime"
    "fisherevans.com/project/f/internal/setup"
    "github.com/gopxl/pixel/v2/backends/opengl"
    "github.com/rs/zerolog/log"
)

func main() {
    setup.SetupLogging()

    queue := debugapi.NewCommandQueue(64)

    instance := runtime.NewInstance("default", func() any {
        return game.StartupDeviceIntent{}
    }).
        WithDevScenes(devscenes.Scenes()).
        WithDebugDrain(queue.DrainOnGameThread)

    debugServer := debugapi.NewServer(queue, instance.DebugHighlight)

    go func() { // expose pprof to diagnose memory usage
        http.ListenAndServe("localhost:6060", nil)
    }()
    go func() {
        log.Info().Msg("debug API listening on http://localhost:8091")
        http.ListenAndServe("localhost:8091", debugServer)
    }()
    opengl.Run(instance.Run)
}
