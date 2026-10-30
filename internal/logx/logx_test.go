package logx

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInfoGoesToOutputOnly(t *testing.T) {
	var out bytes.Buffer
	file := filepath.Join(t.TempDir(), "sub", "collector.log")
	l := New(&out, file)
	l.Info("hello %d", 1)
	if out.String() != "info: hello 1\n" {
		t.Errorf("output %q", out.String())
	}
	if _, err := os.Stat(file); err == nil {
		t.Error("info must not create the log file")
	}
}

func TestErrorGoesToOutputAndFile(t *testing.T) {
	var out bytes.Buffer
	file := filepath.Join(t.TempDir(), "sub", "collector.log")
	l := New(&out, file)
	l.now = func() time.Time { return time.Date(2026, 10, 19, 12, 30, 5, 0, time.UTC) }
	l.Error("boom %s", "x")
	if out.String() != "error: boom x\n" {
		t.Errorf("output %q", out.String())
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "2026-10-19 12:30:05 error: boom x\n" {
		t.Errorf("file %q", data)
	}
}

func TestRotation(t *testing.T) {
	file := filepath.Join(t.TempDir(), "collector.log")
	if err := os.WriteFile(file, bytes.Repeat([]byte("x"), MaxFileSize), 0o644); err != nil {
		t.Fatal(err)
	}
	l := New(nil, file)
	l.Error("fresh")
	old, err := os.ReadFile(file + ".old")
	if err != nil || len(old) != MaxFileSize {
		t.Fatalf("old file: %d bytes, %v", len(old), err)
	}
	data, _ := os.ReadFile(file)
	if !strings.HasSuffix(string(data), "error: fresh\n") || len(data) > 100 {
		t.Errorf("new file %q", data)
	}
}

func TestNilAndZeroLoggerAreSafe(t *testing.T) {
	var l *Logger
	l.Info("x")
	l.Error("x")
	(&Logger{}).Error("x")
}

func TestConcurrentUse(t *testing.T) {
	var out bytes.Buffer
	l := New(&out, filepath.Join(t.TempDir(), "c.log"))
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); l.Error("line"); l.Info("line") }()
	}
	wg.Wait()
	if n := strings.Count(out.String(), "\n"); n != 40 {
		t.Errorf("%d lines", n)
	}
}
