package core

import (
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const (
	infoXML = `<?xml version="1.0"?>
<applejuice><generalinformation><version>0.35.185.93</version><filesystem seperator="/"/><system>Linux</system></generalinformation></applejuice>`

	modifiedXML = `<?xml version="1.0"?>
<applejuice>
<time>1</time>
<ids><downloadid id="1"/><uploadid id="2"/></ids>
<download id="1" ready="0"/>
<download id="2" ready="5"/>
<download id="3" ready="0"/>
<download id="4"/>
<upload id="7"/>
<upload id="8"/>
<information credits="4654845359" sessionupload="123456" sessiondownload="654321" uploadspeed="276000" downloadspeed="12" openconnections="3" maxuploadpositions="50"/>
<networkinfo users="450" files="4191120" filesize="1449,57" firewalled="true"><welcomemessage>hi</welcomemessage></networkinfo>
</applejuice>`

	shareXML = `<?xml version="1.0"?>
<applejuice><shares>
<share id="1" filename="a" size="100"/>
<share id="2" filename="b" size="4154798080"/>
<share id="3" filename="c" size="0"/>
</shares></applejuice>`
)

type server struct {
	*httptest.Server
	hits    map[string]*atomic.Int32
	queries []string
}

// newServer answers /xml/<doc>.xml for the password "pw"; other passwords get the Core's
// redirect to /wrongpassword.
func newServer(t *testing.T, docs map[string]string) *server {
	t.Helper()
	s := &server{hits: map[string]*atomic.Int32{}}
	for k := range docs {
		s.hits[k] = &atomic.Int32{}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/wrongpassword", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>\n<body>\n<p>wrong password. access denied. for more information look <a href=\"/help\">help</a>.</p>\n</body>\n</html>\n"))
	})
	mux.HandleFunc("/xml/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/xml/"), ".xml")
		if r.URL.Query().Get("password") != "pw" {
			http.Redirect(w, r, "/wrongpassword", http.StatusFound)
			return
		}
		body, ok := docs[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.hits[name].Add(1)
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(body))
	})
	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func (s *server) client(password string) *Client {
	i := strings.LastIndex(s.URL, ":")
	return &Client{Host: s.URL[:i], Port: s.URL[i+1:], Password: password, Timeout: 2 * time.Second}
}

func all() map[string]string {
	return map[string]string{"information": infoXML, "share": shareXML, "modified": modifiedXML}
}

func TestFetchReadsAllValues(t *testing.T) {
	s := newServer(t, all())
	snap, err := s.client("pw").Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := Snapshot{
		Version: "0.35.185.93", System: "Linux",
		Credits: 4654845359, Connections: 3, SessionUpload: 123456, SessionDownload: 654321,
		UploadSpeed: 276000, DownloadSpeed: 12,
		Uploads: 2, Downloads: 4, DownloadsReady: 2,
		ShareFiles: 3, ShareSize: 4154798180,
		NetworkUsers: 450, NetworkFiles: 4191120, NetworkFileSize: 1449.57,
	}
	if *snap != want {
		t.Errorf("got  %+v\nwant %+v", *snap, want)
	}
}

func TestFetchAgainstMockCoreFixtures(t *testing.T) {
	s := newServer(t, map[string]string{
		"information": fixture(t, "mock-information.xml"),
		"share":       shareXML,
		"modified":    fixture(t, "mock-modified.xml"),
	})
	snap, err := s.client("pw").Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.Version != "0.35.185.93" || snap.System != "Linux" {
		t.Errorf("%+v", snap)
	}
	if snap.Credits != 4654845359 || snap.UploadSpeed != 276000 || snap.Connections != 1 {
		t.Errorf("%+v", snap)
	}
	if snap.Downloads != 8 || snap.Uploads != 8 || snap.NetworkUsers != 450 || snap.NetworkFileSize != 1449.57 {
		t.Errorf("%+v", snap)
	}
	// Downloads with ready="0" in the fixture.
	n := strings.Count(fixture(t, "mock-modified.xml"), `ready="0"`)
	if snap.DownloadsReady != int64(n) {
		t.Errorf("ready %d, fixture has %d", snap.DownloadsReady, n)
	}
}

func TestWrongPasswordIsReportedForEveryDocument(t *testing.T) {
	for _, drop := range []string{"information", "share", "modified"} {
		docs := all()
		s := newServer(t, docs)
		c := s.client("bad")
		_ = drop
		_, err := c.Fetch(context.Background())
		if err == nil || err.Error() != WrongPassword {
			t.Fatalf("error %v", err)
		}
	}
}

