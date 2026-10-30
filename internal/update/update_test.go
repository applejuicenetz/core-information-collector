package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"4.0.1", "4.0.1", 0},
		{"4.0.2", "4.0.1", 1},
		{"4.0.1", "4.0.2", -1},
		{"4.10.0", "4.9.9", 1},
		{"10.0.0", "9.99.99", 1},
		{"4.0", "4.0.0", 0},
		{"4.0.0.1", "4.0.0", 1},
		{"4.0.0", "4.0.0.1", -1},
		{"v5.0.0", "4.0.9", 1},
		{"004.001", "4.1", 0},
		{"0.0.0", "4.0.1", -1},
		{"abc", "0", 0},
	}
	for _, tc := range tests {
		if got := Compare(tc.a, tc.b); got != tc.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNewer(t *testing.T) {
	if !Newer("5.0.0", "4.0.1") {
		t.Error("5.0.0 should be newer")
	}
	if Newer("4.0.1", "4.0.1") || Newer("4.0.0", "4.0.1") {
		t.Error("same or older is not newer")
	}
	for _, dev := range []string{"", "dev", "5.0.0-SNAPSHOT", "SNAPSHOT"} {
		if Newer("99.0.0", dev) {
			t.Errorf("development build %q must not report updates", dev)
		}
	}
}

func TestLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			_, _ = w.Write([]byte(`{"tag_name":"4.2.0","name":"x"}`))
		case "/empty":
			_, _ = w.Write([]byte(`{}`))
		case "/bad":
			_, _ = w.Write([]byte(`not json`))
		default:
			http.Error(w, "limit", http.StatusForbidden)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	if tag, err := Latest(ctx, srv.Client(), srv.URL+"/ok", time.Second); err != nil || tag != "4.2.0" {
		t.Errorf("ok: %q, %v", tag, err)
	}
	if tag, err := Latest(ctx, srv.Client(), srv.URL+"/empty", time.Second); err != nil || tag != "0.0.0" {
		t.Errorf("empty: %q, %v", tag, err)
	}
	if _, err := Latest(ctx, srv.Client(), srv.URL+"/bad", time.Second); err == nil {
		t.Error("invalid JSON must fail")
	}
	if _, err := Latest(ctx, srv.Client(), srv.URL+"/denied", time.Second); err == nil {
		t.Error("HTTP 403 must fail")
	}
	if _, err := Latest(ctx, srv.Client(), "http://127.0.0.1:1/x", time.Second); err == nil {
		t.Error("connection failure must fail")
	}
}
