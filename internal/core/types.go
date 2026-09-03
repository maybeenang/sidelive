package core

import "time"

type StreamStatus string

const (
	StatusIdle         StreamStatus = "idle"
	StatusConnecting   StreamStatus = "connecting"
	StatusConnected    StreamStatus = "connected"
	StatusReconnecting StreamStatus = "reconnecting"
	StatusOffline      StreamStatus = "offline"
	StatusDisconnected StreamStatus = "disconnected"
	StatusFailed       StreamStatus = "failed"
)

type ProviderCapabilities struct {
	Comments, Donations, Subscriptions, Follows, Likes, Viewers bool
}

type Identity struct{ ID, Username, DisplayName string }

type Stream struct {
	ID         string       `json:"id"`
	ProviderID string       `json:"providerId"`
	Identity   Identity     `json:"identity"`
	Status     StreamStatus `json:"status"`
	Error      string       `json:"error,omitempty"`
}

type LiveUser struct {
	ID          string `json:"id,omitempty"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
}

type CommentPayload struct {
	Text string `json:"text"`
}

type LiveEvent struct {
	ID        string         `json:"id"`
	StreamID  string         `json:"streamId"`
	Provider  string         `json:"provider"`
	Type      string         `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	User      LiveUser       `json:"user"`
	Payload   CommentPayload `json:"payload"`
}

type Bounds struct{ X, Y, Width, Height int }
type Appearance struct {
	Opacity     float64 `json:"opacity"`
	FontSize    int     `json:"fontSize"`
	MaxComments int     `json:"maxComments"`
}
type Behavior struct {
	AlwaysOnTop  bool `json:"alwaysOnTop"`
	ClickThrough bool `json:"clickThrough"`
	Visible      bool `json:"visible"`
	Locked       bool `json:"locked"`
}
type OverlaySource struct {
	OverlayID string `json:"overlayId"`
	StreamID  string `json:"streamId"`
}
type Overlay struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Sources    []OverlaySource `json:"sources"`
	Bounds     Bounds          `json:"bounds"`
	Appearance Appearance      `json:"appearance"`
	Behavior   Behavior        `json:"behavior"`
}

type Workspace struct {
	Streams  []Stream  `json:"streams"`
	Overlays []Overlay `json:"overlays"`
}
