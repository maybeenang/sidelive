package routing

import (
	"github.com/sidelive/sidelive/internal/core"
	"sync"
)

type Router struct {
	mu          sync.RWMutex
	next        uint64
	subscribers map[uint64]chan core.LiveEvent
}

func New(buffer int) *Router { return &Router{subscribers: map[uint64]chan core.LiveEvent{}} }
func (r *Router) Publish(e core.LiveEvent) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ch := range r.subscribers {
		select {
		case ch <- e:
		default:
			<-ch
			ch <- e
		}
	}
}
func (r *Router) Subscribe(buffer int) (<-chan core.LiveEvent, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.next++
	id := r.next
	ch := make(chan core.LiveEvent, buffer)
	r.subscribers[id] = ch
	return ch, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if c, ok := r.subscribers[id]; ok {
			delete(r.subscribers, id)
			close(c)
		}
	}
}
