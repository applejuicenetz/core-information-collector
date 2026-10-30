package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// safeBuffer is a concurrency-safe log sink.
type safeBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}
func (s *safeBuffer) String() string { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestRunEndToEnd(t *testing.T) {
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "information.xml"):
			_, _ = w.Write([]byte(`<applejuice><generalinformation><version>1.2.3</version><system>Test</system></generalinformation></applejuice>`))
		case strings.HasSuffix(r.URL.Path, "share.xml"):
			_, _ = w.Write([]byte(`<applejuice><shares><share size="10"/></shares></applejuice>`))
		default:
			_, _ = w.Write([]byte(`<applejuice><information credits="1024" sessionupload="0" sessiondownload="0" uploadspeed="0" downloadspeed="0" openconnections="0"/><networkinfo users="1" files="2" filesize="3"/></applejuice>`))
		}
	}))
	defer core.Close()

	var mu sync.Mutex
	var bodies []string
	var gotAuth string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		gotAuth = r.Header.Get("Authorization")
		mu.Unlock()
	}))
	defer target.Close()

	dir := t.TempDir()
	i := strings.LastIndex(core.URL, ":")
	xml := `<collector intervall="60000" trayIcon="false" taskbarIcon="false"><infoLine>Credits %coreCredits%</infoLine>` +
		`<core host="` + core.URL[:i] + `" port="` + core.URL[i+1:] + `" password="pw" timeout="2000"/>` +
		`<targets><target><url>` + target.URL + `</url><token>tk</token><line>%coreVersion% on %coreSystem%</line></target></targets></collector>`
	if err := os.WriteFile(filepath.Join(dir, "collector.xml"), []byte(xml), 0o644); err != nil {
		t.Fatal(err)
	}

	out := &safeBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, cancel, dir, true, out) }()

	waitFor(t, "forward", func() bool { mu.Lock(); defer mu.Unlock(); return len(bodies) == 1 })
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not stop")
	}
	mu.Lock()
	defer mu.Unlock()
	if !strings.Contains(bodies[0], `"forward_line":"1.2.3 on Test"`) || gotAuth != "Token tk" {
		t.Errorf("%s / %q", bodies[0], gotAuth)
	}
	if !strings.Contains(out.String(), "info: Credits 1kB") || !strings.Contains(out.String(), "Collector Version dev") {
		t.Errorf("log: %q", out.String())
	}
}

func TestRunWritesDefaultConfigAndKeepsRunningWhenCoreIsDown(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "appleJuice", "collector")
	// Default config targets the public API; replace it before the first run by creating the
	// file ourselves through an invalid port so nothing leaves the machine.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	xml := `<collector trayIcon="false"><core host="http://127.0.0.1" port="1" password="secret-pw" timeout="500"/></collector>`
	if err := os.WriteFile(filepath.Join(dir, "collector.xml"), []byte(xml), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &safeBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, cancel, dir, true, out) }()
	waitFor(t, "error log", func() bool { return strings.Contains(out.String(), "error:") })
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	log, err := os.ReadFile(filepath.Join(dir, "collector.log"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(log), "secret-pw") || strings.Contains(out.String(), "secret-pw") {
		t.Error("password leaked into the log")
	}
}

func TestRunRejectsBrokenConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "collector.xml"), []byte("<collector"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &safeBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := run(ctx, cancel, dir, true, out); err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(out.String(), "error:") {
		t.Errorf("%q", out.String())
	}
}

func TestRunFailsWhenConfigDirIsNotWritable(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := run(ctx, cancel, filepath.Join(blocker, "x"), true, io.Discard); err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckUpdateIsSkippedForDevelopmentBuilds(t *testing.T) {
	// buildinfo.Version is "dev" in tests: no request may be made and nothing is sent.
	found := make(chan string, 1)
	checkUpdate(context.Background(), nil, found)
	select {
	case v := <-found:
		t.Errorf("unexpected update %q", v)
	default:
	}
}
