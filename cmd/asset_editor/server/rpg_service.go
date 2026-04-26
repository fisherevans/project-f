package server

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "fisherevans.com/project/f/internal/schema"
    "gopkg.in/yaml.v3"
)

type RPGService struct {
    assetsDir string
}

func NewRPGService(assetsDir string) *RPGService {
    return &RPGService{assetsDir: assetsDir}
}

func (s *RPGService) skillsDir() string {
    return filepath.Join(s.assetsDir, "rpg", "skills")
}

func (s *RPGService) primortalsDir() string {
    return filepath.Join(s.assetsDir, "rpg", "primortals")
}

func (s *RPGService) ListSkills() ([]schema.RPGSkill, error) {
    return readAllYAML[schema.RPGSkill](s.skillsDir())
}

func (s *RPGService) GetSkill(id string) (*schema.RPGSkill, error) {
    if _, err := safePath(s.skillsDir(), id); err != nil {
        return nil, err
    }
    return readYAML[schema.RPGSkill](filepath.Join(s.skillsDir(), id+".yaml"))
}

func (s *RPGService) SaveSkill(skill schema.RPGSkill) error {
    if _, err := safePath(s.skillsDir(), skill.Id); err != nil {
        return err
    }
    return writeYAML(filepath.Join(s.skillsDir(), skill.Id+".yaml"), skill)
}

func (s *RPGService) DeleteSkill(id string) error {
    if _, err := safePath(s.skillsDir(), id); err != nil {
        return err
    }
    path := filepath.Join(s.skillsDir(), id+".yaml")
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return err
    }
    return nil
}

func (s *RPGService) ListPrimortals() ([]schema.RPGPrimortal, error) {
    return readAllYAML[schema.RPGPrimortal](s.primortalsDir())
}

func (s *RPGService) GetPrimortal(type_ string) (*schema.RPGPrimortal, error) {
    if _, err := safePath(s.primortalsDir(), type_); err != nil {
        return nil, err
    }
    return readYAML[schema.RPGPrimortal](filepath.Join(s.primortalsDir(), type_+".yaml"))
}

func (s *RPGService) SavePrimortal(p schema.RPGPrimortal) error {
    if _, err := safePath(s.primortalsDir(), p.Type); err != nil {
        return err
    }
    return writeYAML(filepath.Join(s.primortalsDir(), p.Type+".yaml"), p)
}

func (s *RPGService) DeletePrimortal(type_ string) error {
    if _, err := safePath(s.primortalsDir(), type_); err != nil {
        return err
    }
    path := filepath.Join(s.primortalsDir(), type_+".yaml")
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return err
    }
    return nil
}

func readAllYAML[T any](dir string) ([]T, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, err
    }
    var items []T
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
            continue
        }
        item, err := readYAML[T](filepath.Join(dir, entry.Name()))
        if err != nil {
            return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
        }
        items = append(items, *item)
    }
    return items, nil
}

func readYAML[T any](path string) (*T, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var v T
    if err := yaml.Unmarshal(data, &v); err != nil {
        return nil, err
    }
    return &v, nil
}

func writeYAML(path string, v any) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return err
    }
    tmp := path + ".tmp"
    f, err := os.Create(tmp)
    if err != nil {
        return err
    }
    enc := yaml.NewEncoder(f)
    enc.SetIndent(2)
    if err := enc.Encode(v); err != nil {
        f.Close()
        os.Remove(tmp)
        return err
    }
    if err := enc.Close(); err != nil {
        f.Close()
        os.Remove(tmp)
        return err
    }
    if err := f.Close(); err != nil {
        os.Remove(tmp)
        return err
    }
    return os.Rename(tmp, path)
}
