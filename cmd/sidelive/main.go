package main

import (
	"embed"
	"github.com/sidelive/sidelive/internal/app"
	"github.com/sidelive/sidelive/internal/config"
	"github.com/sidelive/sidelive/internal/desktop"
	"github.com/sidelive/sidelive/internal/providers"
	"github.com/sidelive/sidelive/providers/demo"
	"github.com/sidelive/sidelive/providers/tiktok"
	"github.com/wailsapp/wails/v3/pkg/application"
	"log"
	"log/slog"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	path, err := config.DefaultPath()
	if err != nil {
		log.Fatal(err)
	}
	overlay := desktop.NewOverlay()
	registry := providers.NewRegistry(tiktok.New(), demo.New())
	var desktopApp *application.App
	emit := func(name string, payload any) {
		if desktopApp != nil {
			desktopApp.Event.Emit(name, payload)
		}
	}
	service, err := app.New(config.New(path), registry, emit, overlay)
	if err != nil {
		log.Fatal(err)
	}
	desktopApp = application.New(application.Options{Name: "SideLive", Description: "Keep your live audience in sight.", LogLevel: slog.LevelInfo, Services: []application.Service{application.NewService(service)}, Assets: application.AssetOptions{Handler: application.BundledAssetFileServer(assets)}})
	desktopApp.Window.NewWithOptions(application.WebviewWindowOptions{Name: "main", Title: "SideLive", URL: "/", Width: 1080, Height: 760, MinWidth: 760, MinHeight: 620, BackgroundColour: application.NewRGBA(9, 9, 11, 255)})
	snapshot := service.Snapshot()
	overlayModel := config.Defaults().Workspace.Overlays[0]
	if len(snapshot.Workspace.Overlays) > 0 {
		overlayModel = snapshot.Workspace.Overlays[0]
	}
	overlay.Attach(desktopApp, overlayModel)
	if err = desktopApp.Run(); err != nil {
		log.Fatal(err)
	}
}
