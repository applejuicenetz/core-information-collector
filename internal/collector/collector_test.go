package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/applejuicenetz/core-information-collector/internal/config"
	"github.com/applejuicenetz/core-information-collector/internal/core"
	"github.com/applejuicenetz/core-information-collector/internal/logx"
)

func snapshot() *core.Snapshot {
	return &core.Snapshot{
		Version: "0.35.185.93", System: "Linux",
		Credits: 15*1024*1024*1024 + 512*1024*1024, Connections: 21,
		SessionUpload: 31 << 30, SessionDownload: 2 << 30,
		UploadSpeed: 1258291, DownloadSpeed: 61440,
		Uploads: 12, Downloads: 8, DownloadsReady: 3,
		ShareFiles: 36, ShareSize: 4939212390,
		NetworkUsers: 700, NetworkFiles: 3182468, NetworkFileSize: 798 * 1024 * 1024,
	}
}

func TestPlaceholdersCoverEveryDocumentedKey(t *testing.T) {
	v := Placeholders(snapshot())
	want := map[string]string{
		"%coreVersion%": "0.35.185.93", "%coreSystem%": "Linux", "%coreCredits%": "15,5GB",
		"%coreConnections%": "21", "%coreSessionUpload%": "31GB", "%coreSessionDownload%": "2GB",
		"%coreUploadSpeed%": "1,2MB/s", "%coreDownloadSpeed%": "60kB/s",
		"%coreUploads%": "12", "%coreDownloads%": "8", "%coreDownloadsReady%": "3",
		"%shareFiles%": "36", "%shareSize%": "4,6GB",
		"%networkUser%": "700", "%networkFiles%": "3.182.468", "%networkFileSize%": "798,0TB",
	}
	if len(v) != len(want) {
		t.Errorf("%d placeholders, want %d", len(v), len(want))
	}
	for k, w := range want {
		if v[k] != w {
			t.Errorf("%s = %q, want %q", k, v[k], w)
		}
	}
}

func TestReplace(t *testing.T) {
	v := Placeholders(snapshot())
	got := Replace("Core `%coreVersion%` - %coreCredits% - %unknown% - %coreVersion%", v)
	if got != "Core `0.35.185.93` - 15,5GB - %unknown% - 0.35.185.93" {
		t.Errorf("%q", got)
	}
	if Replace("", v) != "" || Replace("plain", v) != "plain" {
		t.Error("plain text must stay unchanged")
	}
}

func TestStatusRowsDropStaleValues(t *testing.T) {
	v := map[string]string{"%coreCredits%": "123"}
	rows := StatusRows(v, false)
	if rows["Status"] != "Core-Abfrage fehlgeschlagen" || len(rows) != 1 {
		t.Errorf("%v", rows)
	}
	rows = StatusRows(v, true)
	if rows["Status"] != "Core verbunden" || rows["%coreCredits%"] != "123" {
		t.Errorf("%v", rows)
	}
	v["x"] = "y"
	if _, ok := rows["x"]; ok {
		t.Error("rows must be a copy")
	}
}

// fakeCore serves the three documents for password "pw".
func fakeCore(t *testing.T, fail *atomic.Bool) *httptest.Server {
	t.Helper()
	docs := map[string]string{
		"information": `<applejuice><generalinformation><version>0.35.185.93</version><system>Linux</system></generalinformation></applejuice>`,
		"share":       `<applejuice><shares><share size="1000"/><share size="24"/></shares></applejuice>`,
		"modified": `<applejuice><download id="1" ready="0"/><download id="2" ready="9"/><upload id="3"/>
<information credits="2048" sessionupload="1" sessiondownload="2" uploadspeed="3" downloadspeed="4" openconnections="5"/>
<networkinfo users="6" files="7000" filesize="1536.5"/></applejuice>`,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail != nil && fail.Load() {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/xml/"), ".xml")
		_, _ = w.Write([]byte(docs[name]))
	}))
	t.Cleanup(srv.Close)
	return srv
}

type received struct {
	mu     sync.Mutex
	bodies []map[string]any
	auth   []string
	ua     []string
}

