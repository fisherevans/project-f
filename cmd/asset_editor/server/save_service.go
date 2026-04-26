package server

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
)

type SaveService struct {
    savesDir string
}

func NewSaveService(assetsDir string) *SaveService {
    projectRoot := filepath.Dir(assetsDir)
    return &SaveService{savesDir: filepath.Join(projectRoot, "game_data", "saves")}
}

type SaveSummary struct {
    SaveId        string `json:"save_id"`
    CharacterName string `json:"character_name"`
}

type SaveDetail struct {
    SaveSummary
    RawYaml string `json:"raw_yaml"`
}

func (s *SaveService) ListSaves() ([]SaveSummary, error) {
    entries, err := os.ReadDir(s.savesDir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, err
    }
    var saves []SaveSummary
    for _, e := range entries {
        if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
            continue
        }
        id := strings.TrimSuffix(e.Name(), ".yaml")
        data, err := os.ReadFile(filepath.Join(s.savesDir, e.Name()))
        if err != nil {
            continue
        }
        var partial struct {
            SaveId        string `yaml:"save_id"`
            CharacterName string `yaml:"character_name"`
        }
        if err := yaml.Unmarshal(data, &partial); err != nil {
            continue
        }
        if partial.SaveId == "" {
            partial.SaveId = id
        }
        saves = append(saves, SaveSummary{
            SaveId:        partial.SaveId,
            CharacterName: partial.CharacterName,
        })
    }
    return saves, nil
}

func (s *SaveService) GetSave(id string) (*SaveDetail, error) {
    if _, err := safePath(s.savesDir, id); err != nil {
        return nil, err
    }
    path := filepath.Join(s.savesDir, id+".yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var partial struct {
        SaveId        string `yaml:"save_id"`
        CharacterName string `yaml:"character_name"`
    }
    yaml.Unmarshal(data, &partial)
    return &SaveDetail{
        SaveSummary: SaveSummary{
            SaveId:        partial.SaveId,
            CharacterName: partial.CharacterName,
        },
        RawYaml: string(data),
    }, nil
}

func (s *SaveService) SaveRaw(id string, rawYaml string) error {
    if _, err := safePath(s.savesDir, id); err != nil {
        return err
    }
    var check map[string]any
    if err := yaml.Unmarshal([]byte(rawYaml), &check); err != nil {
        return fmt.Errorf("invalid YAML: %w", err)
    }
    if err := os.MkdirAll(s.savesDir, 0o755); err != nil {
        return err
    }
    path := filepath.Join(s.savesDir, id+".yaml")
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, []byte(rawYaml), 0o644); err != nil {
        return err
    }
    return os.Rename(tmp, path)
}

func (s *SaveService) DeleteSave(id string) error {
    if _, err := safePath(s.savesDir, id); err != nil {
        return err
    }
    path := filepath.Join(s.savesDir, id+".yaml")
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return err
    }
    return nil
}

func (s *SaveService) CloneSave(srcId, dstId string) error {
    if _, err := safePath(s.savesDir, srcId); err != nil {
        return err
    }
    if _, err := safePath(s.savesDir, dstId); err != nil {
        return err
    }
    srcPath := filepath.Join(s.savesDir, srcId+".yaml")
    data, err := os.ReadFile(srcPath)
    if err != nil {
        return err
    }
    var doc map[string]any
    if err := yaml.Unmarshal(data, &doc); err != nil {
        return err
    }
    doc["save_id"] = dstId
    if err := os.MkdirAll(s.savesDir, 0o755); err != nil {
        return err
    }
    dstPath := filepath.Join(s.savesDir, dstId+".yaml")
    if _, err := os.Stat(dstPath); err == nil {
        return fmt.Errorf("save %q already exists", dstId)
    }
    tmp := dstPath + ".tmp"
    f, err := os.Create(tmp)
    if err != nil {
        return err
    }
    enc := yaml.NewEncoder(f)
    enc.SetIndent(2)
    if err := enc.Encode(doc); err != nil {
        f.Close()
        os.Remove(tmp)
        return err
    }
    enc.Close()
    f.Close()
    return os.Rename(tmp, dstPath)
}
