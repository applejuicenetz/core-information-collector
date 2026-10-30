// Package collector ties the Core client, the placeholder rendering and the forwarding together.
package collector

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/applejuicenetz/core-information-collector/internal/config"
	"github.com/applejuicenetz/core-information-collector/internal/core"
	"github.com/applejuicenetz/core-information-collector/internal/format"
	"github.com/applejuicenetz/core-information-collector/internal/forward"
	"github.com/applejuicenetz/core-information-collector/internal/logx"
)

// StatusConnected and StatusFailed are the values of the "Status" row.
const (
	StatusConnected = "Core verbunden"
	StatusFailed    = "Core-Abfrage fehlgeschlagen"
)

// Placeholders renders every placeholder of a snapshot.
func Placeholders(s *core.Snapshot) map[string]string {
	return map[string]string{
		"%coreSystem%":          s.System,
		"%coreVersion%":         s.Version,
		"%coreCredits%":         format.FileSize(s.Credits),
		"%coreConnections%":     strconv.FormatInt(s.Connections, 10),
		"%coreSessionUpload%":   format.FileSize(s.SessionUpload),
		"%coreSessionDownload%": format.FileSize(s.SessionDownload),
		"%coreUploadSpeed%":     format.FileSize(s.UploadSpeed) + "/s",
		"%coreDownloadSpeed%":   format.FileSize(s.DownloadSpeed) + "/s",
		"%coreUploads%":         strconv.FormatInt(s.Uploads, 10),
		"%coreDownloads%":       strconv.FormatInt(s.Downloads, 10),
		"%coreDownloadsReady%":  strconv.FormatInt(s.DownloadsReady, 10),
		"%shareFiles%":          format.Integer(s.ShareFiles),
		"%shareSize%":           format.FileSize(s.ShareSize),
		"%networkUser%":         strconv.FormatInt(s.NetworkUsers, 10),
		"%networkFiles%":        format.Integer(s.NetworkFiles),
		"%networkFileSize%":     format.NetworkShareSize(s.NetworkFileSize, format.FactorAuto),
	}
}

// Replace substitutes the placeholders in line. Keys are applied in sorted order, as the
// values themselves never contain a placeholder this is order independent.
func Replace(line string, values map[string]string) string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		line = strings.ReplaceAll(line, k, values[k])
	}
	return line
}

// StatusRows returns the rows of the status window. After a failed query only the status row
// remains, so no stale values are shown.
func StatusRows(values map[string]string, connected bool) map[string]string {
	rows := map[string]string{"Status": StatusFailed}
	if connected {
		rows["Status"] = StatusConnected
		for k, v := range values {
			rows[k] = v
		}
	}
	return rows
}

// Payload builds the forward body for one target line.
func Payload(s *core.Snapshot, forwardLine string) forward.Payload {
	return forward.Payload{
		ForwardLine:         forwardLine,
		CoreVersion:         s.Version,
		CoreSystem:          s.System,
		CoreCredits:         s.Credits,
		CoreConnections:     s.Connections,
		CoreUpload:          s.UploadSpeed,
		CoreDownload:        s.DownloadSpeed,
		CoreSessionUpload:   s.SessionUpload,
		CoreSessionDownload: s.SessionDownload,
		CoreUploads:         s.Uploads,
		CoreDownloads:       s.Downloads,
		CoreDownloadsReady:  s.DownloadsReady,
		ShareFiles:          s.ShareFiles,
		ShareSize:           s.ShareSize,
		NetworkUser:         s.NetworkUsers,
		NetworkFiles:        s.NetworkFiles,
		NetworkFileSize:     s.NetworkFileSize,
	}
}

// Collector runs queries and publishes the result.
type Collector struct {
	Config   *config.Config
	Log      *logx.Logger
	Version  string
	HTTP     *http.Client // used for forwarding; nil means http.DefaultClient
	CoreHTTP *http.Client // used for Core queries; nil means a fresh client per request
	OnInfo   func(line string)
	OnStatus func(rows map[string]string)

	mu sync.Mutex // serializes Run
}

// Run performs one query and forwards the result. It returns the Core error, if any. Forward
// errors are logged only and never fail the run. Runs never overlap.
func (c *Collector) Run(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	client := &core.Client{
		Host:     c.Config.CoreHost,
		Port:     c.Config.CorePort,
		Password: c.Config.PasswordHash(),
		Timeout:  c.Config.CoreTimeout,
		HTTP:     c.CoreHTTP,
	}
	snap, err := client.Fetch(ctx)
	if err != nil {
		c.Log.Error("%s", err.Error())
		if c.OnStatus != nil {
			c.OnStatus(StatusRows(nil, false))
		}
		return err
	}

	values := Placeholders(snap)
	line := Replace(c.Config.InfoLine, values)
	c.Log.Info("%s", line)
	if c.OnInfo != nil {
		c.OnInfo(line)
	}
	if c.OnStatus != nil {
		c.OnStatus(StatusRows(values, true))
	}

	ua := forward.UserAgent(c.Version)
	for _, t := range c.Config.Targets {
		err := forward.Post(ctx, c.HTTP, t.URL, t.Token, ua, Payload(snap, Replace(t.Line, values)))
		if err != nil {
			c.Log.Error("%s", core.MaskPassword(err.Error()))
			continue
		}
		c.Log.Info("forward successful: %s", t.URL)
	}
	return nil
}

// Loop runs once after delay and then every interval until ctx ends.
func (c *Collector) Loop(ctx context.Context, delay, interval time.Duration) {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			_ = c.Run(ctx)
			timer.Reset(interval)
		}
	}
}
