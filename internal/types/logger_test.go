package types

import (
	"sync"
	"testing"
)

func TestLogWriter_Write(t *testing.T) {
	t.Run("appends log entry", func(t *testing.T) {
		w := NewLogWriter()

		input := []byte(`{"level":"INFO","msg":"test"}`)
		n, err := w.Write(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != len(input) {
			t.Errorf("Write returned %d bytes, want %d", n, len(input))
		}
		logs := w.GetLogs()
		if len(logs) != 1 {
			t.Fatalf("expected 1 log entry, got %d", len(logs))
		}
		if logs[0] != `{"level":"INFO","msg":"test"}` {
			t.Errorf("log entry = %q, want %q", logs[0], `{"level":"INFO","msg":"test"}`)
		}
	})

	t.Run("filters DEBUG messages", func(t *testing.T) {
		w := NewLogWriter()

		input := []byte(`{"level":"DEBUG","msg":"debug info"}`)
		n, err := w.Write(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != len(input) {
			t.Errorf("Write returned %d bytes, want %d", n, len(input))
		}
		if w.Len() != 0 {
			t.Errorf("expected 0 log entries for DEBUG, got %d", w.Len())
		}
	})

	t.Run("trims at MaxLogs", func(t *testing.T) {
		w := NewLogWriter()

		for i := 0; i < MaxLogs+5; i++ {
			_, err := w.Write([]byte(`{"level":"INFO","msg":"entry"}`))
			if err != nil {
				t.Fatalf("unexpected error at entry %d: %v", i, err)
			}
		}
		if w.Len() != MaxLogs {
			t.Errorf("expected %d log entries, got %d", MaxLogs, w.Len())
		}
	})

	t.Run("returns correct byte count", func(t *testing.T) {
		w := NewLogWriter()
		input := []byte("some log message")

		n, err := w.Write(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != len(input) {
			t.Errorf("Write returned %d, want %d", n, len(input))
		}
	})

	t.Run("returns correct byte count for filtered DEBUG", func(t *testing.T) {
		w := NewLogWriter()
		input := []byte(`{"level":"DEBUG","msg":"filtered"}`)

		n, err := w.Write(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != len(input) {
			t.Errorf("Write returned %d, want %d", n, len(input))
		}
	})

	t.Run("concurrent writes are safe", func(t *testing.T) {
		w := NewLogWriter()
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = w.Write([]byte(`{"level":"INFO","msg":"concurrent"}`))
			}()
		}
		wg.Wait()
		if w.Len() > MaxLogs {
			t.Errorf("expected at most %d entries, got %d", MaxLogs, w.Len())
		}
	})
}
