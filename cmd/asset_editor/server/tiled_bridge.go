package server

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

// TiledBridge manages real-time communication between the Tiled map editor
// plugin and the companion web page. The plugin pushes selection state and
// polls for commands; the companion page subscribes via WebSocket and
// enqueues commands.
type TiledBridge struct {
	mu            sync.Mutex
	selection     *TiledSelection
	commands      []TiledCommand
	listeners     []chan TiledBridgeEvent
	lastHeartbeat time.Time
}

type TiledSelection struct {
	MapFile string               `json:"mapFile"`
	Objects []TiledSelectedObject `json:"objects"`
	Time    time.Time            `json:"time"`
	Acks    []TiledCommandAck    `json:"acks,omitempty"`
}

type TiledSelectedObject struct {
	ID         int               `json:"id"`
	Name       string            `json:"name,omitempty"`
	ClassName  string            `json:"className,omitempty"`
	X          float64           `json:"x"`
	Y          float64           `json:"y"`
	Width      float64           `json:"width,omitempty"`
	Height     float64           `json:"height,omitempty"`
	Properties map[string]any    `json:"properties"`
}

type TiledCommand struct {
	ID        string         `json:"id"`
	ObjectID  int            `json:"objectId"`
	Action    string         `json:"action"`
	Name      string         `json:"name,omitempty"`
	Value     any            `json:"value,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type TiledCommandAck struct {
	ID      string `json:"id"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

type TiledBridgeEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func NewTiledBridge() *TiledBridge {
	return &TiledBridge{}
}

func (tb *TiledBridge) SetSelection(sel *TiledSelection) []TiledCommand {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	sel.Time = time.Now()
	tb.selection = sel
	tb.lastHeartbeat = sel.Time
	event := TiledBridgeEvent{Type: "tiled-selection", Data: sel}
	for _, ch := range tb.listeners {
		select {
		case ch <- event:
		default:
		}
	}
	cmds := tb.commands
	tb.commands = nil
	return cmds
}

func (tb *TiledBridge) GetSelection() *TiledSelection {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.selection
}

func (tb *TiledBridge) EnqueueCommand(cmd TiledCommand) {
	tb.mu.Lock()
	cmd.Timestamp = time.Now()
	if cmd.ID == "" {
		cmd.ID = time.Now().Format("20060102150405.000")
	}
	tb.commands = append(tb.commands, cmd)
	tb.mu.Unlock()
}

func (tb *TiledBridge) DrainCommands() []TiledCommand {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	cmds := tb.commands
	tb.commands = nil
	return cmds
}

func (tb *TiledBridge) Heartbeat() []TiledCommand {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.lastHeartbeat = time.Now()
	cmds := tb.commands
	tb.commands = nil
	return cmds
}

func (tb *TiledBridge) AckCommands(acks []TiledCommandAck) {
	tb.mu.Lock()
	event := TiledBridgeEvent{Type: "tiled-command-ack", Data: acks}
	for _, ch := range tb.listeners {
		select {
		case ch <- event:
		default:
		}
	}
	tb.mu.Unlock()
}

func (tb *TiledBridge) IsConnected() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return !tb.lastHeartbeat.IsZero() && time.Since(tb.lastHeartbeat) < 10*time.Second
}

func (tb *TiledBridge) Subscribe() chan TiledBridgeEvent {
	ch := make(chan TiledBridgeEvent, 16)
	tb.mu.Lock()
	tb.listeners = append(tb.listeners, ch)
	tb.mu.Unlock()
	return ch
}

func (tb *TiledBridge) Unsubscribe(ch chan TiledBridgeEvent) {
	tb.mu.Lock()
	for i, l := range tb.listeners {
		if l == ch {
			tb.listeners = append(tb.listeners[:i], tb.listeners[i+1:]...)
			break
		}
	}
	tb.mu.Unlock()
	close(ch)
}

// HTTP handlers

func (s *Server) handleTiledBridgeSelection(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var sel TiledSelection
	if err := json.Unmarshal(body, &sel); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	acks := sel.Acks
	sel.Acks = nil
	if len(acks) > 0 {
		s.tiledBridge.AckCommands(acks)
	}
	cmds := s.tiledBridge.SetSelection(&sel)
	if cmds == nil {
		cmds = []TiledCommand{}
	}
	writeJSON(w, cmds)
}

func (s *Server) handleTiledBridgeGetSelection(w http.ResponseWriter, r *http.Request) {
	sel := s.tiledBridge.GetSelection()
	if sel == nil {
		writeJSON(w, map[string]any{"mapFile": "", "objects": []any{}})
		return
	}
	writeJSON(w, sel)
}

func (s *Server) handleTiledBridgeCommands(w http.ResponseWriter, r *http.Request) {
	cmds := s.tiledBridge.DrainCommands()
	if cmds == nil {
		cmds = []TiledCommand{}
	}
	writeJSON(w, cmds)
}

func (s *Server) handleTiledBridgeEnqueueCommand(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var cmd TiledCommand
	if err := json.Unmarshal(body, &cmd); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	s.tiledBridge.EnqueueCommand(cmd)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTiledBridgeHeartbeat(w http.ResponseWriter, r *http.Request) {
	cmds := s.tiledBridge.Heartbeat()
	if cmds == nil {
		cmds = []TiledCommand{}
	}
	writeJSON(w, cmds)
}

func (s *Server) handleTiledBridgeAck(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var acks []TiledCommandAck
	if err := json.Unmarshal(body, &acks); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	s.tiledBridge.AckCommands(acks)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTiledBridgeStatus(w http.ResponseWriter, r *http.Request) {
	sel := s.tiledBridge.GetSelection()
	writeJSON(w, map[string]any{
		"connected": s.tiledBridge.IsConnected(),
		"lastSeen":  sel,
	})
}
