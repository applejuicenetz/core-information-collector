// Package forward posts the collected values to the configured targets.
package forward

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
)

// Timeout is the connect and response timeout of a forward request.
const Timeout = 5 * time.Second

// Payload is the JSON body. Field order and names are the contract with the receiving API.
type Payload struct {
	ForwardLine         string  `json:"forward_line"`
	CoreVersion         string  `json:"core_version"`
	CoreSystem          string  `json:"core_system"`
	CoreCredits         int64   `json:"core_credits"`
	CoreConnections     int64   `json:"core_connections"`
	CoreUpload          int64   `json:"core_upload"`
	CoreDownload        int64   `json:"core_download"`
	CoreSessionUpload   int64   `json:"core_session_upload"`
	CoreSessionDownload int64   `json:"core_session_download"`
	CoreUploads         int64   `json:"core_uploads"`
	CoreDownloads       int64   `json:"core_downloads"`
	CoreDownloadsReady  int64   `json:"core_downloads_ready"`
	ShareFiles          int64   `json:"share_files"`
	ShareSize           int64   `json:"share_size"`
	NetworkUser         int64   `json:"network_user"`
	NetworkFiles        int64   `json:"network_files"`
	NetworkFileSize     float64 `json:"network_file_size"`
}

// UserAgent builds the User-Agent header, for example
// "Collector/5.0.0; Go/go1.27.2; (linux/arm64)".
func UserAgent(version string) string {
	return fmt.Sprintf("Collector/%s; Go/%s; (%s/%s)", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// Post sends payload to url. A non-empty token becomes "Authorization: Token <token>". Any
// status of 400 or above is an error; the response body is discarded.
func Post(ctx context.Context, client *http.Client, url, token, userAgent string, payload Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Server returned HTTP response code: %d for URL: %s", resp.StatusCode, url)
	}
	return nil
}
