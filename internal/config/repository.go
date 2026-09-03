package config

import (
	"encoding/json"
	"errors"
	"github.com/sidelive/sidelive/internal/core"
	"os"
	"path/filepath"
	"sync"
)

type Document struct {
	Version   int            `json:"version"`
	Workspace core.Workspace `json:"workspace"`
}
type Repository struct {
	path string
	mu   sync.Mutex
}

func New(path string) *Repository { return &Repository{path: path} }
func DefaultPath() (string, error) {
	d, e := os.UserConfigDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(d, "SideLive", "config.json"), nil
}
func Defaults() Document {
	return Document{Version: 1, Workspace: core.Workspace{Streams: []core.Stream{}, Overlays: []core.Overlay{{ID: "overlay_main", Name: "Live comments", Sources: []core.OverlaySource{}, Bounds: core.Bounds{X: 80, Y: 80, Width: 420, Height: 560}, Appearance: core.Appearance{Opacity: .86, FontSize: 16, MaxComments: 8}, Behavior: core.Behavior{AlwaysOnTop: true, Visible: false}}}}}
}
func (r *Repository) Load() (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, e := os.ReadFile(r.path)
	if errors.Is(e, os.ErrNotExist) {
		return Defaults(), nil
	}
	if e != nil {
		return Document{}, e
	}
	var d Document
	if e = json.Unmarshal(b, &d); e != nil {
		return Document{}, e
	}
	if d.Version != 1 {
		return Document{}, errors.New("unsupported configuration version")
	}
	if len(d.Workspace.Overlays) == 0 {
		d.Workspace.Overlays = Defaults().Workspace.Overlays
	}
	return d, nil
}
func (r *Repository) Save(d Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, e := json.MarshalIndent(d, "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(r.path), 0700); e != nil {
		return e
	}
	tmp := r.path + ".tmp"
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	if e = os.Rename(tmp, r.path); e == nil {
		return nil
	}
	// Windows cannot replace an existing file with Rename.
	_ = os.Remove(r.path)
	if e2 := os.Rename(tmp, r.path); e2 != nil {
		_ = os.Remove(tmp)
		return e2
	}
	return nil
}
