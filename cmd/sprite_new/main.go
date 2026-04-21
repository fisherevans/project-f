// sprite_new scaffolds a new tilesheet asset: a transparent PNG at the correct
// dimensions, a YAML sidecar using the project's tilesheet grammar, and a
// matching .aseprite file if the Aseprite CLI is available.
//
// Usage:
//
//	go run ./cmd/sprite_new \
//	    -name assets/sprites/overlay/icons \
//	    -tile-width 16 -tile-height 16 \
//	    -cols 8 -rows 2 \
//	    -sprites volume_on,volume_off,gear,fullscreen,fullscreen_exit,scale
//
// Pass "-" or an empty entry in -sprites for blank cells. Indices are filled
// row-major (left-to-right, top-to-bottom).
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var asepriteCandidates = []string{
	"aseprite",
	"/Applications/Aseprite.app/Contents/MacOS/aseprite",
	"/Applications/Aseprite.app/Contents/Resources/aseprite",
}

func main() {
	var (
		name       = flag.String("name", "", "output basename, no extension (e.g. assets/sprites/overlay/icons)")
		tileWidth  = flag.Int("tile-width", 16, "tile width in pixels")
		tileHeight = flag.Int("tile-height", 16, "tile height in pixels")
		cols       = flag.Int("cols", 4, "columns")
		rows       = flag.Int("rows", 4, "rows")
		spritesArg = flag.String("sprites", "", "comma-separated sprite aliases in row-major order; use - or blank for empty cells")
		force      = flag.Bool("force", false, "overwrite existing files")
	)
	flag.Parse()

	if *name == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *cols <= 0 || *rows <= 0 || *tileWidth <= 0 || *tileHeight <= 0 {
		die(fmt.Errorf("dimensions must be positive"))
	}

	base := *name
	pngPath := base + ".png"
	yamlPath := base + ".yaml"
	asePath := base + ".aseprite"

	if !*force {
		for _, p := range []string{pngPath, yamlPath, asePath} {
			if _, err := os.Stat(p); err == nil {
				die(fmt.Errorf("%s already exists (pass -force to overwrite)", p))
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
		die(err)
	}

	if err := writePNG(pngPath, *cols**tileWidth, *rows**tileHeight); err != nil {
		die(err)
	}
	fmt.Println("wrote", pngPath)

	if err := writeYAML(yamlPath, *tileWidth, *tileHeight, parseSprites(*spritesArg, *cols)); err != nil {
		die(err)
	}
	fmt.Println("wrote", yamlPath)

	if ase := findAseprite(); ase != "" {
		if err := writeAseprite(ase, pngPath, asePath); err != nil {
			fmt.Fprintln(os.Stderr, "aseprite step failed:", err)
		} else {
			fmt.Println("wrote", asePath)
		}
	} else {
		fmt.Println("aseprite CLI not found; open", pngPath, "in Aseprite and save as", asePath, "manually if you want one")
	}
}

func writePNG(path string, w, h int) error {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	// Fill fully transparent. image.NewNRGBA zeros the buffer so we're done,
	// but be explicit for readers.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

type spriteEntry struct {
	Row    int `yaml:"row"`
	Column int `yaml:"column"`
}

type tilesheetBlock struct {
	TileWidth  int `yaml:"tileWidth"`
	TileHeight int `yaml:"tileHeight"`
}

type sidecar struct {
	Tilesheet tilesheetBlock         `yaml:"tilesheet"`
	Sprites   map[string]spriteEntry `yaml:"sprites,omitempty"`
}

func parseSprites(arg string, cols int) map[string]spriteEntry {
	if arg == "" {
		return nil
	}
	out := map[string]spriteEntry{}
	for i, raw := range strings.Split(arg, ",") {
		n := strings.TrimSpace(raw)
		if n == "" || n == "-" {
			continue
		}
		col := (i % cols) + 1
		row := (i / cols) + 1
		out[n] = spriteEntry{Row: row, Column: col}
	}
	return out
}

func writeYAML(path string, tw, th int, sprites map[string]spriteEntry) error {
	s := sidecar{
		Tilesheet: tilesheetBlock{TileWidth: tw, TileHeight: th},
		Sprites:   sprites,
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func findAseprite() string {
	for _, c := range asepriteCandidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func writeAseprite(asePath, pngPath, outPath string) error {
	cmd := exec.Command(asePath, "-b", pngPath, "--save-as", outPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "sprite_new:", err)
	os.Exit(1)
}
