// Command gamectl drives a running development build of the game through its
// debug HTTP API (see cmd/development/debugapi). It is the ergonomic front end
// for the agent-driving harness: boot status, frame capture, synthetic input,
// time stepping, content reload, and state inspection.
//
// The dev build must be running with the debug API enabled (the default for
// `go run ./cmd/development`). Set GAMECTL_ADDR or pass -addr to point at a
// non-default host.
//
// Examples:
//
//	gamectl health
//	gamectl shot frame.png            # capture the 240x160 scene
//	gamectl shot -layer scaled out.png
//	gamectl pause
//	gamectl step 30 -shot s.png       # advance 30 fixed-dt frames, then capture
//	gamectl move Up -frames 40 -shot s.png
//	gamectl tap A
//	gamectl cmd "tp 10 12"
//	gamectl reset
//	gamectl reload scripts
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "io"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

func main() {
    if len(os.Args) < 2 {
        usage()
        os.Exit(2)
    }
    cmd := os.Args[1]
    args := os.Args[2:]
    var err error
    switch cmd {
    case "health":
        err = cmdHealth(args)
    case "state":
        err = get("/api/v1/debug/state")
    case "entities":
        err = cmdEntities(args)
    case "combat":
        err = get("/api/v1/debug/combat")
    case "save":
        err = cmdSave(args)
    case "shot", "screenshot":
        err = cmdShot(args)
    case "compare":
        err = cmdCompare(args)
    case "pause":
        err = postJSON("/api/v1/debug/time", map[string]any{"mode": "pause"})
    case "run", "resume":
        err = cmdRun(args)
    case "step":
        err = cmdStep(args)
    case "determinism":
        err = cmdDeterminism(args)
    case "move":
        err = cmdMove(args)
    case "tap":
        err = cmdTap(args)
    case "input":
        err = cmdInput(args)
    case "clearinput":
        err = del("/api/v1/debug/input")
    case "reset":
        err = postJSON("/api/v1/debug/reset", map[string]any{})
    case "reload":
        err = cmdReload(args)
    case "cmd":
        err = cmdConsole(args)
    case "tp", "teleport":
        err = cmdTeleport(args)
    case "map":
        err = cmdMap(args)
    case "wait":
        err = cmdWait(args)
    case "help", "-h", "--help":
        usage()
    default:
        fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
        usage()
        os.Exit(2)
    }
    if err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        os.Exit(1)
    }
}

func usage() {
    fmt.Print(`gamectl - drive a running dev build via the debug API

Status:
  health                       readiness, active state, frame counter
  state                        active state detail (map, player pos)
  entities [-player]           list entities (or just the player)
  combat                       combat state (sync/shield/statuses/phase) when in combat
  save [load <id>]             dump the active save, or swap to save <id> and reset
  wait [-timeout 60]           block until the API reports ready

Capture:
  shot [-layer scene|scaled] [-compare baseline.png] [-tol 0.0] [path]   capture a PNG (optionally diff vs a baseline)
  compare <baseline.png> <candidate.png> [-tol 0.0]   pixel-diff two PNGs, write a diff image

Time:
  pause                        freeze game logic (rendering continues)
  run [-speed 1.0]             resume game logic
  step [N] [-dt 0.0167] [-shot path]   advance N fixed-dt frames (blocking), optional capture
  determinism [on|off] [-seed N]       seed the shared RNG for reproducible frames (reapplied on reset)

Input (synthetic controller):
  tap A|B|Start|Select [-frames 1]     momentary press
  move Up|Down|Left|Right [-frames 30] [-shot path]   hold a direction while stepping
  input [-a] [-b] [-start] [-select] [-dir Up] [-frames 1]
  clearinput                   cancel any held injected input

State control:
  reset                        re-issue the boot intent (fresh state)
  reload scripts|rpg|overlays  hot-reload content from disk
  cmd "<console input>"        run a console command, print output
  tp <x> <y> | <target>        teleport player
  map <name> [waypoint]        load a map

Global flags:
  -addr URL   debug API base (default http://localhost:8091, or $GAMECTL_ADDR)
`)
}

// --- HTTP helpers ---

func baseURL() string {
    if v := os.Getenv("GAMECTL_ADDR"); v != "" {
        return v
    }
    return "http://localhost:8091"
}

