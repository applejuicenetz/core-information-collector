// Package core queries the XML API of an appleJuice Core: information.xml (version, system),
// share.xml (shared file count and size) and modified.xml (credits, speeds, transfers, network).
package core

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// WrongPassword is the text a Core answers with (after a redirect) for a wrong password.
const WrongPassword = "wrong password. access denied."

// maxPayload caps the in-memory XML documents (information.xml, modified.xml).
const maxPayload = 128 << 20

// Snapshot is one complete reading of the Core.
type Snapshot struct {
	Version string
	System  string

	Credits         int64
	Connections     int64
	SessionUpload   int64
	SessionDownload int64
	UploadSpeed     int64
	DownloadSpeed   int64

	Uploads        int64
	Downloads      int64
	DownloadsReady int64 // downloads whose "ready" attribute is 0

	ShareFiles int64
	ShareSize  int64

	NetworkUsers    int64
	NetworkFiles    int64
	NetworkFileSize float64 // megabytes, as reported
}

// Client reads from one Core. Host contains the scheme ("http://127.0.0.1"), Port is the XML
// port and Password the MD5 hash of the Core password.
type Client struct {
	Host     string
	Port     string
	Password string
	Timeout  time.Duration

	// HTTP is optional; the default client has no overall timeout (the idle timeout below applies).
	HTTP *http.Client
}

var passwordParam = regexp.MustCompile(`password=[^&\s"]*`)

// MaskPassword hides every password query parameter in a message.
func MaskPassword(message string) string {
	return passwordParam.ReplaceAllString(message, "password=***")
}

// Fetch reads all three documents. The Snapshot is returned only if every request succeeded.
func (c *Client) Fetch(ctx context.Context) (*Snapshot, error) {
	s := &Snapshot{}

	payload, err := c.getString(ctx, "information.xml")
	if err != nil {
		return nil, err
	}
	if err := parseInformation(payload, s); err != nil {
		return nil, fmt.Errorf("information.xml: %w", err)
	}

	if err := c.stream(ctx, "share.xml", func(r io.Reader) error { return parseShare(r, s) }); err != nil {
		return nil, err
	}

	payload, err = c.getString(ctx, "modified.xml")
	if err != nil {
		return nil, err
	}
	if err := parseModified(payload, s); err != nil {
		return nil, fmt.Errorf("modified.xml: %w", err)
	}
	return s, nil
}

func (c *Client) url(document string) string {
	return fmt.Sprintf("%s:%s/xml/%s?password=%s", c.Host, c.Port, document, c.Password)
}

func (c *Client) getString(ctx context.Context, document string) (string, error) {
	var out string
	err := c.stream(ctx, document, func(r io.Reader) error {
		data, err := io.ReadAll(io.LimitReader(r, maxPayload+1))
		if err != nil {
			return err
		}
		if len(data) > maxPayload {
			return errors.New("response too large")
		}
		if bytes.Contains(data, []byte(WrongPassword)) {
			return errors.New(WrongPassword)
		}
		out = string(data)
		return nil
	})
	return out, err
}

// stream performs the request and hands the body to handle. The timeout covers connecting and
// the wait for the next data; a long document that keeps flowing is not cut off. Errors never
// contain the password.
func (c *Client) stream(ctx context.Context, document string, handle func(io.Reader) error) (err error) {
	defer func() {
		if err != nil {
			err = errors.New(MaskPassword(err.Error()))
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var expired atomic.Bool
	idle := time.AfterFunc(c.Timeout, func() { expired.Store(true); cancel() })
	defer idle.Stop()
	fail := func(err error) error {
		if expired.Load() {
			return fmt.Errorf("timeout after %s: %s", c.Timeout, document)
		}
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(document), nil)
	if err != nil {
		return err
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Transport: &http.Transport{
			Proxy:       http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{Timeout: c.Timeout}).DialContext,
		}}
		defer client.CloseIdleConnections()
	}
	resp, err := client.Do(req)
	if err != nil {
		return fail(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Server returned HTTP response code: %d for URL: %s", resp.StatusCode, req.URL)
	}
	body := &idleReader{r: resp.Body, timer: idle, d: c.Timeout}
	if err := handle(body); err != nil {
		return fail(err)
	}
	return nil
}

// idleReader restarts the idle timer on every read.
type idleReader struct {
	r     io.Reader
	timer *time.Timer
	d     time.Duration
}

func (i *idleReader) Read(p []byte) (int, error) {
	n, err := i.r.Read(p)
	if n > 0 {
		i.timer.Reset(i.d)
	}
	return n, err
}

// decoder returns an XML decoder that never resolves entities or DTDs and understands the
// single-byte encodings a Core may declare.
func decoder(r io.Reader) *xml.Decoder {
	d := xml.NewDecoder(r)
	d.Strict = true
	d.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(label) {
		case "utf-8", "utf8", "us-ascii", "ascii":
			return input, nil
		case "iso-8859-1", "iso8859-1", "latin1", "latin-1", "l1":
			return &latin1Reader{r: bufio.NewReader(input)}, nil
		}
		return nil, fmt.Errorf("unsupported encoding %q", label)
	}
	return d
}

