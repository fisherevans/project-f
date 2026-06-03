package main

import (
    "fmt"
    // used in dev builds to expose profiler
    "net/http"
    _ "net/http/pprof"
    "os"

    "fisherevans.com/project/f/cmd/development/debugapi"
    "fisherevans.com/project/f/internal/game"
    "fisherevans.com/project/f/internal/game/devscenes"
    "fisherevans.com/project/f/internal/game/rpg"
    "fisherevans.com/project/f/internal/game/runtime"
    "fisherevans.com/project/f/internal/game/states/adventure"
    "fisherevans.com/project/f/internal/overlays"
    "fisherevans.com/project/f/internal/setup"
    "github.com/gopxl/pixel/v2/backends/opengl"
    "github.com/rs/zerolog/log"
)

func main() {
    setup.SetupLogging()

    queue := debugapi.NewCommandQueue(64)
    harness := runtime.NewHarness()

    saveId := os.Getenv("PRIMORTAL_SAVE")
    if saveId == "" {
        saveId = "default"
    }
    instance := runtime.NewInstance(saveId, bootIntentFactory()).
        WithDevScenes(devscenes.Scenes()).
        WithDebugDrain(queue.DrainOnGameThread).
        WithHarness(harness)

    installReload()
    debugServer := debugapi.NewServer(queue, instance.DebugHighlight, instance)

    go func() { // expose pprof to diagnose memory usage
        http.ListenAndServe("localhost:6060", nil)
    }()
    go func() {
        log.Info().Msg("debug API listening on http://localhost:8091")
        http.ListenAndServe("localhost:8091", debugServer)
    }()
    opengl.Run(instance.Run)
}

// bootIntentFactory returns the factory the runtime uses for the initial state
// and for resets. By default it boots the startup device screen; environment
// variables let an external driver land directly in a target state, skipping the
// menu flow:
//
//   PRIMORTAL_BOOT_MAP=hq            boot straight into an adventure map
//   PRIMORTAL_BOOT_WAYPOINT=foo      optional waypoint within that map
//   PRIMORTAL_BOOT_SCENE="Adventure: HQ"  boot a named dev scene
func bootIntentFactory() func() any {
    if mapName := os.Getenv("PRIMORTAL_BOOT_MAP"); mapName != "" {
        waypoint := os.Getenv("PRIMORTAL_BOOT_WAYPOINT")
        log.Info().Str("map", mapName).Str("waypoint", waypoint).Msg("boot directive: adventure map")
        return func() any {
            return game.AdventureIntent{MapName: mapName, Waypoint: waypoint}
        }
    }
    if sceneName := os.Getenv("PRIMORTAL_BOOT_SCENE"); sceneName != "" {
        for _, sc := range devscenes.Scenes() {
            if sc.Name == sceneName {
                log.Info().Str("scene", sceneName).Msg("boot directive: dev scene")
                factory := sc.Factory
                return func() any { return factory() }
            }
        }
        log.Warn().Str("scene", sceneName).Msg("boot directive: unknown scene, falling back to startup")
    }
    return func() any {
        return game.StartupDeviceIntent{}
    }
}

// installReload wires the content hot-reload endpoint. It reads YAML from disk
// (not the embedded snapshot) so edits take effect without a rebuild. The assets
// directory defaults to ./assets relative to the launch cwd; override with
// PRIMORTAL_ASSETS_DIR. The callback runs on the game thread (the debug API
// dispatches it through the command queue).
func installReload() {
    assetsDir := os.Getenv("PRIMORTAL_ASSETS_DIR")
    if assetsDir == "" {
        assetsDir = "assets"
    }
    debugapi.SetReloadFunc(func(kind string) (msg string, err error) {
        // A malformed edit (or a registry that loading panics on) must not take
        // down the dev process. Recover and report it as an error instead.
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("reload panicked: %v", r)
            }
        }()
        if _, statErr := os.Stat(assetsDir); statErr != nil {
            return "", fmt.Errorf("assets dir %q not found (set PRIMORTAL_ASSETS_DIR): %w", assetsDir, statErr)
        }
        diskFS := os.DirFS(assetsDir)
        switch kind {
        case "scripts":
            if err := adventure.ReloadScriptDefs(diskFS); err != nil {
                return "", err
            }
            if adv, ok := game.GetActiveState().(*adventure.State); ok {
                adv.ReloadMap()
                return "scripts reloaded; current map reloading (player returns to spawn)", nil
            }
            return "scripts reloaded (no live adventure map to rebind)", nil
        case "rpg":
            if err := rpg.Reload(diskFS); err != nil {
                return "", err
            }
            return "rpg skills/primortals reloaded", nil
        case "overlays":
            if err := overlays.LoadFromFS(diskFS); err != nil {
                return "", err
            }
            return "overlay flows/rects reloaded", nil
        default:
            return "", fmt.Errorf("unknown reload kind %q (scripts|rpg|overlays)", kind)
        }
    })
}
