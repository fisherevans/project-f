package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

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

func redirectLogsToFile(path string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory %q: %v\n", dir, err)
		return
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file %q for writing: %v\n", path, err)
		return
	}

	if err := syscall.Dup2(int(f.Fd()), int(os.Stdout.Fd())); err != nil {
		fmt.Fprintf(os.Stderr, "failed to redirect stdout to log file %q: %v\n", path, err)
		_ = f.Close()
		return
	}
	if err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())); err != nil {
		fmt.Fprintf(os.Stderr, "failed to redirect stderr to log file %q: %v\n", path, err)
		_ = f.Close()
		return
	}

	_ = f.Close()
}
