package demo

import (
	"context"
	"fmt"
	"github.com/sidelive/sidelive/internal/core"
	"github.com/sidelive/sidelive/internal/providers"
	"math/rand/v2"
	"time"
)

type Provider struct{}

func New() *Provider           { return &Provider{} }
func (*Provider) ID() string   { return "demo" }
func (*Provider) Name() string { return "Demo stream" }
func (*Provider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{Comments: true}
}
func (*Provider) Connect(ctx context.Context, s core.Stream, sink providers.Sink) error {
	sink(core.LiveEvent{StreamID: s.ID, Provider: "demo", Type: "_connected", Timestamp: time.Now()})
	names := []string{"Ayu", "Bima", "Citra", "Danu", "Eka"}
	messages := []string{"Halo! Semangat live-nya 👋", "Audio jernih banget", "Salam dari Bandung!", "Boleh jelaskan bagian tadi?", "Keren, lanjutkan! ✨"}
	t := time.NewTicker(2200 * time.Millisecond)
	defer t.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-t.C:
			n := names[i%len(names)]
			sink(core.LiveEvent{ID: fmt.Sprintf("demo_%d", now.UnixNano()), StreamID: s.ID, Provider: "demo", Type: "comment", Timestamp: now, User: core.LiveUser{Username: "demo_user", DisplayName: n}, Payload: core.CommentPayload{Text: messages[rand.IntN(len(messages))]}})
			i++
		}
	}
}
