// Command collector reads statistics from an appleJuice Core and forwards them to targets.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/applejuicenetz/core-information-collector/internal/buildinfo"
	"github.com/applejuicenetz/core-information-collector/internal/collector"
	"github.com/applejuicenetz/core-information-collector/internal/config"
	"github.com/applejuicenetz/core-information-collector/internal/logx"
	"github.com/applejuicenetz/core-information-collector/internal/status"
	"github.com/applejuicenetz/core-information-collector/internal/update"
)

// AppName is shown in the tray and in the window titles.
const AppName = "AJCollector"

func main() {
	dir := flag.String("config", os.Getenv("AJ_COLLECTOR_CONFIG_DIR"), "directory with collector.xml (default: ~/appleJuice/collector)")
	noTray := flag.Bool("notray", os.Getenv("AJ_NOTRAY") != "", "do not show a window or tray icon")
	version := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *version {
		fmt.Println(buildinfo.Version)
		return
	}
	if *dir == "" {
		d, err := config.DefaultDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		*dir = d
	}
	if !acquireInstance(*dir) {
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, stop, *dir, *noTray, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		stop()
		os.Exit(1)
	}
}

// run loads the configuration and starts the collector. With a window the GUI loop owns the
// calling goroutine (required on some platforms); without it (no display, -notray, build without
// CGO) run blocks until ctx ends.
func run(ctx context.Context, stop context.CancelFunc, dir string, noTray bool, out io.Writer) error {
	log := logx.New(out, filepath.Join(dir, config.LogFileName))
	file, err := config.Path(dir, log)
	if err != nil {
		return err
	}
	cfg, err := config.Load(file, log)
	if err != nil {
		log.Error("%v", err)
		return err
	}
	log.Info("Collector Version %s", buildinfo.Version)

	store := &status.Store{}
	store.Set(collector.StatusRows(nil, false))
	c := &collector.Collector{
		Config:   cfg,
		Log:      log,
		Version:  buildinfo.Version,
		OnStatus: store.Set,
	}

	var ui *trayUI
	if !noTray && hasDisplay() {
		ui = newTrayUI(file, store, c.Run, stop, cfg.TrayIcon)
		c.OnInfo = ui.setTooltip
	}

	updateCh := make(chan string, 1)
	go checkUpdate(ctx, log, updateCh)
	go c.Loop(ctx, 0, cfg.Interval)

	if ui != nil {
		go func() {
			for {
				select {
				case v := <-updateCh:
					ui.setUpdate(v)
				case <-ctx.Done():
					return
				}
			}
		}()
		ui.run(ctx)
	}
	<-ctx.Done()
	return nil
}

// checkUpdate looks for a newer release once and sends its tag to found.
func checkUpdate(ctx context.Context, log *logx.Logger, found chan<- string) {
	if os.Getenv("AJ_COLLECTOR_DISABLE_UPDATE_CHECK") == "yes" || buildinfo.IsDevelopment() {
		return
	}
	latest, err := update.Latest(ctx, nil, update.APIURL, 5*time.Second)
	if err != nil {
		log.Error("%v", err)
		return
	}
	if update.Newer(latest, buildinfo.Version) {
		log.Info("neue Version %s gefunden", latest)
		select {
		case found <- latest:
		default:
		}
	}
}
