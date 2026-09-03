package streams

import (
	"context"
	"fmt"
	"github.com/sidelive/sidelive/internal/core"
	"github.com/sidelive/sidelive/internal/providers"
	"math/rand/v2"
	"sync"
	"time"
)

type StatusSink func(core.Stream)
type runtime struct {
	cancel context.CancelFunc
	done   chan struct{}
}
type Manager struct {
	mu       sync.RWMutex
	registry *providers.Registry
	sink     providers.Sink
	status   StatusSink
	runtimes map[string]*runtime
}

func New(r *providers.Registry, sink providers.Sink, status StatusSink) *Manager {
	return &Manager{registry: r, sink: sink, status: status, runtimes: map[string]*runtime{}}
}
func (m *Manager) Connect(s core.Stream) error {
	m.mu.Lock()
	if _, ok := m.runtimes[s.ID]; ok {
		m.mu.Unlock()
		return fmt.Errorf("stream already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	rt := &runtime{cancel: cancel, done: make(chan struct{})}
	m.runtimes[s.ID] = rt
	m.mu.Unlock()
	go m.run(ctx, rt, s)
	return nil
}
func (m *Manager) run(ctx context.Context, rt *runtime, s core.Stream) {
	defer close(rt.done)
	defer func() { m.mu.Lock(); delete(m.runtimes, s.ID); m.mu.Unlock() }()
	p, e := m.registry.Get(s.ProviderID)
	if e != nil {
		m.emit(s, core.StatusFailed, e)
		return
	}
	attempt := 0
	for {
		if ctx.Err() != nil {
			m.emit(s, core.StatusDisconnected, nil)
			return
		}
		if attempt == 0 {
			m.emit(s, core.StatusConnecting, nil)
		} else {
			m.emit(s, core.StatusReconnecting, nil)
		}
		connected := false
		e = p.Connect(ctx, s, func(event core.LiveEvent) {
			if !connected {
				connected = true
				attempt = 0
				m.emit(s, core.StatusConnected, nil)
			}
			if event.Type == "_connected" {
				return
			}
			m.sink(event)
		})
		if ctx.Err() != nil {
			m.emit(s, core.StatusDisconnected, nil)
			return
		}
		attempt++
		m.emit(s, core.StatusOffline, e)
		delay := min(30*time.Second, time.Second*time.Duration(1<<min(attempt, 5))) + time.Duration(rand.IntN(500))*time.Millisecond
		select {
		case <-ctx.Done():
			m.emit(s, core.StatusDisconnected, nil)
			return
		case <-time.After(delay):
		}
	}
}
func (m *Manager) emit(s core.Stream, status core.StreamStatus, e error) {
	s.Status = status
	if e != nil {
		s.Error = e.Error()
	} else {
		s.Error = ""
	}
	m.status(s)
}
func (m *Manager) Connected(s core.Stream) { m.emit(s, core.StatusConnected, nil) }
func (m *Manager) Disconnect(id string) {
	m.mu.RLock()
	rt := m.runtimes[id]
	m.mu.RUnlock()
	if rt != nil {
		rt.cancel()
	}
}
func (m *Manager) Close() {
	m.mu.RLock()
	ids := make([]string, 0, len(m.runtimes))
	for id := range m.runtimes {
		ids = append(ids, id)
	}
	m.mu.RUnlock()
	for _, id := range ids {
		m.Disconnect(id)
	}
}
