package main

import (
	"fmt"
	"path/filepath"
	"runtime/debug"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/runtime"
	"fisherevans.com/project/f/internal/setup"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/ncruces/zenity"
)

func main() {
	logPath := "./game_data/logs.txt"
	absLogPath, _ := filepath.Abs(logPath)
	handlePanic := func(r any) {
		if r == nil {
			return
		}
		setup.LogPanic(r, debug.Stack(), "The game has crashed unexpectedly")
		msg := fmt.Sprintf("The game has crashed unexpectedly.\n\nPlease share the log file with fisher for debugging:\n%s\n\nPanic details:\n%v", absLogPath, r)
		_ = zenity.Error(msg, zenity.Title("Primortal Panic"))
	}

	defer func() { handlePanic(recover()) }()
	opengl.Run(func() {
		defer func() { handlePanic(recover()) }()
		redirectLogsToFile(logPath)
		noColor := true
		setup.SetupLoggingWithOptions(setup.LoggingOptions{NoColor: &noColor})
		setup.LogMetadata()
		runtime.NewInstance("default", func() any {
			return game.StartupDeviceIntent{}
		}).Run()
	})
}
