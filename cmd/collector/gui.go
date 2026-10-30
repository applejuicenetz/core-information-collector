//go:build cgo && !nogui

package main

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/systray"

	"github.com/applejuicenetz/core-information-collector/internal/buildinfo"
	"github.com/applejuicenetz/core-information-collector/internal/status"
	"github.com/applejuicenetz/core-information-collector/internal/update"
)

//go:embed tray.png
var iconPNG []byte

// trayUI is the status window and the tray icon with its menu: Status, Execute, Config, About,
// Quit and, when a newer release exists, an update entry. The Fyne loop owns the calling
// goroutine (required on some platforms).
type trayUI struct {
	configFile string
	store      *status.Store
	execute    func(context.Context) error
	quit       context.CancelFunc
	showTray   bool

	mu      sync.Mutex
	app     fyne.App
	window  fyne.Window
	grid    *fyne.Container
	about   fyne.Window
	started bool
	tooltip string
	update  string
}

func newTrayUI(configFile string, store *status.Store, execute func(context.Context) error, quit context.CancelFunc, showTray bool) *trayUI {
	return &trayUI{configFile: configFile, store: store, execute: execute, quit: quit, showTray: showTray}
}

// hasDisplay reports whether a window can be opened. Linux needs an X11 or Wayland session.
func hasDisplay() bool {
	if runtime.GOOS != "linux" {
		return true
	}
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

func (t *trayUI) setTooltip(line string) {
	t.mu.Lock()
	t.tooltip = line
	started := t.started
	t.mu.Unlock()
	if started && t.showTray {
		fyne.Do(func() { applyTooltip(line) })
	}
}

func applyTooltip(line string) {
	defer func() { _ = recover() }() // a host without a tray has nothing to update
	systray.SetTooltip(line)
}

func (t *trayUI) setUpdate(version string) {
	t.mu.Lock()
	t.update = version
	started := t.started
	t.mu.Unlock()
	if started && t.showTray {
		fyne.Do(t.refreshMenu)
	}
}

// open opens a file or URL with the platform handler.
func open(target string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
	case "darwin":
		_ = exec.Command("open", target).Start()
	default:
		_ = exec.Command("xdg-open", target).Start()
	}
}

func (t *trayUI) showWindow() {
	t.window.Show()
	t.window.RequestFocus()
}

// showAbout opens a small fixed-size window with the product name and the running version.
func (t *trayUI) showAbout() {
	if t.about != nil {
		t.about.Show()
		t.about.RequestFocus()
		return
	}
	w := t.app.NewWindow("About " + AppName)
	w.SetIcon(t.app.Icon())
	w.SetFixedSize(true)
	name := canvas.NewText(AppName, theme.Color(theme.ColorNameForeground))
	name.TextSize = 16
	name.TextStyle = fyne.TextStyle{Bold: true}
	name.Alignment = fyne.TextAlignCenter
	ver := canvas.NewText("Version "+buildinfo.Version, theme.Color(theme.ColorNameForeground))
	ver.TextSize = compactTextSize
	ver.Alignment = fyne.TextAlignCenter
	w.SetContent(container.NewPadded(container.NewVBox(name, ver)))
	w.Resize(fyne.NewSize(220, 80))
	w.SetCloseIntercept(w.Hide)
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEscape {
			w.Hide()
		}
	})
	t.about = w
	w.Show()
}

func (t *trayUI) refreshMenu() {
	t.mu.Lock()
	version := t.update
	t.mu.Unlock()
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Status", t.showWindow),
		fyne.NewMenuItem("Execute", func() { go func() { _ = t.execute(context.Background()) }() }),
		fyne.NewMenuItem("Config", func() { open(t.configFile) }),
		fyne.NewMenuItem("About", func() { fyne.Do(t.showAbout) }),
	}
	if version != "" {
		items = append(items, fyne.NewMenuItem("Neue Version "+version+" herunterladen", func() { open(update.ReleaseURL) }))
	}
	menu := fyne.NewMenu(AppName, items...)
	menu.Items = append(menu.Items, fyne.NewMenuItemSeparator())
	quit := fyne.NewMenuItem("Quit", nil)
	quit.IsQuit = true
	menu.Items = append(menu.Items, quit)
	t.app.(desktop.App).SetSystemTrayMenu(menu)
}

// compactTextSize is the font size of the status rows.
const compactTextSize = 12

// compactMinWidth keeps the window wide enough for the longest values.
const compactMinWidth = 300

// rebuild fills the grid with the current rows and shrinks the window to the content.
func (t *trayUI) rebuild() {
	rows := t.store.Rows()
	objects := make([]fyne.CanvasObject, 0, 2*len(rows))
	for _, r := range rows {
		key := canvas.NewText(r.Key, theme.Color(theme.ColorNameForeground))
		key.TextSize = compactTextSize
		key.TextStyle = fyne.TextStyle{Bold: true}
		value := canvas.NewText(r.Value, theme.Color(theme.ColorNameForeground))
		value.TextSize = compactTextSize
		objects = append(objects, key, value)
	}
	t.grid.Objects = objects
	t.grid.Refresh()
	size := container.NewPadded(t.grid).MinSize()
	if size.Width < compactMinWidth {
		size.Width = compactMinWidth
	}
	t.window.Resize(size)
}

func (t *trayUI) buildWindow() {
	t.window = t.app.NewWindow(AppName)
	t.window.SetIcon(t.app.Icon())
	t.window.SetFixedSize(true)
	t.grid = container.New(layout.NewGridLayoutWithColumns(2))
	t.window.SetContent(container.NewPadded(t.grid))
	t.rebuild()
	t.window.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEscape {
			t.window.Hide()
		}
	})
	t.store.OnChange(func() { fyne.Do(t.rebuild) })
}

// run shows the window and the tray and blocks until Quit or ctx ends. Closing the window only
// hides it while a tray exists; without a tray it ends the program.
func (t *trayUI) run(ctx context.Context) {
	t.app = app.NewWithID("io.github.applejuicenetz.collector")
	t.app.SetIcon(fyne.NewStaticResource("icon.png", iconPNG))
	t.buildWindow()
	// Fyne starts the tray loop in Run only when the menu was set before it.
	if t.showTray {
		t.refreshMenu()
		t.app.(desktop.App).SetSystemTrayWindow(t.window)
	} else {
		t.window.SetCloseIntercept(func() { t.quit(); t.app.Quit() })
	}
	t.app.Lifecycle().SetOnStarted(func() {
		t.mu.Lock()
		t.started = true
		tooltip := t.tooltip
		t.mu.Unlock()
		if t.showTray && tooltip != "" {
			applyTooltip(tooltip)
		}
	})
	go func() {
		<-ctx.Done()
		fyne.Do(t.app.Quit)
	}()
	t.app.Lifecycle().SetOnStopped(t.quit)
	t.window.Show()
	t.app.Run()
}
