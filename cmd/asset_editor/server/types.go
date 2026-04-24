package server

import "fisherevans.com/project/f/internal/schema"

type SpriteEntry struct {
    Path        string `json:"path"`
    Directory   string `json:"directory"`
    Name        string `json:"name"`
    Type        string `json:"type"`
    ImageWidth  int    `json:"imageWidth"`
    ImageHeight int    `json:"imageHeight"`
    HasYAML     bool   `json:"hasYaml"`
}

type SpriteDetail struct {
    SpriteEntry
    Metadata *schema.SpriteMetadata `json:"metadata,omitempty"`
    Computed *ComputedFields        `json:"computed,omitempty"`
    RawYAML  string                 `json:"rawYaml,omitempty"`
}

type ComputedFields struct {
    Columns int `json:"columns,omitempty"`
    Rows    int `json:"rows,omitempty"`
}

type ScaffoldRequest struct {
    Name       string   `json:"name"`
    TileWidth  int      `json:"tileWidth"`
    TileHeight int      `json:"tileHeight"`
    Cols       int      `json:"cols"`
    Rows       int      `json:"rows"`
    Sprites    []string `json:"sprites"`
    Force      bool     `json:"force"`
}
