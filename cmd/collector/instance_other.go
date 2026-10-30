//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// instanceLock keeps the lock file open (and the lock held) until the process ends.
var instanceLock *os.File

// acquireInstance reports whether this is the only collector using the configuration directory
// dir. It holds an exclusive flock on dir/collector.lock; the kernel releases it when the
// process ends, also after a crash. The Flatpak shares ~/appleJuice with every instance, so the
// lock also covers a second Flatpak start. A directory that cannot be locked does not block.
func acquireInstance(dir string) bool {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return true
	}
	f, err := os.OpenFile(filepath.Join(dir, "collector.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return true
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if err == syscall.EWOULDBLOCK {
			fmt.Fprintln(os.Stderr, AppName+" läuft bereits.")
			return false
		}
		return true
	}
	instanceLock = f
	return true
}
