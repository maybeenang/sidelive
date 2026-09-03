package config

import (
	"path/filepath"
	"testing"
)

func TestRepositoryRoundTrip(t *testing.T) {
	r := New(filepath.Join(t.TempDir(), "nested", "config.json"))
	d, err := r.Load()
	if err != nil {
		t.Fatal(err)
	}
	if d.Version != 1 || len(d.Workspace.Overlays) != 1 {
		t.Fatalf("unexpected defaults: %#v", d)
	}
	d.Workspace.Overlays[0].Appearance.FontSize = 22
	if err = r.Save(d); err != nil {
		t.Fatal(err)
	}
	loaded, err := r.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Workspace.Overlays[0].Appearance.FontSize; got != 22 {
		t.Fatalf("font size = %d, want 22", got)
	}
}
