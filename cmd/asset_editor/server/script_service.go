package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ScriptService struct {
	assetsDir string
}

func NewScriptService(assetsDir string) *ScriptService {
	return &ScriptService{assetsDir: assetsDir}
}

func (s *ScriptService) scriptsDir() string {
	return filepath.Join(s.assetsDir, "scripts")
}

func (s *ScriptService) ListScripts() ([]ScriptFileEntry, error) {
	root := s.scriptsDir()
	var entries []ScriptFileEntry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		dir := filepath.Dir(rel)
		if dir == "." {
			dir = ""
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		summary := parseScriptSummary(data)

		entries = append(entries, ScriptFileEntry{
			Path:           rel,
			Directory:      dir,
			Name:           strings.TrimSuffix(filepath.Base(rel), ext),
			HandlerCount:   len(summary.Handlers),
			SequenceCount:  len(summary.Sequences),
			HandlerNames:   summary.Handlers,
			SequenceNames:  summary.Sequences,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking scripts: %w", err)
	}
	return entries, nil
}

func (s *ScriptService) GetScript(name string) (*ScriptFileDetail, error) {
	root := s.scriptsDir()
	path := filepath.Join(root, name)
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		path += ".yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading script: %w", err)
	}

	rel, _ := filepath.Rel(root, path)
	dir := filepath.Dir(rel)
	if dir == "." {
		dir = ""
	}
	ext := filepath.Ext(rel)

	summary := parseScriptSummary(data)

	return &ScriptFileDetail{
		ScriptFileEntry: ScriptFileEntry{
			Path:          rel,
			Directory:     dir,
			Name:          strings.TrimSuffix(filepath.Base(rel), ext),
			HandlerCount:  len(summary.Handlers),
			SequenceCount: len(summary.Sequences),
			HandlerNames:  summary.Handlers,
			SequenceNames: summary.Sequences,
		},
		RawYAML: string(data),
	}, nil
}

func (s *ScriptService) SaveScript(name string, content string) error {
	root := s.scriptsDir()
	path := filepath.Join(root, name)
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		path += ".yaml"
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".script-*.yaml")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

func (s *ScriptService) DeleteScript(name string) error {
	root := s.scriptsDir()
	path := filepath.Join(root, name)
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		path += ".yaml"
	}
	return os.Remove(path)
}

type scriptSummary struct {
	Handlers  []string
	Sequences []string
}

func parseScriptSummary(data []byte) scriptSummary {
	var raw struct {
		Handlers  yaml.Node `yaml:"handlers"`
		Sequences yaml.Node `yaml:"sequences"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return scriptSummary{}
	}

	var summary scriptSummary
	if raw.Handlers.Kind == yaml.MappingNode {
		for i := 0; i < len(raw.Handlers.Content); i += 2 {
			summary.Handlers = append(summary.Handlers, raw.Handlers.Content[i].Value)
		}
	}
	if raw.Sequences.Kind == yaml.MappingNode {
		for i := 0; i < len(raw.Sequences.Content); i += 2 {
			summary.Sequences = append(summary.Sequences, raw.Sequences.Content[i].Value)
		}
	}
	return summary
}
