package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"fisherevans.com/project/f/assets"
	"fisherevans.com/project/f/cmd/asset_editor/server"
)

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
	dev := flag.Bool("dev", false, "enable dev mode (CORS, Vite proxy)")
	port := flag.Int("port", 8090, "HTTP listen port")
	assetsDir := flag.String("assets-dir", "", "path to assets directory (default: auto-detect from module)")
	flag.Parse()

	dir := *assetsDir
	if dir == "" {
		dir = assets.LocalFolderPath()
	}

	var frontFS fs.FS
	if !*dev {
		sub, err := fs.Sub(frontendFS, "frontend/dist")
		if err != nil {
			log.Fatalf("embedded frontend: %v", err)
		}
		frontFS = sub
	}

	srv := server.New(dir, *dev, frontFS)

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)
	log.Printf("asset editor listening on %s (assets: %s, dev: %v)", url, dir, *dev)

	go openBrowser(url)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
