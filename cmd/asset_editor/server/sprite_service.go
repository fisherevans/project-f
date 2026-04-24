package server

import (
    "errors"
    "fmt"
    "image"
    _ "image/png"
    "io/fs"
    "os"
    "path/filepath"
    "strings"

    "fisherevans.com/project/f/internal/schema"
    "gopkg.in/yaml.v3"
)

type SpriteService struct {
    assetsDir string
}

func NewSpriteService(assetsDir string) *SpriteService {
    return &SpriteService{assetsDir: assetsDir}
}

func (s *SpriteService) spritesDir() string {
    return filepath.Join(s.assetsDir, "sprites")
}

func (s *SpriteService) ListSprites() ([]SpriteEntry, error) {
    root := s.spritesDir()
    var entries []SpriteEntry

    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".png") {
            return nil
        }

        rel, _ := filepath.Rel(root, path)
        spritePath := strings.TrimSuffix(rel, filepath.Ext(rel))
        dir := filepath.Dir(spritePath)
        if dir == "." {
            dir = ""
        }
        name := filepath.Base(spritePath)

        w, h := imageSize(path)

        yamlPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".yaml"
        _, yamlErr := os.Stat(yamlPath)
        hasYAML := yamlErr == nil

        spriteType := "plain"
        if hasYAML {
            spriteType = detectSpriteType(yamlPath)
        }

        entries = append(entries, SpriteEntry{
            Path:        spritePath,
            Directory:   dir,
            Name:        name,
            Type:        spriteType,
            ImageWidth:  w,
            ImageHeight: h,
            HasYAML:     hasYAML,
        })
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("walking sprites: %w", err)
    }
    return entries, nil
}

func (s *SpriteService) GetSprite(name string) (*SpriteDetail, error) {
    root := s.spritesDir()
    pngPath := filepath.Join(root, name+".png")

    if _, err := os.Stat(pngPath); err != nil {
        return nil, fmt.Errorf("sprite not found: %s", name)
    }

    w, h := imageSize(pngPath)
    dir := filepath.Dir(name)
    if dir == "." {
        dir = ""
    }

    yamlPath := strings.TrimSuffix(pngPath, ".png") + ".yaml"
    _, yamlErr := os.Stat(yamlPath)
    hasYAML := yamlErr == nil

    spriteType := "plain"
    detail := &SpriteDetail{
        SpriteEntry: SpriteEntry{
            Path:        name,
            Directory:   dir,
            Name:        filepath.Base(name),
            ImageWidth:  w,
            ImageHeight: h,
            HasYAML:     hasYAML,
        },
    }

    if hasYAML {
        data, err := os.ReadFile(yamlPath)
        if err != nil {
            return nil, fmt.Errorf("reading sidecar: %w", err)
        }
        var meta schema.SpriteMetadata
        if err := yaml.Unmarshal(data, &meta); err != nil {
            return nil, fmt.Errorf("parsing sidecar: %w", err)
        }
        detail.Metadata = &meta
        detail.RawYAML = string(data)

        if meta.NonAtlasSprite {
            spriteType = "nonAtlas"
        } else if meta.Frame != nil {
            spriteType = "frame"
        } else if meta.Tilesheet != nil {
            spriteType = "tilesheet"
        }

        if meta.Tilesheet != nil && w > 0 && h > 0 {
            detail.Computed = &ComputedFields{
                Columns: w / meta.Tilesheet.TileWidth.Int(),
                Rows:    h / meta.Tilesheet.TileHeight.Int(),
            }
        }
    }

    detail.Type = spriteType
    return detail, nil
}

func (s *SpriteService) SaveSprite(name string, meta *schema.SpriteMetadata) error {
    root := s.spritesDir()
    yamlPath := filepath.Join(root, name+".yaml")

    dir := filepath.Dir(yamlPath)
    tmp, err := os.CreateTemp(dir, ".sprite-*.yaml")
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

func (s *SpriteService) DeleteSpriteSidecar(name string) error {
    root := s.spritesDir()
    yamlPath := filepath.Join(root, name+".yaml")
    err := os.Remove(yamlPath)
    if errors.Is(err, fs.ErrNotExist) {
        return nil
    }
    return err
}

func detectSpriteType(yamlPath string) string {
    data, err := os.ReadFile(yamlPath)
    if err != nil {
        return "plain"
    }
    var meta schema.SpriteMetadata
    if err := yaml.Unmarshal(data, &meta); err != nil {
        return "plain"
    }
    if meta.NonAtlasSprite {
        return "nonAtlas"
    }
    if meta.Frame != nil {
        return "frame"
    }
    if meta.Tilesheet != nil {
        return "tilesheet"
    }
    return "plain"
}

func imageSize(path string) (int, int) {
    f, err := os.Open(path)
    if err != nil {
        return 0, 0
    }
    defer f.Close()

    cfg, _, err := image.DecodeConfig(f)
    if err != nil {
        return 0, 0
    }
    return cfg.Width, cfg.Height
}