// popAddr extracts a -addr flag from anywhere in args, returning the cleaned
// args. Keeps per-command flag parsing simple.
func popAddr(args []string) []string {
    out := args[:0:0]
    for i := 0; i < len(args); i++ {
        if args[i] == "-addr" || args[i] == "--addr" {
            if i+1 < len(args) {
                os.Setenv("GAMECTL_ADDR", args[i+1])
                i++
            }
            continue
        }
        out = append(out, args[i])
    }
    return out
}

func get(path string) error {
    resp, err := http.Get(baseURL() + path)
    if err != nil {
        return err
    }
    return dump(resp)
}

func del(path string) error {
    req, _ := http.NewRequest(http.MethodDelete, baseURL()+path, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    return dump(resp)
}

func postJSON(path string, body any) error {
    b, _ := json.Marshal(body)
    resp, err := http.Post(baseURL()+path, "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    return dump(resp)
}

// postQuiet posts without printing the response on success. Used to orchestrate
// multi-step convenience commands (move, step) without flooding stdout.
func postQuiet(path string, body any) error {
    b, _ := json.Marshal(body)
    resp, err := http.Post(baseURL()+path, "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        rb, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("[http %d] %s", resp.StatusCode, string(rb))
    }
    io.Copy(io.Discard, resp.Body)
    return nil
}

func dump(resp *http.Response) error {
    defer resp.Body.Close()
    b, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 400 {
        fmt.Fprintf(os.Stderr, "[http %d] ", resp.StatusCode)
    }
    // Pretty-print JSON when possible.
    var pretty bytes.Buffer
    if json.Indent(&pretty, b, "", "  ") == nil {
        fmt.Println(pretty.String())
    } else {
        fmt.Println(string(b))
    }
    if resp.StatusCode >= 400 {
        return fmt.Errorf("request failed")
    }
    return nil
}

// --- flag parsing (minimal) ---

type flags struct {
    bools   map[string]bool
    strs    map[string]string
    ints    map[string]int
    floats  map[string]float64
    pos     []string
}

// parse pulls -key value / -bool flags out of args. boolKeys are flags with no
// value. Everything else is positional.
func parse(args []string, boolKeys ...string) *flags {
    f := &flags{bools: map[string]bool{}, strs: map[string]string{}, ints: map[string]int{}, floats: map[string]float64{}}
    isBool := map[string]bool{}
    for _, k := range boolKeys {
        isBool[k] = true
    }
    for i := 0; i < len(args); i++ {
        a := args[i]
        if strings.HasPrefix(a, "-") {
            key := strings.TrimLeft(a, "-")
            if isBool[key] {
                f.bools[key] = true
                continue
            }
            if i+1 < len(args) {
                v := args[i+1]
                f.strs[key] = v
                if n, err := strconv.Atoi(v); err == nil {
                    f.ints[key] = n
                }
                if fl, err := strconv.ParseFloat(v, 64); err == nil {
                    f.floats[key] = fl
                }
                i++
            }
            continue
        }
        f.pos = append(f.pos, a)
    }
    return f
}

// --- commands ---

func cmdHealth(args []string) error {
    popAddr(args)
    return get("/api/v1/debug/health")
}

func cmdEntities(args []string) error {
    f := parse(popAddr(args), "player")
    resp, err := http.Get(baseURL() + "/api/v1/debug/entities")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if !f.bools["player"] {
        return dump(resp)
    }
    var data struct {
        Entities []map[string]any `json:"entities"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return err
    }
    for _, e := range data.Entities {
        if e["is_player"] == true {
            b, _ := json.MarshalIndent(e, "", "  ")
            fmt.Println(string(b))
            return nil
        }
    }
    return fmt.Errorf("no player entity found")
}

func cmdRun(args []string) error {
    f := parse(popAddr(args))
    body := map[string]any{"mode": "run"}
    if s, ok := f.floats["speed"]; ok {
        body["speed"] = s
    }
    return postJSON("/api/v1/debug/time", body)
}

func cmdShot(args []string) error {
    f := parse(popAddr(args))
    layer := f.strs["layer"]
    if layer == "" {
        layer = "scene"
    }
    path := "/tmp/gamectl-shot-" + time.Now().Format("150405") + ".png"
    if len(f.pos) > 0 {
        path = f.pos[0]
    }
    if err := capture(layer, path); err != nil {
        return err
    }
    if baseline := f.strs["compare"]; baseline != "" {
        tol := f.floats["tol"]
        return compareImages(baseline, path, tol)
    }
    return nil
}

func cmdCompare(args []string) error {
    f := parse(popAddr(args))
    if len(f.pos) < 2 {
        return fmt.Errorf("usage: gamectl compare <baseline.png> <candidate.png> [-tol 0.0]")
    }
    return compareImages(f.pos[0], f.pos[1], f.floats["tol"])
}

// compareImages reports the fraction of differing pixels between two PNGs,
// writes a diff image (differing pixels in magenta) next to the candidate, and
// returns a non-nil error when the diff fraction exceeds tol. A tol of 0
// requires a byte-for-byte pixel match.
func compareImages(baselinePath, candidatePath string, tol float64) error {
    a, err := loadPNG(baselinePath)
    if err != nil {
        return fmt.Errorf("baseline: %w", err)
    }
    b, err := loadPNG(candidatePath)
    if err != nil {
        return fmt.Errorf("candidate: %w", err)
    }
    ab, bb := a.Bounds(), b.Bounds()
    if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
        return fmt.Errorf("dimension mismatch: baseline %dx%d vs candidate %dx%d", ab.Dx(), ab.Dy(), bb.Dx(), bb.Dy())
    }
    diff := image.NewRGBA(ab)
    var differing int
    total := ab.Dx() * ab.Dy()
    for y := ab.Min.Y; y < ab.Max.Y; y++ {
        for x := ab.Min.X; x < ab.Max.X; x++ {
            ar, ag, abl, aa := a.At(x, y).RGBA()
            br, bg, bbl, ba := b.At(x, y).RGBA()
            if ar != br || ag != bg || abl != bbl || aa != ba {
                differing++
                diff.Set(x, y, color.RGBA{R: 255, G: 0, B: 255, A: 255})
            } else {
                diff.Set(x, y, color.RGBA{R: uint8(ar >> 8), G: uint8(ag >> 8), B: uint8(abl >> 8), A: 64})
            }
        }
    }
    frac := float64(differing) / float64(total)
    diffPath := candidatePath + ".diff.png"
    if f, ferr := os.Create(diffPath); ferr == nil {
        png.Encode(f, diff)
        f.Close()
    }
    fmt.Printf("diff: %.4f%% (%d/%d pixels), diff image: %s\n", frac*100, differing, total, diffPath)
    if frac > tol {
        return fmt.Errorf("diff %.4f exceeds tolerance %.4f", frac, tol)
    }
    return nil
}

func loadPNG(path string) (image.Image, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    return png.Decode(f)
}

func capture(layer, path string) error {
    resp, err := http.Get(baseURL() + "/api/v1/debug/screenshot?layer=" + layer)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("capture failed [%d]: %s", resp.StatusCode, string(b))
    }
    b, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    if err := os.WriteFile(path, b, 0644); err != nil {
        return err
    }
    fmt.Println(path)
    return nil
}

func cmdStep(args []string) error {
    f := parse(popAddr(args))
    frames := 1
    if len(f.pos) > 0 {
        if n, err := strconv.Atoi(f.pos[0]); err == nil {
            frames = n
        }
    } else if n, ok := f.ints["frames"]; ok {
        frames = n
    }
    body := map[string]any{"frames": frames}
    if dt, ok := f.floats["dt"]; ok {
        body["dt"] = dt
    }
    if err := postJSON("/api/v1/debug/step", body); err != nil {
        return err
    }
    if shot := f.strs["shot"]; shot != "" {
        return capture("scene", shot)
    }
    return nil
}

func cmdDeterminism(args []string) error {
    f := parse(popAddr(args))
    on := true
    if len(f.pos) > 0 && (f.pos[0] == "off" || f.pos[0] == "false") {
        on = false
    }
    body := map[string]any{"enabled": on}
    if seed, ok := f.ints["seed"]; ok {
        body["seed"] = seed
    }
    return postJSON("/api/v1/debug/determinism", body)
}

func cmdMove(args []string) error {
    f := parse(popAddr(args))
    if len(f.pos) == 0 {
        return fmt.Errorf("usage: gamectl move Up|Down|Left|Right [-frames N] [-shot path]")
    }
    dir := f.pos[0]
    frames := 30
    if n, ok := f.ints["frames"]; ok {
        frames = n
    }
    // Ensure paused so stepping is deterministic.
    if err := postQuiet("/api/v1/debug/time", map[string]any{"mode": "pause"}); err != nil {
        return err
    }
    if err := postQuiet("/api/v1/debug/input", map[string]any{"dir": dir, "frames": frames}); err != nil {
        return err
    }
    if err := postQuiet("/api/v1/debug/step", map[string]any{"frames": frames}); err != nil {
        return err
    }
    fmt.Printf("moved %s (%d frames)\n", dir, frames)
    if shot := f.strs["shot"]; shot != "" {
        return capture("scene", shot)
    }
    return nil
}

func cmdTap(args []string) error {
    f := parse(popAddr(args))
    if len(f.pos) == 0 {
        return fmt.Errorf("usage: gamectl tap A|B|Start|Select [-frames N]")
    }
    frames := 1
    if n, ok := f.ints["frames"]; ok {
        frames = n
    }
    body := map[string]any{"frames": frames}
    switch strings.ToLower(f.pos[0]) {
    case "a":
        body["a"] = true
    case "b":
        body["b"] = true
    case "start":
        body["start"] = true
    case "select":
        body["select"] = true
    default:
        return fmt.Errorf("unknown button %q (A|B|Start|Select)", f.pos[0])
    }
    return postJSON("/api/v1/debug/input", body)
}

func cmdInput(args []string) error {
    f := parse(popAddr(args), "a", "b", "start", "select")
    body := map[string]any{}
    if f.bools["a"] {
        body["a"] = true
    }
    if f.bools["b"] {
        body["b"] = true
    }
    if f.bools["start"] {
        body["start"] = true
    }
    if f.bools["select"] {
        body["select"] = true
    }
    if d := f.strs["dir"]; d != "" {
        body["dir"] = d
    }
    if n, ok := f.ints["frames"]; ok {
        body["frames"] = n
    }
    return postJSON("/api/v1/debug/input", body)
}

func cmdSave(args []string) error {
    a := popAddr(args)
    if len(a) >= 2 && a[0] == "load" {
        return postJSON("/api/v1/debug/save/load", map[string]any{"id": a[1]})
    }
    // default: dump the active save
    return get("/api/v1/debug/save")
}

func cmdReload(args []string) error {
    f := parse(popAddr(args))
    if len(f.pos) == 0 {
        return fmt.Errorf("usage: gamectl reload scripts|rpg|overlays")
    }
    return postJSON("/api/v1/debug/reload/"+f.pos[0], map[string]any{})
}

func cmdConsole(args []string) error {
    a := popAddr(args)
    if len(a) == 0 {
        return fmt.Errorf("usage: gamectl cmd \"<console input>\"")
    }
    return postJSON("/api/v1/debug/command", map[string]any{"command": strings.Join(a, " ")})
}

func cmdTeleport(args []string) error {
    a := popAddr(args)
    if len(a) == 2 {
        x, ex := strconv.Atoi(a[0])
        y, ey := strconv.Atoi(a[1])
        if ex == nil && ey == nil {
            return postJSON("/api/v1/debug/teleport", map[string]any{"x": x, "y": y})
        }
    }
    if len(a) == 1 {
        return postJSON("/api/v1/debug/teleport", map[string]any{"target": a[0]})
    }
    return fmt.Errorf("usage: gamectl tp <x> <y> | <target>")
}

func cmdMap(args []string) error {
    a := popAddr(args)
    if len(a) == 0 {
        return fmt.Errorf("usage: gamectl map <name> [waypoint]")
    }
    body := map[string]any{"name": a[0]}
    if len(a) > 1 {
        body["waypoint"] = a[1]
    }
    return postJSON("/api/v1/debug/map", body)
}

func cmdWait(args []string) error {
    f := parse(popAddr(args))
    timeout := 60
    if n, ok := f.ints["timeout"]; ok {
        timeout = n
    }
    deadline := time.Now().Add(time.Duration(timeout) * time.Second)
    for time.Now().Before(deadline) {
        resp, err := http.Get(baseURL() + "/api/v1/debug/health")
        if err == nil {
            var h struct {
                Ready bool   `json:"ready"`
                State string `json:"state"`
            }
            json.NewDecoder(resp.Body).Decode(&h)
            resp.Body.Close()
            if h.Ready {
                fmt.Printf("ready: %s\n", h.State)
                return nil
            }
        }
        time.Sleep(500 * time.Millisecond)
    }
    return fmt.Errorf("timed out after %ds waiting for ready", timeout)
}
