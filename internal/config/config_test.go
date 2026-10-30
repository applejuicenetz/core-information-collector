package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/applejuicenetz/core-information-collector/internal/logx"
)

func quiet() (*logx.Logger, *bytes.Buffer) {
	var b bytes.Buffer
	return logx.New(&b, ""), &b
}

func write(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), FileName)
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return file
}

func doc(interval, timeout, password string) string {
	return `<collector intervall="` + interval + `" trayIcon="false" taskbarIcon="false">` +
		`<infoLine>test</infoLine><core host="http://127.0.0.1" port="9851" password="` + password +
		`" timeout="` + timeout + `"/></collector>`
}

func TestExplicitValues(t *testing.T) {
	log, _ := quiet()
	cfg, err := Load(write(t, doc("12345", "2345", "test-hash")), log)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Interval != 12345*time.Millisecond || cfg.CoreTimeout != 2345*time.Millisecond {
		t.Errorf("interval %v timeout %v", cfg.Interval, cfg.CoreTimeout)
	}
	if cfg.PasswordHash() != "test-hash" || cfg.TrayIcon || cfg.TaskbarIcon || len(cfg.Targets) != 0 {
		t.Errorf("%+v", cfg)
	}
	if cfg.InfoLine != "test" || cfg.CoreHost != "http://127.0.0.1" || cfg.CorePort != "9851" {
		t.Errorf("%+v", cfg)
	}
}

func TestInvalidNumbersFallBackToDefaults(t *testing.T) {
	for _, v := range []string{"", "0", "-1", "bad", "2147483648", " "} {
		log, out := quiet()
		cfg, err := Load(write(t, doc(v, v, "")), log)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Interval != 60*time.Second || cfg.CoreTimeout != 5*time.Second {
			t.Errorf("%q: interval %v timeout %v", v, cfg.Interval, cfg.CoreTimeout)
		}
		if strings.TrimSpace(v) != "" && !strings.Contains(out.String(), "Ungültiger Wert") {
			t.Errorf("%q: no log line: %q", v, out.String())
		}
	}
}

func TestEmptyPasswordUsesEmptyMD5(t *testing.T) {
	log, _ := quiet()
	cfg, _ := Load(write(t, doc("5000", "5000", "")), log)
	if cfg.PasswordHash() != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("%s", cfg.PasswordHash())
	}
}

func TestIconFlags(t *testing.T) {
	log, _ := quiet()
	for attr, want := range map[string]bool{`trayIcon="true"`: true, `trayIcon=""`: true, ``: true, `trayIcon="false"`: false, `trayIcon="no"`: false} {
		cfg, err := Load(write(t, `<collector `+attr+`><infoLine>x</infoLine></collector>`), log)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.TrayIcon != want {
			t.Errorf("%q: %v", attr, cfg.TrayIcon)
		}
	}
}

func TestMissingElementsUseDefaults(t *testing.T) {
	log, _ := quiet()
	cfg, err := Load(write(t, `<collector/>`), log)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.InfoLine != DefaultInfoLine || cfg.CoreHost != DefaultCoreHost || cfg.CorePort != DefaultCorePort ||
		cfg.Interval != 60*time.Second || cfg.CoreTimeout != 5*time.Second || !cfg.TrayIcon {
		t.Errorf("%+v", cfg)
	}
}

func TestTargets(t *testing.T) {
	log, out := quiet()
	cfg, err := Load(write(t, `<collector><targets>
<target><url>https://a.example/api</url><token>tok</token><line>`+"`%coreVersion%`"+`</line></target>
<target><url>http://5f297e.online-server.cloud/old</url><token></token><line>l2</line></target>
<target><url>https://c.example</url></target>
</targets></collector>`), log)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Targets) != 3 {
		t.Fatalf("%d targets", len(cfg.Targets))
	}
	if cfg.Targets[0] != (Target{"https://a.example/api", "tok", "`%coreVersion%`"}) {
		t.Errorf("%+v", cfg.Targets[0])
	}
	if cfg.Targets[1].URL != DefaultForwardURL || !strings.Contains(out.String(), "Update forward URL") {
		t.Errorf("old URL not replaced: %+v / %q", cfg.Targets[1], out.String())
	}
	if cfg.Targets[2].Token != "" || cfg.Targets[2].Line != DefaultForwardLine {
		t.Errorf("%+v", cfg.Targets[2])
	}
}

