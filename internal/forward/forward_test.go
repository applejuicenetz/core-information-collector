package forward

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func sample() Payload {
	return Payload{
		ForwardLine: "line ÄÖ `x`", CoreVersion: "0.35.185.93", CoreSystem: "Linux", CoreCredits: 5,
		CoreConnections: 6, CoreUpload: 7, CoreDownload: 8, CoreSessionUpload: 9, CoreSessionDownload: 10,
		CoreUploads: 11, CoreDownloads: 12, CoreDownloadsReady: 13, ShareFiles: 14, ShareSize: 15,
		NetworkUser: 16, NetworkFiles: 17, NetworkFileSize: 1449.57,
	}
}

func TestPostSendsContractFields(t *testing.T) {
	var gotBody []byte
	var gotHeader http.Header
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotHeader, gotMethod = r.Header.Clone(), r.Method
	}))
	defer srv.Close()

	if err := Post(context.Background(), nil, srv.URL, "secret", "Collector/1", sample()); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method %s", gotMethod)
	}
	if gotHeader.Get("Authorization") != "Token secret" {
		t.Errorf("Authorization %q", gotHeader.Get("Authorization"))
	}
	if gotHeader.Get("Content-Type") != "application/json;charset=UTF-8" {
		t.Errorf("Content-Type %q", gotHeader.Get("Content-Type"))
	}
	if gotHeader.Get("User-Agent") != "Collector/1" || gotHeader.Get("Accept-Charset") != "UTF-8" {
		t.Errorf("headers %v", gotHeader)
	}
	var m map[string]any
	if err := json.Unmarshal(gotBody, &m); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"forward_line": "line ÄÖ `x`", "core_version": "0.35.185.93", "core_system": "Linux",
		"core_credits": 5.0, "core_connections": 6.0, "core_upload": 7.0, "core_download": 8.0,
		"core_session_upload": 9.0, "core_session_download": 10.0, "core_uploads": 11.0,
		"core_downloads": 12.0, "core_downloads_ready": 13.0, "share_files": 14.0, "share_size": 15.0,
		"network_user": 16.0, "network_files": 17.0, "network_file_size": 1449.57,
	}
	if len(m) != len(want) {
		t.Errorf("%d fields, want %d: %s", len(m), len(want), gotBody)
	}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("%s = %v, want %v", k, m[k], v)
		}
	}
}

func TestPostWithoutTokenSendsNoAuthorization(t *testing.T) {
	var has bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, has = r.Header["Authorization"]
	}))
	defer srv.Close()
	if err := Post(context.Background(), nil, srv.URL, "", "ua", sample()); err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("Authorization header must be absent without token")
	}
}

func TestPostLargeNumbersStayIntegers(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
	}))
	defer srv.Close()
	p := sample()
	p.ShareSize = 9007199254740993 // 2^53 + 1
	if err := Post(context.Background(), nil, srv.URL, "", "ua", p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"share_size":9007199254740993`) {
		t.Errorf("body %s", body)
	}
}

func TestPostErrorStatuses(t *testing.T) {
	for _, code := range []int{400, 401, 404, 500} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		}))
		err := Post(context.Background(), nil, srv.URL, "", "ua", sample())
		srv.Close()
		if err == nil || !strings.Contains(err.Error(), "response code") {
			t.Errorf("status %d: %v", code, err)
		}
	}
	for _, code := range []int{200, 201, 204, 302} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		}))
		err := Post(context.Background(), srv.Client(), srv.URL, "", "ua", sample())
		srv.Close()
		if err != nil {
			t.Errorf("status %d: %v", code, err)
		}
	}
}

func TestPostTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer srv.Close()
	defer close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := Post(ctx, nil, srv.URL, "", "ua", sample()); err == nil {
		t.Fatal("expected timeout")
	}
	if time.Since(start) > 3*time.Second {
		t.Error("timeout took too long")
	}
}

func TestPostInvalidURL(t *testing.T) {
	if err := Post(context.Background(), nil, "://bad", "", "ua", sample()); err == nil {
		t.Error("expected error")
	}
}

func TestUserAgent(t *testing.T) {
	ua := UserAgent("5.0.0")
	if !strings.HasPrefix(ua, "Collector/5.0.0; Go/go") || !strings.HasSuffix(ua, ")") {
		t.Errorf("User-Agent %q", ua)
	}
}
