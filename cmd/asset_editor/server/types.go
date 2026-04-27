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

type AudioEntry struct {
    Path         string   `json:"path"`
    Directory    string   `json:"directory"`
    Name         string   `json:"name"`
    Format       string   `json:"format"`
    FileSize     int64    `json:"fileSize"`
    Category     string   `json:"category"`
    HasYAML      bool     `json:"hasYaml"`
    Gain         *float64 `json:"gain,omitempty"`
    ResourceName string   `json:"resourceName"`
}

type AudioDetail struct {
    AudioEntry
    ResourceName string                `json:"resourceName"`
    Metadata     *schema.AudioMetadata `json:"metadata,omitempty"`
    RawYAML      string                `json:"rawYaml,omitempty"`
}

type ScriptFileEntry struct {
    Path              string   `json:"path"`
    Directory         string   `json:"directory"`
    Name              string   `json:"name"`
    HandlerCount      int      `json:"handlerCount"`
    SequenceCount     int      `json:"sequenceCount"`
    HandlerNames      []string `json:"handlerNames"`
    SequenceNames     []string `json:"sequenceNames,omitempty"`
    CustomActionNames []string `json:"customActionNames,omitempty"`
    ConstNames            []string `json:"constNames,omitempty"`
    PropertyTemplateNames []string `json:"propertyTemplateNames,omitempty"`
}

type ScriptFileDetail struct {
    ScriptFileEntry
    RawYAML string `json:"rawYaml"`
}

type PropertyTemplateEntry struct {
    Name       string         `json:"name"`
    Properties map[string]any `json:"properties"`
    ScriptFile string         `json:"scriptFile"`
}
