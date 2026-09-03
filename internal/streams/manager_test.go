package streams

import (
	"context"
	"github.com/sidelive/sidelive/internal/core"
	"github.com/sidelive/sidelive/internal/providers"
	"testing"
	"time"
)

type fakeProvider struct{}

func (fakeProvider) ID() string   { return "fake" }
func (fakeProvider) Name() string { return "Fake" }
func (fakeProvider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{Comments: true}
}
func (fakeProvider) Connect(ctx context.Context, s core.Stream, sink providers.Sink) error {
	sink(core.LiveEvent{ID: "event", StreamID: s.ID})
	<-ctx.Done()
	return nil
}
func TestManagerLifecycle(t *testing.T) {
	events := make(chan core.LiveEvent, 1)
	statuses := make(chan core.StreamStatus, 4)
	m := New(providers.NewRegistry(fakeProvider{}), func(e core.LiveEvent) { events <- e }, func(s core.Stream) { statuses <- s.Status })
	stream := core.Stream{ID: "stream", ProviderID: "fake"}
	if err := m.Connect(stream); err != nil {
		t.Fatal(err)
	}
	select {
	case <-events:
	case <-time.After(time.Second):
		t.Fatal("event not received")
	}
	m.Disconnect(stream.ID)
	deadline := time.After(time.Second)
	for {
		select {
		case status := <-statuses:
			if status == core.StatusDisconnected {
				return
			}
		case <-deadline:
			t.Fatal("disconnect status not received")
		}
	}
}
