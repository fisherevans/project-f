export interface StateInfo {
    type: string;
    capabilities: string[];
    map_name?: string;
    player_position?: { x: number; y: number };
}

export interface GlobalEntry {
    key: string;
    value: unknown;
    exists: boolean;
}

export interface EntitySnapshot {
    id: string;
    x: number;
    y: number;
    is_moving: boolean;
    direction: string;
    is_player?: boolean;
    debug_type?: string;
    handler_ref?: string;
}

export interface EntityListResponse {
    entities: EntitySnapshot[];
    map_width: number;
    map_height: number;
}

export interface EntityDetail extends EntitySnapshot {
    precise_x: number;
    precise_y: number;
    behavior_enabled: boolean;
    has_behavior: boolean;
    has_presence: boolean;
    has_renderer: boolean;
    sound_enabled: boolean;
    metadata?: unknown;
}

export interface TeleportEntry {
    reference: string;
    destination: string;
    x: number;
    y: number;
    exit_direction: string;
}

export interface ZoneRect {
    id: string;
    x: number;
    y: number;
    w: number;
    h: number;
}

export interface CommandResponse {
    output: string[];
}