// latin1Reader converts ISO-8859-1 to UTF-8.
type latin1Reader struct {
	r   *bufio.Reader
	buf []byte
}

func (l *latin1Reader) Read(p []byte) (int, error) {
	for len(l.buf) == 0 {
		b, err := l.r.ReadByte()
		if err != nil {
			return 0, err
		}
		l.buf = append(l.buf[:0], string(rune(b))...)
	}
	n := copy(p, l.buf)
	l.buf = l.buf[n:]
	return n, nil
}

func attr(se xml.StartElement, name string) (string, bool) {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value, true
		}
	}
	return "", false
}

func intAttr(se xml.StartElement, name string) (int64, error) {
	v, ok := attr(se, name)
	if !ok {
		return 0, fmt.Errorf("<%s>: attribute %q missing", se.Name.Local, name)
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("<%s>: attribute %q: invalid number %q", se.Name.Local, name, v)
	}
	return n, nil
}

func parseInformation(payload string, s *Snapshot) error {
	d := decoder(strings.NewReader(payload))
	var path []string
	var haveVersion, haveSystem bool
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			path = append(path, t.Name.Local)
			if len(path) == 3 && path[0] == "applejuice" && path[1] == "generalinformation" {
				switch t.Name.Local {
				case "version", "system":
					var text string
					if err := d.DecodeElement(&text, &t); err != nil {
						return err
					}
					path = path[:len(path)-1]
					if t.Name.Local == "version" {
						s.Version, haveVersion = text, true
					} else {
						s.System, haveSystem = text, true
					}
				}
			}
		case xml.EndElement:
			path = path[:len(path)-1]
		}
	}
	if !haveVersion {
		return errors.New("/applejuice/generalinformation/version missing")
	}
	if !haveSystem {
		return errors.New("/applejuice/generalinformation/system missing")
	}
	return nil
}

// parseShare counts <share> elements and sums their size attribute without building a tree.
func parseShare(r io.Reader, s *Snapshot) error {
	in := bufio.NewReader(r)
	if head, _ := in.Peek(128); bytes.Contains(head, []byte(WrongPassword)) {
		return errors.New(WrongPassword)
	}
	d := decoder(in)
	var files, size int64
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "share" {
			n, err := intAttr(se, "size")
			if err != nil {
				return err
			}
			files++
			size += n
		}
	}
	s.ShareFiles, s.ShareSize = files, size
	return nil
}

func parseModified(payload string, s *Snapshot) error {
	d := decoder(strings.NewReader(payload))
	var depth int
	var rootOK bool
	var info, network *xml.StartElement
	var uploads, downloads, ready int64
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				rootOK = t.Name.Local == "applejuice"
			}
			switch t.Name.Local {
			case "upload":
				uploads++
			case "download":
				downloads++
			}
			if depth == 2 && rootOK {
				switch t.Name.Local {
				case "information":
					if info == nil {
						c := t.Copy()
						info = &c
					}
				case "networkinfo":
					if network == nil {
						c := t.Copy()
						network = &c
					}
				case "download":
					if v, ok := attr(t, "ready"); ok {
						if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && f == 0 {
							ready++
						}
					}
				}
			}
		case xml.EndElement:
			depth--
		}
	}
	if !rootOK || info == nil {
		return errors.New("/applejuice/information missing")
	}
	if network == nil {
		return errors.New("/applejuice/networkinfo missing")
	}

	fields := []struct {
		el   *xml.StartElement
		name string
		dst  *int64
	}{
		{info, "credits", &s.Credits},
		{info, "openconnections", &s.Connections},
		{info, "sessionupload", &s.SessionUpload},
		{info, "sessiondownload", &s.SessionDownload},
		{info, "uploadspeed", &s.UploadSpeed},
		{info, "downloadspeed", &s.DownloadSpeed},
		{network, "users", &s.NetworkUsers},
		{network, "files", &s.NetworkFiles},
	}
	for _, f := range fields {
		n, err := intAttr(*f.el, f.name)
		if err != nil {
			return err
		}
		*f.dst = n
	}
	raw, ok := attr(*network, "filesize")
	if !ok {
		return errors.New("<networkinfo>: attribute \"filesize\" missing")
	}
	size, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64)
	if err != nil || math.IsNaN(size) || math.IsInf(size, 0) {
		return fmt.Errorf("<networkinfo>: attribute \"filesize\": invalid number %q", raw)
	}
	s.NetworkFileSize = size
	s.Uploads, s.Downloads, s.DownloadsReady = uploads, downloads, ready
	return nil
}
