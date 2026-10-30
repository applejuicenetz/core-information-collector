//go:build !cgo || nogui

package main

import (
	"context"

	"github.com/applejuicenetz/core-information-collector/internal/status"
)

// trayUI is the headless stand-in: builds without CGO (container, cross builds) have no window.
type trayUI struct{}

func newTrayUI(string, *status.Store, func(context.Context) error, context.CancelFunc, bool) *trayUI {
	return &trayUI{}
}
func (*trayUI) setTooltip(string)   {}
func (*trayUI) setUpdate(string)    {}
func (*trayUI) run(context.Context) {}

func hasDisplay() bool { return false }
