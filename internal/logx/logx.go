// Package logx writes log lines to an output (stdout) and errors additionally to a log file
// that is rotated once it exceeds MaxFileSize.
package logx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MaxFileSize is the size at which the log file is moved to <name>.old.
const MaxFileSize = 1 << 20

// Logger is safe for concurrent use. The zero value discards everything.
type Logger struct {
	mu   sync.Mutex
	out  io.Writer
	file string
	now  func() time.Time
}

// New returns a Logger writing "info: ..." and "error: ..." lines to out and error lines
// with a timestamp to file. Either may be empty/nil.
func New(out io.Writer, file string) *Logger {
	return &Logger{out: out, file: file, now: time.Now}
}

// Info logs a line to the output only.
func (l *Logger) Info(format string, args ...any) {
	l.write("info", false, format, args)
}

// Error logs a line to the output and the log file.
func (l *Logger) Error(format string, args ...any) {
	l.write("error", true, format, args)
}

func (l *Logger) write(level string, toFile bool, format string, args []any) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.out != nil {
		fmt.Fprintf(l.out, "%s: %s\n", level, msg)
	}
	if toFile && l.file != "" {
		l.appendFile(fmt.Sprintf("%s %s: %s\n", l.now().Format("2006-01-02 15:04:05"), level, msg))
	}
}

func (l *Logger) appendFile(line string) {
	if fi, err := os.Stat(l.file); err == nil && fi.Size() >= MaxFileSize {
		_ = os.Rename(l.file, l.file+".old")
	}
	if err := os.MkdirAll(filepath.Dir(l.file), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}
