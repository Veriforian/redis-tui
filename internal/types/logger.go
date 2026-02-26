package types

import (
	"io"
	"strings"
	"sync"
)

// LogWriter implements io.Writer for capturing logs
type LogWriter struct {
	mu   sync.Mutex
	logs []string
}

// MaxLogs is the maximum number of log entries to keep
const MaxLogs = 100

// NewLogWriter creates a new LogWriter
func NewLogWriter() *LogWriter {
	return &LogWriter{}
}

// Write implements io.Writer
func (w *LogWriter) Write(p []byte) (n int, err error) {
	logStr := string(p)
	if strings.Contains(logStr, `"level":"DEBUG"`) {
		return len(p), nil
	}
	w.mu.Lock()
	w.logs = append(w.logs, logStr)
	if len(w.logs) > MaxLogs {
		newLogs := make([]string, MaxLogs)
		copy(newLogs, w.logs[len(w.logs)-MaxLogs:])
		w.logs = newLogs
	}
	w.mu.Unlock()
	return len(p), nil
}

// GetLogs returns a snapshot copy of all log entries
func (w *LogWriter) GetLogs() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	cp := make([]string, len(w.logs))
	copy(cp, w.logs)
	return cp
}

// Len returns the number of log entries
func (w *LogWriter) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.logs)
}

var _ io.Writer = &LogWriter{}
