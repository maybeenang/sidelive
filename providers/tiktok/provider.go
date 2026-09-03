package tiktok

import (
	"context"
	"fmt"
	golive "github.com/PirateTok/live-go"
	"github.com/PirateTok/live-go/events"
	pb "github.com/PirateTok/live-go/proto"
	"github.com/sidelive/sidelive/internal/core"
	"github.com/sidelive/sidelive/internal/providers"
	"strings"
	"time"
)

type Provider struct{}

func New() *Provider           { return &Provider{} }
func (*Provider) ID() string   { return "tiktok" }
func (*Provider) Name() string { return "TikTok LIVE" }
func (*Provider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{Comments: true}
}
func (*Provider) Connect(ctx context.Context, s core.Stream, sink providers.Sink) error {
	username := strings.TrimPrefix(strings.TrimSpace(s.Identity.Username), "@")
	if username == "" {
		return fmt.Errorf("username TikTok wajib diisi")
	}
	ch, err := golive.NewClient(username).MaxRetries(0).Connect(ctx)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-ch:
			if !ok {
				return fmt.Errorf("koneksi TikTok terputus")
			}
			if evt.Type == events.EventConnected {
				sink(core.LiveEvent{StreamID: s.ID, Provider: "tiktok", Type: "_connected", Timestamp: time.Now()})
				continue
			}
			if evt.Type != events.EventChat {
				continue
			}
			msg, ok := evt.Data.(*pb.WebcastChatMessage)
			if !ok {
				continue
			}
			u := msg.GetUser()
			avatar := ""
			if u.GetAvatarThumb() != nil && len(u.GetAvatarThumb().GetUrlList()) > 0 {
				avatar = u.GetAvatarThumb().GetUrlList()[0]
			}
			sink(core.LiveEvent{ID: fmt.Sprintf("tiktok_%s_%d", s.ID, time.Now().UnixNano()), StreamID: s.ID, Provider: "tiktok", Type: "comment", Timestamp: time.Now(), User: core.LiveUser{ID: fmt.Sprint(u.GetId()), Username: u.GetUniqueId(), DisplayName: u.GetNickname(), AvatarURL: avatar}, Payload: core.CommentPayload{Text: msg.GetContent()}})
		}
	}
}
