package app

import (
	"fmt"
	"github.com/sidelive/sidelive/internal/config"
	"github.com/sidelive/sidelive/internal/core"
	"github.com/sidelive/sidelive/internal/providers"
	"github.com/sidelive/sidelive/internal/routing"
	"github.com/sidelive/sidelive/internal/streams"
	"strings"
	"sync"
	"time"
)

type Emitter func(string, any)
type OverlayAdapter interface {
	Apply(core.Overlay)
	Show()
	Hide()
}
type Snapshot struct {
	Workspace core.Workspace   `json:"workspace"`
	Providers []providers.Info `json:"providers"`
}
type Service struct {
	mu        sync.RWMutex
	repo      *config.Repository
	registry  *providers.Registry
	manager   *streams.Manager
	document  config.Document
	emit      Emitter
	overlay   OverlayAdapter
	router    *routing.Router
	stopRoute func()
}

func New(repo *config.Repository, registry *providers.Registry, emit Emitter, overlay OverlayAdapter) (*Service, error) {
	d, e := repo.Load()
	if e != nil {
		return nil, e
	}
	router := routing.New(128)
	events, unsubscribe := router.Subscribe(128)
	s := &Service{repo: repo, registry: registry, document: d, emit: emit, overlay: overlay, router: router, stopRoute: unsubscribe}
	s.manager = streams.New(registry, router.Publish, s.onStatus)
	go func() {
		for event := range events {
			emit("live:event", event)
		}
	}()
	return s, nil
}
func (s *Service) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{Workspace: s.document.Workspace, Providers: s.registry.List()}
}
func (s *Service) AddStream(providerID, username string) (core.Stream, error) {
	username = strings.TrimSpace(strings.TrimPrefix(username, "@"))
	if username == "" {
		return core.Stream{}, fmt.Errorf("username wajib diisi")
	}
	if _, e := s.registry.Get(providerID); e != nil {
		return core.Stream{}, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.document.Workspace.Streams) > 0 {
		return core.Stream{}, fmt.Errorf("MVP mendukung satu stream")
	}
	stream := core.Stream{ID: fmt.Sprintf("stream_%d", time.Now().UnixMilli()), ProviderID: providerID, Identity: core.Identity{Username: username, DisplayName: "@" + username}, Status: core.StatusIdle}
	s.document.Workspace.Streams = append(s.document.Workspace.Streams, stream)
	s.document.Workspace.Overlays[0].Sources = []core.OverlaySource{{OverlayID: "overlay_main", StreamID: stream.ID}}
	return stream, s.repo.Save(s.document)
}
func (s *Service) RemoveStream(id string) error {
	s.manager.Disconnect(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.document.Workspace.Streams[:0]
	for _, v := range s.document.Workspace.Streams {
		if v.ID != id {
			out = append(out, v)
		}
	}
	s.document.Workspace.Streams = out
	s.document.Workspace.Overlays[0].Sources = nil
	return s.repo.Save(s.document)
}
func (s *Service) Connect(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.document.Workspace.Streams {
		if v.ID == id {
			return s.manager.Connect(v)
		}
	}
	return fmt.Errorf("stream tidak ditemukan")
}
func (s *Service) Disconnect(id string) { s.manager.Disconnect(id) }
func (s *Service) UpdateOverlay(o core.Overlay) error {
	if o.Appearance.Opacity < .2 || o.Appearance.Opacity > 1 {
		return fmt.Errorf("opacity harus 0.2–1")
	}
	if o.Appearance.FontSize < 12 || o.Appearance.FontSize > 32 {
		return fmt.Errorf("font size harus 12–32")
	}
	if o.Appearance.MaxComments < 1 || o.Appearance.MaxComments > 50 {
		return fmt.Errorf("maksimum komentar harus 1–50")
	}
	s.mu.Lock()
	s.document.Workspace.Overlays[0] = o
	e := s.repo.Save(s.document)
	s.mu.Unlock()
	if e == nil && s.overlay != nil {
		s.overlay.Apply(o)
		if o.Behavior.Visible {
			s.overlay.Show()
		} else {
			s.overlay.Hide()
		}
	}
	return e
}
func (s *Service) SetOverlayVisible(visible bool) error {
	s.mu.RLock()
	o := s.document.Workspace.Overlays[0]
	s.mu.RUnlock()
	o.Behavior.Visible = visible
	return s.UpdateOverlay(o)
}
func (s *Service) onStatus(stream core.Stream) {
	s.mu.Lock()
	for i := range s.document.Workspace.Streams {
		if s.document.Workspace.Streams[i].ID == stream.ID {
			s.document.Workspace.Streams[i] = stream
		}
	}
	_ = s.repo.Save(s.document)
	s.mu.Unlock()
	s.emit("stream:status", stream)
}
func (s *Service) Close() { s.manager.Close(); s.stopRoute() }
