package debugapi

type StateInfo struct {
    Type         string   `json:"type"`
    Capabilities []string `json:"capabilities"`
    MapName      string   `json:"map_name,omitempty"`
    PlayerPos    *Vec2    `json:"player_position,omitempty"`
}

type Vec2 struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type GlobalEntry struct {
    Key    string `json:"key"`
    Value  any    `json:"value"`
    Exists bool   `json:"exists"`
}

type EntitySnapshot struct {
    Id         string `json:"id"`
    X          int    `json:"x"`
    Y          int    `json:"y"`
    IsMoving   bool   `json:"is_moving"`
    Direction  string `json:"direction"`
    IsPlayer   bool   `json:"is_player,omitempty"`
    DebugType  string `json:"debug_type,omitempty"`
    HandlerRef string `json:"handler_ref,omitempty"`
}

type EntityListResponse struct {
    Entities []EntitySnapshot `json:"entities"`
    MapWidth int              `json:"map_width"`
    MapHeight int             `json:"map_height"`
}

type EntityDetail struct {
    EntitySnapshot
    PreciseX        float64  `json:"precise_x"`
    PreciseY        float64  `json:"precise_y"`
    BehaviorEnabled bool     `json:"behavior_enabled"`
    HasBehavior     bool     `json:"has_behavior"`
    HasPresence     bool     `json:"has_presence"`
    HasRenderer     bool     `json:"has_renderer"`
    SoundEnabled    bool     `json:"sound_enabled"`
    Metadata        any      `json:"metadata,omitempty"`
}

type TeleportEntry struct {
    Reference     string `json:"reference"`
    DestinationRef string `json:"destination"`
    X             int    `json:"x"`
    Y             int    `json:"y"`
    ExitDirection string `json:"exit_direction"`
}

type ZoneRect struct {
    Id string `json:"id"`
    X  int    `json:"x"`
    Y  int    `json:"y"`
    W  int    `json:"w"`
    H  int    `json:"h"`
}

type CommandRequest struct {
    Command string `json:"command"`
}

type CommandResponse struct {
    Output []string `json:"output"`
}

type TeleportRequest struct {
    Target string  `json:"target,omitempty"`
    X      *int    `json:"x,omitempty"`
    Y      *int    `json:"y,omitempty"`
}

type MapRequest struct {
    Name     string `json:"name"`
    Waypoint string `json:"waypoint,omitempty"`
}

type SetGlobalRequest struct {
    Type  string `json:"type"`
    Value any    `json:"value"`
}