func TestTargetWithoutURLIsAnError(t *testing.T) {
	log, _ := quiet()
	for _, body := range []string{`<target><token>t</token></target>`, `<target><url>  </url></target>`} {
		if _, err := Load(write(t, `<collector><targets>`+body+`</targets></collector>`), log); err == nil {
			t.Errorf("%s: expected error", body)
		}
	}
}

func TestInvalidXMLAndMissingFile(t *testing.T) {
	log, _ := quiet()
	if _, err := Load(write(t, `<collector`), log); err == nil {
		t.Error("broken XML must fail")
	}
	if _, err := Load(write(t, `<other/>`), log); err == nil {
		t.Error("wrong root element must fail")
	}
	if _, err := Load(filepath.Join(t.TempDir(), "none.xml"), log); err == nil {
		t.Error("missing file must fail")
	}
}

func TestInfoLineWithSpecialCharacters(t *testing.T) {
	log, _ := quiet()
	cfg, err := Load(write(t, "<collector><infoLine>Größe &amp; `x` &lt;b&gt;\n  next</infoLine></collector>"), log)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.InfoLine != "Größe & `x` <b>\n  next" {
		t.Errorf("%q", cfg.InfoLine)
	}
}

func TestPathCreatesDefaultFileThatLoads(t *testing.T) {
	log, _ := quiet()
	dir := filepath.Join(t.TempDir(), "appleJuice", "collector")
	file, err := Path(dir, log)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(file) != FileName {
		t.Errorf("%s", file)
	}
	cfg, err := Load(file, log)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TrayIcon || !cfg.TaskbarIcon || cfg.Interval != 60*time.Second || cfg.CoreTimeout != 5*time.Second ||
		cfg.InfoLine != DefaultInfoLine || cfg.CoreHost != DefaultCoreHost || cfg.CorePort != DefaultCorePort ||
		cfg.PasswordHash() != EmptyPasswordMD5 {
		t.Errorf("%+v", cfg)
	}
	if len(cfg.Targets) != 1 || cfg.Targets[0] != (Target{DefaultForwardURL, "", DefaultForwardLine}) {
		t.Errorf("%+v", cfg.Targets)
	}
	raw, _ := os.ReadFile(file)
	if !strings.HasPrefix(string(raw), "<?xml") || !strings.Contains(string(raw), `intervall="60000"`) {
		t.Errorf("%s", raw)
	}
}

func TestPathKeepsExistingFile(t *testing.T) {
	log, _ := quiet()
	dir := t.TempDir()
	file := filepath.Join(dir, FileName)
	os.WriteFile(file, []byte(`<collector intervall="1"/>`), 0o644)
	got, err := Path(dir, log)
	if err != nil || got != file {
		t.Fatal(got, err)
	}
	raw, _ := os.ReadFile(file)
	if string(raw) != `<collector intervall="1"/>` {
		t.Errorf("file overwritten: %s", raw)
	}
}

func TestPathMovesWritableLegacyFile(t *testing.T) {
	log, _ := quiet()
	dir := t.TempDir()
	legacy := filepath.Join(dir, LegacyFileName)
	os.WriteFile(legacy, []byte(`<collector intervall="777"/>`), 0o644)
	file, err := Path(dir, log)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(file) != FileName {
		t.Fatalf("%s", file)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Error("legacy file still present")
	}
	cfg, _ := Load(file, log)
	if cfg.Interval != 777*time.Millisecond {
		t.Errorf("%v", cfg.Interval)
	}
}

func TestPathUsesReadOnlyLegacyInPlace(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write read-only files")
	}
	log, out := quiet()
	dir := t.TempDir()
	legacy := filepath.Join(dir, LegacyFileName)
	os.WriteFile(legacy, []byte(`<collector/>`), 0o444)
	file, err := Path(dir, log)
	if err != nil || file != legacy {
		t.Fatal(file, err)
	}
	if !strings.Contains(out.String(), "schreibgeschützt") {
		t.Errorf("%q", out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, FileName)); err == nil {
		t.Error("collector.xml must not be created next to a read-only legacy file")
	}
}

func TestPathFailsWhenDirectoryIsAFile(t *testing.T) {
	log, _ := quiet()
	blocker := filepath.Join(t.TempDir(), "file")
	os.WriteFile(blocker, nil, 0o644)
	if _, err := Path(filepath.Join(blocker, "sub"), log); err == nil {
		t.Error("expected error")
	}
}

func TestDefaultDir(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	t.Setenv("USERPROFILE", `C:\Users\test`)
	d, err := DefaultDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(d), "appleJuice/collector") {
		t.Errorf("%s", d)
	}
}
