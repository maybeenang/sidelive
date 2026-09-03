package desktop

import (
	"github.com/sidelive/sidelive/internal/core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"runtime"
	"sync"
)

type Overlay struct {
	mu     sync.RWMutex
	app    *application.App
	window *application.WebviewWindow
}

func NewOverlay() *Overlay { return &Overlay{} }
func (o *Overlay) Attach(app *application.App, model core.Overlay) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.app = app
	background := application.BackgroundTypeTransparent
	if runtime.GOOS == "linux" {
		background = application.BackgroundTypeSolid
	}
	o.window = app.Window.NewWithOptions(application.WebviewWindowOptions{Name: model.ID, Title: model.Name, URL: "/?view=overlay", Width: model.Bounds.Width, Height: model.Bounds.Height, InitialPosition: application.WindowXY, X: model.Bounds.X, Y: model.Bounds.Y, AlwaysOnTop: model.Behavior.AlwaysOnTop, Frameless: true, BackgroundType: background, IgnoreMouseEvents: model.Behavior.ClickThrough, Hidden: !model.Behavior.Visible, MinWidth: 300, MinHeight: 250})
}
func (o *Overlay) Apply(v core.Overlay) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if o.window == nil {
		return
	}
	o.window.SetSize(v.Bounds.Width, v.Bounds.Height)
	o.window.SetPosition(v.Bounds.X, v.Bounds.Y)
	o.window.SetAlwaysOnTop(v.Behavior.AlwaysOnTop)
	o.window.SetIgnoreMouseEvents(v.Behavior.ClickThrough)
}
func (o *Overlay) Show() {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if o.window != nil {
		o.window.Show()
	}
}
func (o *Overlay) Hide() {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if o.window != nil {
		o.window.Hide()
	}
}