func target(t *testing.T, r *received, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, _ := io.ReadAll(req.Body)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		r.mu.Lock()
		r.bodies = append(r.bodies, m)
		r.auth = append(r.auth, req.Header.Get("Authorization"))
		r.ua = append(r.ua, req.Header.Get("User-Agent"))
		r.mu.Unlock()
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newCollector(t *testing.T, coreURL string, targets []config.Target) (*Collector, *bytes.Buffer) {
	t.Helper()
	i := strings.LastIndex(coreURL, ":")
	var out bytes.Buffer
	return &Collector{
		Config: &config.Config{
			InfoLine: "Credits %coreCredits% - Down %coreDownloadSpeed%", Interval: time.Minute,
			CoreHost: coreURL[:i], CorePort: coreURL[i+1:], CorePassword: "pw",
			CoreTimeout: 2 * time.Second, Targets: targets,
		},
		Log:     logx.New(&out, ""),
		Version: "5.0.0",
	}, &out
}

func TestRunCollectsAndForwardsToEveryTarget(t *testing.T) {
	var r1, r2 received
	t1, t2 := target(t, &r1, 200), target(t, &r2, 200)
	c, out := newCollector(t, fakeCore(t, nil).URL, []config.Target{
		{URL: t1.URL, Token: "tok1", Line: "v=%coreVersion% s=%shareFiles%"},
		{URL: t2.URL, Token: "", Line: "n=%networkFiles%"},
	})
	var info string
	var rows map[string]string
	c.OnInfo = func(l string) { info = l }
	c.OnStatus = func(m map[string]string) { rows = m }

	if err := c.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if info != "Credits 2kB - Down 4B/s" {
		t.Errorf("info line %q", info)
	}
	if rows["Status"] != "Core verbunden" || rows["%shareFiles%"] != "2" {
		t.Errorf("rows %v", rows)
	}
	if len(r1.bodies) != 1 || len(r2.bodies) != 1 {
		t.Fatalf("%d / %d posts", len(r1.bodies), len(r2.bodies))
	}
	b := r1.bodies[0]
	if b["forward_line"] != "v=0.35.185.93 s=2" || b["core_credits"] != 2048.0 || b["share_size"] != 1024.0 ||
		b["share_files"] != 2.0 || b["core_downloads"] != 2.0 || b["core_downloads_ready"] != 1.0 ||
		b["core_uploads"] != 1.0 || b["network_file_size"] != 1536.5 || b["core_connections"] != 5.0 ||
		b["core_upload"] != 3.0 || b["core_download"] != 4.0 || b["network_user"] != 6.0 {
		t.Errorf("%v", b)
	}
	if r2.bodies[0]["forward_line"] != "n=7.000" {
		t.Errorf("%v", r2.bodies[0])
	}
	if r1.auth[0] != "Token tok1" || r2.auth[0] != "" {
		t.Errorf("auth %q %q", r1.auth[0], r2.auth[0])
	}
	if !strings.HasPrefix(r1.ua[0], "Collector/5.0.0; Go/") {
		t.Errorf("ua %q", r1.ua[0])
	}
	if strings.Count(out.String(), "forward successful") != 2 || !strings.Contains(out.String(), "info: Credits 2kB") {
		t.Errorf("log %q", out.String())
	}
}

func TestRunCoreFailureResetsStatusAndSkipsForward(t *testing.T) {
	var down atomic.Bool
	var r received
	tg := target(t, &r, 200)
	c, out := newCollector(t, fakeCore(t, &down).URL, []config.Target{{URL: tg.URL, Line: "x"}})
	var rows map[string]string
	var infos int
	c.OnInfo = func(string) { infos++ }
	c.OnStatus = func(m map[string]string) { rows = m }

	if err := c.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	down.Store(true)
	if err := c.Run(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	if rows["Status"] != StatusFailed || len(rows) != 1 {
		t.Errorf("rows %v", rows)
	}
	if infos != 1 || len(r.bodies) != 1 {
		t.Errorf("infos %d, posts %d: a failed run must not publish or forward", infos, len(r.bodies))
	}
	if !strings.Contains(out.String(), "error:") || strings.Contains(out.String(), "password=pw") {
		t.Errorf("log %q", out.String())
	}
	down.Store(false)
	if err := c.Run(context.Background()); err != nil || rows["Status"] != StatusConnected {
		t.Errorf("recovery: %v %v", err, rows)
	}
}

func TestForwardFailureDoesNotStopOtherTargets(t *testing.T) {
	var bad, good received
	tb, tg := target(t, &bad, 500), target(t, &good, 200)
	c, out := newCollector(t, fakeCore(t, nil).URL, []config.Target{
		{URL: tb.URL, Line: "a"},
		{URL: "http://127.0.0.1:1/unreachable", Line: "b"},
		{URL: tg.URL, Line: "c"},
	})
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("forward errors must not fail the run: %v", err)
	}
	if len(good.bodies) != 1 {
		t.Errorf("good target got %d posts", len(good.bodies))
	}
	if strings.Count(out.String(), "error:") != 2 || strings.Count(out.String(), "forward successful") != 1 {
		t.Errorf("log %q", out.String())
	}
}

func TestRunWithoutTargets(t *testing.T) {
	c, _ := newCollector(t, fakeCore(t, nil).URL, nil)
	if err := c.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRunsNeverOverlap(t *testing.T) {
	var active, peak atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "information.xml") {
			n := active.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			active.Add(-1)
		}
		http.Error(w, "stop", http.StatusInternalServerError)
	}))
	defer srv.Close()
	c, _ := newCollector(t, srv.URL, nil)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = c.Run(context.Background()) }()
	}
	wg.Wait()
	if peak.Load() != 1 {
		t.Errorf("%d concurrent runs", peak.Load())
	}
}

func TestLoopRunsRepeatedlyAndStopsOnCancel(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "information.xml") {
			hits.Add(1)
		}
		http.Error(w, "x", http.StatusInternalServerError)
	}))
	defer srv.Close()
	c, _ := newCollector(t, srv.URL, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { c.Loop(ctx, 0, 20*time.Millisecond); close(done) }()
	deadline := time.Now().Add(3 * time.Second)
	for hits.Load() < 3 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Loop did not stop")
	}
	if hits.Load() < 3 {
		t.Errorf("only %d runs", hits.Load())
	}
}

func TestLoopHonoursInitialDelay(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "x", http.StatusInternalServerError)
	}))
	defer srv.Close()
	c, _ := newCollector(t, srv.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	c.Loop(ctx, time.Hour, time.Hour)
	if hits.Load() != 0 {
		t.Error("ran before the delay")
	}
}
