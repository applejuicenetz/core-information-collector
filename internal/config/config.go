// Package config reads and creates collector.xml.
package config

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/applejuicenetz/core-information-collector/internal/logx"
)

// File names inside the configuration directory.
const (
	FileName       = "collector.xml"
	LegacyFileName = "core-information-collector.xml"
	LogFileName    = "collector.log"
)

// Defaults of the configuration file.
const (
	DefaultTrayIcon    = "true"
	DefaultTaskbarIcon = "true"
	DefaultInfoLine    = "Credits %coreCredits% - Uploaded %coreSessionUpload% - Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download %coreDownloadSpeed%"
	DefaultInterval    = 60000
	DefaultCoreHost    = "http://127.0.0.1"
	DefaultCorePort    = "9851"
	DefaultTimeoutMS   = 5000
	DefaultForwardLine = "Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%` - Share `%shareFiles%` Files (`%shareSize%`)"
	DefaultForwardURL  = "https://discord.applejuicenet.cc/api/core-collector"

	// EmptyPasswordMD5 is the MD5 of the empty password, used when no password is configured.
	EmptyPasswordMD5 = "d41d8cd98f00b204e9800998ecf8427e"

	// oldForwardHost marks a retired forward URL that is replaced by DefaultForwardURL.
	oldForwardHost = "5f297e.online-server.cloud"
)

// Target is one forward destination.
type Target struct {
	URL   string
	Token string
	Line  string
}

// Config is the parsed configuration.
type Config struct {
	TrayIcon     bool
	TaskbarIcon  bool
	InfoLine     string
	Interval     time.Duration
	CoreHost     string
	CorePort     string
	CorePassword string // as configured; empty means the empty password
	CoreTimeout  time.Duration
	Targets      []Target
}

// PasswordHash returns the configured password hash, or the hash of the empty password.
func (c *Config) PasswordHash() string {
	if c.CorePassword != "" {
		return c.CorePassword
	}
	return EmptyPasswordMD5
}

// DefaultDir returns <home>/appleJuice/collector.
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "appleJuice", "collector"), nil
}

// xml document model
type xmlConfig struct {
	XMLName     xml.Name `xml:"collector"`
	Interval    string   `xml:"intervall,attr"`
	TrayIcon    string   `xml:"trayIcon,attr"`
	TaskbarIcon string   `xml:"taskbarIcon,attr"`
	InfoLine    *string  `xml:"infoLine"`
	Core        *xmlCore `xml:"core"`
	Targets     []xmlTgt `xml:"targets>target"`
}

type xmlCore struct {
	Host     string `xml:"host,attr"`
	Port     string `xml:"port,attr"`
	Password string `xml:"password,attr"`
	Timeout  string `xml:"timeout,attr"`
}

type xmlTgt struct {
	URL   *string `xml:"url"`
	Token string  `xml:"token"`
	Line  *string `xml:"line"`
}

// Path resolves the configuration file in dir. A legacy file name is renamed to collector.xml,
// unless it is read-only (then it is used in place). A missing file is created with defaults.
func Path(dir string, log *logx.Logger) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	file := filepath.Join(dir, FileName)
	if _, err := os.Stat(file); err == nil {
		return file, nil
	}
	legacy := filepath.Join(dir, LegacyFileName)
	if fi, err := os.Stat(legacy); err == nil && fi.Mode().IsRegular() {
		if !writable(legacy) {
			log.Info("%s ist schreibgeschützt und wird nicht nach %s verschoben", LegacyFileName, FileName)
			return legacy, nil
		}
		if err := os.Rename(legacy, file); err != nil {
			log.Error("%v", err)
			return legacy, nil
		}
		return file, nil
	}
	if err := WriteDefault(file); err != nil {
		return "", err
	}
	return file, nil
}

func writable(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// WriteDefault writes a configuration with default values.
func WriteDefault(file string) error {
	doc := xmlConfig{
		Interval:    strconv.Itoa(DefaultInterval),
		TrayIcon:    DefaultTrayIcon,
		TaskbarIcon: DefaultTaskbarIcon,
		InfoLine:    ptr(DefaultInfoLine),
		Core: &xmlCore{
			Host:     DefaultCoreHost,
			Port:     DefaultCorePort,
			Password: "",
			Timeout:  strconv.Itoa(DefaultTimeoutMS),
		},
		Targets: []xmlTgt{{URL: ptr(DefaultForwardURL), Token: "", Line: ptr(DefaultForwardLine)}},
	}
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, append([]byte(xml.Header), append(out, '\n')...), 0o644)
}

func ptr(s string) *string { return &s }

// Load reads the configuration file. Invalid or missing numbers fall back to the defaults and
// are logged; a file that is not valid XML, or a target without URL, is an error.
func Load(file string, log *logx.Logger) (*Config, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var doc xmlConfig
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(file), err)
	}

	cfg := &Config{
		TrayIcon:    flag(doc.TrayIcon),
		TaskbarIcon: flag(doc.TaskbarIcon),
		InfoLine:    DefaultInfoLine,
		Interval:    time.Duration(positive(log, "intervall", doc.Interval, DefaultInterval)) * time.Millisecond,
		CoreHost:    DefaultCoreHost,
		CorePort:    DefaultCorePort,
		CoreTimeout: time.Duration(DefaultTimeoutMS) * time.Millisecond,
	}
	if doc.InfoLine != nil {
		cfg.InfoLine = *doc.InfoLine
	}
	if doc.Core != nil {
		if doc.Core.Host != "" {
			cfg.CoreHost = doc.Core.Host
		}
		if doc.Core.Port != "" {
			cfg.CorePort = doc.Core.Port
		}
		cfg.CorePassword = doc.Core.Password
		cfg.CoreTimeout = time.Duration(positive(log, "core timeout", doc.Core.Timeout, DefaultTimeoutMS)) * time.Millisecond
	}
	for i, t := range doc.Targets {
		if t.URL == nil || strings.TrimSpace(*t.URL) == "" {
			return nil, fmt.Errorf("%s: target %d hat keine url", filepath.Base(file), i+1)
		}
		u := *t.URL
		if strings.Contains(u, oldForwardHost) {
			log.Info("Update forward URL from %s to %s", u, DefaultForwardURL)
			u = DefaultForwardURL
		}
		line := DefaultForwardLine
		if t.Line != nil {
			line = *t.Line
		}
		cfg.Targets = append(cfg.Targets, Target{URL: u, Token: t.Token, Line: line})
	}
	return cfg, nil
}

// flag reports an icon switch: "true" and a missing/empty attribute are on.
func flag(v string) bool { return v == "true" || v == "" }

// positive parses a positive 32-bit integer; anything else logs and returns def.
func positive(log *logx.Logger, name, value string, def int) int {
	if strings.TrimSpace(value) == "" {
		return def
	}
	n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	if err == nil && n > 0 {
		return int(n)
	}
	log.Error("Ungültiger Wert '%s' für %s, verwende %d", value, name, def)
	return def
}
