package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/sidelive/sidelive/internal/core"
)

type Sink func(core.LiveEvent)
type Provider interface {
	ID() string
	Name() string
	Capabilities() core.ProviderCapabilities
	Connect(context.Context, core.Stream, Sink) error
}

type Registry struct {
	mu    sync.RWMutex
	items map[string]Provider
}

func NewRegistry(ps ...Provider) *Registry {
	r := &Registry{items: map[string]Provider{}}
	for _, p := range ps {
		r.Register(p)
	}
	return r
}
func (r *Registry) Register(p Provider) { r.mu.Lock(); defer r.mu.Unlock(); r.items[p.ID()] = p }
func (r *Registry) Get(id string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok {
		return nil, fmt.Errorf("provider %q is not registered", id)
	}
	return p, nil
}

type Info struct {
	ID, Name     string
	Capabilities core.ProviderCapabilities
}

func (r *Registry) List() []Info {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Info, 0, len(r.items))
	for _, p := range r.items {
		out = append(out, Info{p.ID(), p.Name(), p.Capabilities()})
	}
	return out
}