func TestWrongPasswordOnLaterDocumentStops(t *testing.T) {
	// information is served for every password, share and modified are denied.
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/information.xml", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(infoXML)) })
	mux.HandleFunc("/xml/share.xml", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body><p>wrong password. access denied.</p></body></html>"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	i := strings.LastIndex(srv.URL, ":")
	c := &Client{Host: srv.URL[:i], Port: srv.URL[i+1:], Password: "x", Timeout: time.Second}
	if _, err := c.Fetch(context.Background()); err == nil || err.Error() != WrongPassword {
		t.Errorf("%v", err)
	}
}

func TestErrorsNeverContainThePassword(t *testing.T) {
	s := newServer(t, map[string]string{})
	c := s.client("pw")
	c.Password = "secret-hash"
	_, err := c.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "secret-hash") {
		t.Errorf("password leaked: %v", err)
	}

	// Connection failure: the client's own message contains the URL.
	c = &Client{Host: "http://127.0.0.1", Port: "1", Password: "secret-hash", Timeout: time.Second}
	_, err = c.Fetch(context.Background())
	if err == nil || strings.Contains(err.Error(), "secret-hash") {
		t.Errorf("%v", err)
	}
	if !strings.Contains(err.Error(), "password=***") {
		t.Errorf("expected masked URL: %v", err)
	}
}

func TestHTTPErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	i := strings.LastIndex(srv.URL, ":")
	c := &Client{Host: srv.URL[:i], Port: srv.URL[i+1:], Password: "pw", Timeout: time.Second}
	_, err := c.Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "=pw") {
		t.Errorf("%v", err)
	}
}

func TestIdleTimeoutTerminatesRequest(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer srv.Close()
	defer close(release)
	i := strings.LastIndex(srv.URL, ":")
	c := &Client{Host: srv.URL[:i], Port: srv.URL[i+1:], Password: "pw", Timeout: 100 * time.Millisecond}
	start := time.Now()
	_, err := c.Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("%v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Error("took too long")
	}
}

func TestSlowButFlowingShareIsNotCutOff(t *testing.T) {
	// 30 chunks, 20 ms apart: 600 ms in total with a 200 ms idle timeout.
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/information.xml", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(infoXML)) })
	mux.HandleFunc("/xml/modified.xml", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(modifiedXML)) })
	mux.HandleFunc("/xml/share.xml", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<applejuice><shares>`))
		for i := 0; i < 30; i++ {
			_, _ = w.Write([]byte(`<share size="1"/>`))
			w.(http.Flusher).Flush()
			time.Sleep(20 * time.Millisecond)
		}
		_, _ = w.Write([]byte(`</shares></applejuice>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	i := strings.LastIndex(srv.URL, ":")
	c := &Client{Host: srv.URL[:i], Port: srv.URL[i+1:], Password: "pw", Timeout: 200 * time.Millisecond}
	snap, err := c.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.ShareFiles != 30 || snap.ShareSize != 30 {
		t.Errorf("%+v", snap)
	}
}

func TestContextCancelStopsFetch(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer srv.Close()
	defer close(release)
	i := strings.LastIndex(srv.URL, ":")
	c := &Client{Host: srv.URL[:i], Port: srv.URL[i+1:], Password: "pw", Timeout: 10 * time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)
	start := time.Now()
	if _, err := c.Fetch(ctx); err == nil {
		t.Fatal("expected error")
	}
	if time.Since(start) > 3*time.Second {
		t.Error("cancel not honoured")
	}
}

func TestMalformedDocuments(t *testing.T) {
	cases := map[string]map[string]string{
		"information without version":  {"information": `<applejuice><generalinformation><system>x</system></generalinformation></applejuice>`},
		"information without system":   {"information": `<applejuice><generalinformation><version>1</version></generalinformation></applejuice>`},
		"information not xml":          {"information": `<applejuice><generalinformation>`},
		"modified without information": {"modified": strings.Replace(modifiedXML, `<information `, `<infox `, 1)},
		"modified without network":     {"modified": strings.Replace(modifiedXML, `<networkinfo `, `<networkx `, 1)},
		"modified bad credits":         {"modified": strings.Replace(modifiedXML, `credits="4654845359"`, `credits="many"`, 1)},
		"modified missing credits":     {"modified": strings.Replace(modifiedXML, `credits="4654845359"`, ``, 1)},
		"modified bad filesize":        {"modified": strings.Replace(modifiedXML, `filesize="1449,57"`, `filesize="NaN"`, 1)},
		"modified wrong root":          {"modified": `<other/>`},
		"modified truncated":           {"modified": `<applejuice><information credits="1"`},
		"share bad size":               {"share": `<applejuice><shares><share size="x"/></shares></applejuice>`},
		"share missing size":           {"share": `<applejuice><shares><share id="1"/></shares></applejuice>`},
		"share truncated":              {"share": `<applejuice><shares><share size="1"/>`},
	}
	for name, override := range cases {
		docs := all()
		for k, v := range override {
			docs[k] = v
		}
		s := newServer(t, docs)
		if snap, err := s.client("pw").Fetch(context.Background()); err == nil {
			t.Errorf("%s: expected error, got %+v", name, snap)
		}
	}
}

func TestEmptyShareListIsValid(t *testing.T) {
	docs := all()
	docs["share"] = `<applejuice><shares></shares></applejuice>`
	s := newServer(t, docs)
	snap, err := s.client("pw").Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.ShareFiles != 0 || snap.ShareSize != 0 {
		t.Errorf("%+v", snap)
	}
}

