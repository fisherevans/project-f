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
			Path:                  rel,
			Directory:             dir,
			Name:                  strings.TrimSuffix(filepath.Base(rel), ext),
			HandlerCount:          len(summary.Handlers),
			SequenceCount:         len(summary.Sequences),
			HandlerNames:          summary.Handlers,
			SequenceNames:         summary.Sequences,
			CustomActionNames:     summary.CustomActions,
			ConstNames:            summary.Consts,
			PropertyTemplateNames: summary.PropertyTemplates,
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
	path, err := safePath(root, name)
	if err != nil {
		return nil, err
	}
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
			Path:                  rel,
			Directory:             dir,
			Name:                  strings.TrimSuffix(filepath.Base(rel), ext),
			HandlerCount:          len(summary.Handlers),
			SequenceCount:         len(summary.Sequences),
			HandlerNames:          summary.Handlers,
			SequenceNames:         summary.Sequences,
			CustomActionNames:     summary.CustomActions,
			ConstNames:            summary.Consts,
			PropertyTemplateNames: summary.PropertyTemplates,
		},
		RawYAML: string(data),
	}, nil
}

func (s *ScriptService) SaveScript(name string, content string) error {
	root := s.scriptsDir()
	path, err := safePath(root, name)
	if err != nil {
		return err
	}
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
	path, err := safePath(root, name)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		path += ".yaml"
	}
	return os.Remove(path)
}

type scriptSummary struct {
	Handlers          []string
	Sequences         []string
	CustomActions     []string
	Consts            []string
	PropertyTemplates []string
}

func parseScriptSummary(data []byte) scriptSummary {
	var raw struct {
		Handlers          yaml.Node `yaml:"handlers"`
		Sequences         yaml.Node `yaml:"sequences"`
		CustomActions     yaml.Node `yaml:"custom_actions"`
		Consts            yaml.Node `yaml:"consts"`
		PropertyTemplates yaml.Node `yaml:"property_templates"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return scriptSummary{}
	}

	var summary scriptSummary
	for _, node := range []struct {
		n    *yaml.Node
		dest *[]string
	}{
		{&raw.Handlers, &summary.Handlers},
		{&raw.Sequences, &summary.Sequences},
		{&raw.CustomActions, &summary.CustomActions},
		{&raw.Consts, &summary.Consts},
		{&raw.PropertyTemplates, &summary.PropertyTemplates},
	} {
		if node.n.Kind == yaml.MappingNode {
			for i := 0; i < len(node.n.Content); i += 2 {
				*node.dest = append(*node.dest, node.n.Content[i].Value)
			}
		}
	}
	return summary
}
