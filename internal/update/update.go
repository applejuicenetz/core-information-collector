// Package update checks GitHub for a newer release.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// URLs of the release API and the release page.
const (
	APIURL     = "https://api.github.com/repos/applejuicenetz/collector/releases/latest"
	ReleaseURL = "https://github.com/applejuicenetz/collector/releases"
)

// Latest returns the tag of the latest release ("0.0.0" when the answer carries none).
func Latest(ctx context.Context, client *http.Client, url string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Server returned HTTP response code: %d for URL: %s", resp.StatusCode, url)
	}
	var release struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&release); err != nil {
		return "", err
	}
	if release.Tag == "" {
		return "0.0.0", nil
	}
	return release.Tag, nil
}

// Compare compares dotted version numbers numerically and returns -1, 0 or 1. A leading "v"
// is ignored, missing components count as 0, a non-numeric component compares as 0.
func Compare(a, b string) int {
	pa, pb := components(a), components(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y uint64
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			if x > y {
				return 1
			}
			return -1
		}
	}
	return 0
}

func components(v string) []uint64 {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	out := make([]uint64, len(parts))
	for i, p := range parts {
		n, _ := strconv.ParseUint(p, 10, 64)
		out[i] = n
	}
	return out
}

// Newer reports whether latest is a newer release than current. A development build
// (current "dev" or containing SNAPSHOT) never reports an update.
func Newer(latest, current string) bool {
	if current == "" || current == "dev" || strings.Contains(current, "SNAPSHOT") {
		return false
	}
	return Compare(latest, current) > 0
}
