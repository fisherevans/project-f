package server

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"fisherevans.com/project/f/internal/schema"
)

type OverlayService struct {
	assetsDir string
}

func NewOverlayService(assetsDir string) *OverlayService {
	return &OverlayService{assetsDir: assetsDir}
}

func (s *OverlayService) overlaysDir() string {
	return filepath.Join(s.assetsDir, "overlays")
}

func (s *OverlayService) ListFlows() ([]OverlayFlowEntry, error) {
	root := s.overlaysDir()
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var flows []OverlayFlowEntry
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		base := d.Name()
		if base == "rects.yaml" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		name := strings.TrimSuffix(filepath.ToSlash(rel), ".yaml")
		flow, err := readYAML[schema.OverlayFlow](path)
		if err != nil {
			return nil
		}
		flows = append(flows, OverlayFlowEntry{
			Name:        name,
			Description: flow.Description,
			TargetCount: len(flow.Targets),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return flows, nil
}

func (s *OverlayService) GetFlow(name string) (*OverlayFlowDetail, error) {
	if _, err := safePath(s.overlaysDir(), name); err != nil {
		return nil, err
	}
	path := filepath.Join(s.overlaysDir(), name+".yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	flow, err := readYAML[schema.OverlayFlow](path)
	if err != nil {
		return nil, err
	}
	return &OverlayFlowDetail{
		OverlayFlowEntry: OverlayFlowEntry{
			Name:        name,
			Description: flow.Description,
			TargetCount: len(flow.Targets),
		},
		Flow:    *flow,
		RawYaml: string(raw),
	}, nil
}

func (s *OverlayService) SaveFlow(name string, content string) error {
	if _, err := safePath(s.overlaysDir(), name); err != nil {
		return err
	}
	path := filepath.Join(s.overlaysDir(), name+".yaml")
	return atomicWrite(path, []byte(content))
}

func (s *OverlayService) DeleteFlow(name string) error {
	if _, err := safePath(s.overlaysDir(), name); err != nil {
		return err
	}
	path := filepath.Join(s.overlaysDir(), name+".yaml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *OverlayService) GetNamedRects() (*NamedRectsDetail, error) {
	path := filepath.Join(s.overlaysDir(), "rects.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &NamedRectsDetail{
				Rects:   map[string]schema.OverlayRect{},
				RawYaml: "",
			}, nil
		}
		return nil, err
	}
	rects, err := readYAML[map[string]schema.OverlayRect](path)
	if err != nil {
		return nil, err
	}
	return &NamedRectsDetail{
		Rects:   *rects,
		RawYaml: string(raw),
	}, nil
}

func (s *OverlayService) SaveNamedRects(content string) error {
	path := filepath.Join(s.overlaysDir(), "rects.yaml")
	return atomicWrite(path, []byte(content))
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