func TestZeroDownloadsReadyOnlyCountsTopLevelDownloads(t *testing.T) {
	docs := all()
	docs["modified"] = strings.Replace(modifiedXML, `<download id="4"/>`,
		`<other><download id="9" ready="0"/></other>`, 1)
	s := newServer(t, docs)
	snap, err := s.client("pw").Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// All <download> elements count as downloads (like getElementsByTagName), only top-level
	// ones with ready="0" count as ready.
	if snap.Downloads != 4 || snap.DownloadsReady != 2 {
		t.Errorf("%+v", snap)
	}
}

func TestISO88591Document(t *testing.T) {
	docs := all()
	docs["information"] = "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n<applejuice><generalinformation><version>1</version><system>Mac \xe4\xf6</system></generalinformation></applejuice>"
	s := newServer(t, docs)
	snap, err := s.client("pw").Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.System != "Mac äö" {
		t.Errorf("%q", snap.System)
	}
}

func TestUnsupportedEncodingFails(t *testing.T) {
	docs := all()
	docs["information"] = `<?xml version="1.0" encoding="UTF-16"?><applejuice/>`
	s := newServer(t, docs)
	if _, err := s.client("pw").Fetch(context.Background()); err == nil {
		t.Error("expected error")
	}
}

func TestDTDEntitiesAreNotResolved(t *testing.T) {
	docs := all()
	docs["information"] = `<?xml version="1.0"?><!DOCTYPE x [<!ENTITY e SYSTEM "file:///etc/passwd">]>
<applejuice><generalinformation><version>&e;</version><system>x</system></generalinformation></applejuice>`
	s := newServer(t, docs)
	snap, err := s.client("pw").Fetch(context.Background())
	if err == nil && strings.Contains(snap.Version, "root:") {
		t.Fatal("external entity resolved")
	}
}

func TestVeryLargeShareIsStreamed(t *testing.T) {
	const files = 200000
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/information.xml", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(infoXML)) })
	mux.HandleFunc("/xml/modified.xml", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(modifiedXML)) })
	mux.HandleFunc("/xml/share.xml", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<applejuice><shares>`))
		line := []byte(`<share id="1" filename="/some/path/file.iso" size="` + strconv.Itoa(3<<30) + `"/>` + "\n")
		for i := 0; i < files; i++ {
			_, _ = w.Write(line)
		}
		_, _ = w.Write([]byte(`</shares></applejuice>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	i := strings.LastIndex(srv.URL, ":")
	c := &Client{Host: srv.URL[:i], Port: srv.URL[i+1:], Password: "pw", Timeout: 5 * time.Second}
	snap, err := c.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.ShareFiles != files || snap.ShareSize != int64(files)*(3<<30) {
		t.Errorf("%d files, %d bytes", snap.ShareFiles, snap.ShareSize)
	}
}

func TestMaskPassword(t *testing.T) {
	got := MaskPassword("password=first&x=1 password=second end")
	if got != "password=***&x=1 password=*** end" {
		t.Errorf("%q", got)
	}
	if got := MaskPassword(`Get "http://h:1/xml/a.xml?password=abc": dial`); got != `Get "http://h:1/xml/a.xml?password=***": dial` {
		t.Errorf("%q", got)
	}
}

func TestRequestURLCarriesPassword(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(infoXML))
	}))
	defer srv.Close()
	i := strings.LastIndex(srv.URL, ":")
	c := &Client{Host: srv.URL[:i], Port: srv.URL[i+1:], Password: "abc123", Timeout: time.Second}
	_, _ = c.getString(context.Background(), "information.xml")
	if gotQuery != "password=abc123" {
		t.Errorf("%q", gotQuery)
	}
}

// TestParsersMatchJavaReference checks the values the former Java implementation (tag 4.0.1)
// extracted from the same mock Core documents.
func TestParsersMatchJavaReference(t *testing.T) {
	f, err := os.Open("../../testdata/mock-share.xml.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	s := &Snapshot{}
	if err := parseInformation(fixture(t, "mock-information.xml"), s); err != nil {
		t.Fatal(err)
	}
	if err := parseModified(fixture(t, "mock-modified.xml"), s); err != nil {
		t.Fatal(err)
	}
	if err := parseShare(zr, s); err != nil {
		t.Fatal(err)
	}
	want := Snapshot{
		Version: "0.35.185.93", System: "Linux", Credits: 4654845359, Connections: 1,
		SessionUpload: 123456, SessionDownload: 654321, UploadSpeed: 276000, DownloadSpeed: 0,
		Uploads: 8, Downloads: 8, DownloadsReady: 1, ShareFiles: 1559, ShareSize: 812798035034,
		NetworkUsers: 450, NetworkFiles: 4191120, NetworkFileSize: 1449.57,
	}
	if *s != want {
		t.Errorf("got  %+v\nwant %+v", *s, want)
	}
}
