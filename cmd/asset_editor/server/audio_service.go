package server

import (
    "errors"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"

    "fisherevans.com/project/f/internal/schema"
    "gopkg.in/yaml.v3"
)

var audioExtensions = map[string]bool{
    ".wav":  true,
    ".mp3":  true,
    ".ogg":  true,
    ".flac": true,
}

type AudioService struct {
    assetsDir string
}

func NewAudioService(assetsDir string) *AudioService {
    return &AudioService{assetsDir: assetsDir}
}

func (s *AudioService) audioDir() string {
    return filepath.Join(s.assetsDir, "audio")
}

func (s *AudioService) ListAudio() ([]AudioEntry, error) {
    root := s.audioDir()
    var entries []AudioEntry

    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        if d.IsDir() {
            if d.Name() == "source_material" {
                return filepath.SkipDir
            }
            return nil
        }

        ext := strings.ToLower(filepath.Ext(d.Name()))
        if !audioExtensions[ext] {
            return nil
        }

        rel, _ := filepath.Rel(root, path)
        audioPath := strings.TrimSuffix(rel, filepath.Ext(rel))
        dir := filepath.Dir(audioPath)
        if dir == "." {
            dir = ""
        }

        info, _ := d.Info()
        var fileSize int64
        if info != nil {
            fileSize = info.Size()
        }

        yamlPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".yaml"
        _, yamlErr := os.Stat(yamlPath)
        hasYAML := yamlErr == nil

        var gain *float64
        if hasYAML {
            gain = readAudioGain(yamlPath)
        }

        category := detectAudioCategory(rel)

        entries = append(entries, AudioEntry{
            Path:         audioPath,
            Directory:    dir,
            Name:         filepath.Base(audioPath),
            Format:       strings.TrimPrefix(ext, "."),
            FileSize:     fileSize,
            Category:     category,
            HasYAML:      hasYAML,
            Gain:         gain,
            ResourceName: resourceNameForAudio(rel),
        })
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("walking audio: %w", err)
    }
    return entries, nil
}

func (s *AudioService) GetAudio(name string) (*AudioDetail, error) {
    root := s.audioDir()
    audioPath := findAudioFile(root, name)
    if audioPath == "" {
        return nil, fmt.Errorf("audio not found: %s", name)
    }

    ext := strings.ToLower(filepath.Ext(audioPath))
    info, _ := os.Stat(audioPath)
    var fileSize int64
    if info != nil {
        fileSize = info.Size()
    }

    rel, _ := filepath.Rel(root, audioPath)
    cleanPath := strings.TrimSuffix(rel, filepath.Ext(rel))
    dir := filepath.Dir(cleanPath)
    if dir == "." {
        dir = ""
    }

    yamlPath := strings.TrimSuffix(audioPath, filepath.Ext(audioPath)) + ".yaml"
    _, yamlErr := os.Stat(yamlPath)
    hasYAML := yamlErr == nil

    detail := &AudioDetail{
        AudioEntry: AudioEntry{
            Path:      cleanPath,
            Directory: dir,
            Name:      filepath.Base(cleanPath),
            Format:    strings.TrimPrefix(ext, "."),
            FileSize:  fileSize,
            Category:  detectAudioCategory(rel),
            HasYAML:   hasYAML,
        },
    }

    resourceName := resourceNameForAudio(rel)
    detail.ResourceName = resourceName

    if hasYAML {
        data, err := os.ReadFile(yamlPath)
        if err != nil {
            return nil, fmt.Errorf("reading sidecar: %w", err)
        }
        var meta schema.AudioMetadata
        if err := yaml.Unmarshal(data, &meta); err != nil {
            return nil, fmt.Errorf("parsing sidecar: %w", err)
        }
        detail.Metadata = &meta
        detail.RawYAML = string(data)
        detail.Gain = meta.Gain
    }

    return detail, nil
}

func (s *AudioService) SaveAudio(name string, meta *schema.AudioMetadata) error {
    root := s.audioDir()
    audioPath := findAudioFile(root, name)
    if audioPath == "" {
        return fmt.Errorf("audio not found: %s", name)
    }

    yamlPath := strings.TrimSuffix(audioPath, filepath.Ext(audioPath)) + ".yaml"

    if meta.Gain == nil {
        err := os.Remove(yamlPath)
        if errors.Is(err, fs.ErrNotExist) {
            return nil
        }
        return err
    }

    dir := filepath.Dir(yamlPath)
    tmp, err := os.CreateTemp(dir, ".audio-*.yaml")
    if err != nil {
        return fmt.Errorf("creating temp file: %w", err)
    }
    tmpName := tmp.Name()
    defer func() {
        _ = tmp.Close()
        _ = os.Remove(tmpName)
    }()

    enc := yaml.NewEncoder(tmp)
    enc.SetIndent(2)
    if err := enc.Encode(meta); err != nil {
        return fmt.Errorf("marshaling yaml: %w", err)
    }
    if err := enc.Close(); err != nil {
        return fmt.Errorf("closing encoder: %w", err)
    }
    if err := tmp.Close(); err != nil {
        return fmt.Errorf("closing temp file: %w", err)
    }
    return os.Rename(tmpName, yamlPath)
}

func (s *AudioService) AudioFilePath(name string) string {
    return findAudioFile(s.audioDir(), name)
}

func findAudioFile(root, name string) string {
    if _, err := safePath(root, name); err != nil {
        return ""
    }
    for ext := range audioExtensions {
        path := filepath.Join(root, name+ext)
        if _, err := os.Stat(path); err == nil {
            return path
        }
    }
    return ""
}

func detectAudioCategory(rel string) string {
    parts := strings.SplitN(rel, string(filepath.Separator), 2)
    if len(parts) == 0 {
        return ""
    }
    return parts[0]
}

func resourceNameForAudio(rel string) string {
    category := detectAudioCategory(rel)
    if category == "sounds" {
        return strings.TrimPrefix(strings.TrimSuffix(rel, filepath.Ext(rel)), "sounds/")
    }
    return "audio/" + rel
}

func readAudioGain(yamlPath string) *float64 {
    data, err := os.ReadFile(yamlPath)
    if err != nil {
        return nil
    }
    var meta schema.AudioMetadata
    if err := yaml.Unmarshal(data, &meta); err != nil {
        return nil
    }
    return meta.Gain
}
