//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// instanceMutex is the session-local name of the mutex that marks a running collector. Setup
// (packaging/windows.iss, AppMutex) uses the same name.
const instanceMutex = "AJCollectorSingleInstance"

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	user32           = syscall.NewLazyDLL("user32.dll")
	procCreateMutexW = kernel32.NewProc("CreateMutexW")
	procMessageBoxW  = user32.NewProc("MessageBoxW")
)

// acquireInstance reports whether this is the only running collector in the user session. The
// mutex handle stays open until the process ends. A second start shows a short message.
func acquireInstance(_ string) bool {
	name, _ := syscall.UTF16PtrFromString(instanceMutex)
	h, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return true // cannot decide, do not block the start
	}
	if errno, ok := err.(syscall.Errno); ok && errno == syscall.ERROR_ALREADY_EXISTS {
		text, _ := syscall.UTF16PtrFromString(AppName + " läuft bereits.")
		title, _ := syscall.UTF16PtrFromString(AppName)
		procMessageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x40) // MB_ICONINFORMATION
		return false
	}
	return true
}
