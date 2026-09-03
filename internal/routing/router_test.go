package routing

import (
	"github.com/sidelive/sidelive/internal/core"
	"testing"
	"time"
)

func TestRouterFanout(t *testing.T) {
	r := New(2)
	a, offA := r.Subscribe(2)
	defer offA()
	b, offB := r.Subscribe(2)
	defer offB()
	r.Publish(core.LiveEvent{ID: "one"})
	for name, ch := range map[string]<-chan core.LiveEvent{"a": a, "b": b} {
		select {
		case got := <-ch:
			if got.ID != "one" {
				t.Errorf("%s got %q", name, got.ID)
			}
		case <-time.After(time.Second):
			t.Fatalf("%s did not receive event", name)
		}
	}
}
func TestRouterDropsOldestForSlowConsumer(t *testing.T) {
	r := New(1)
	ch, off := r.Subscribe(1)
	defer off()
	r.Publish(core.LiveEvent{ID: "old"})
	r.Publish(core.LiveEvent{ID: "new"})
	if got := <-ch; got.ID != "new" {
		t.Fatalf("got %q, want newest event", got.ID)
	}
}
