package sync

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewFileWatcher(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fw, err := NewFileWatcher(tmpFile, 100*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	if fw.path != tmpFile {
		t.Errorf("expected path %s, got %s", tmpFile, fw.path)
	}
}

func TestFileWatcher_NotFound(t *testing.T) {
	fw, err := NewFileWatcher("/nonexistent/file.txt", 100*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	err = fw.Start()
	if err == nil {
		t.Error("expected error when watching nonexistent file")
	}
}

func TestFileWatcher_Debounce(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var count atomic.Int32
	debounce := 100 * time.Millisecond
	fw, err := NewFileWatcher(tmpFile, debounce, func() {
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	// Start the watcher in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- fw.Start()
	}()

	// Give the watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Write to the file multiple times in quick succession
	for range 5 {
		if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce period plus buffer
	time.Sleep(debounce + 200*time.Millisecond)

	// Stop the watcher
	fw.Stop()
	if err := <-done; err != nil {
		t.Fatalf("file watcher error: %v", err)
	}

	// The callback should have been called only once due to debounce
	c := count.Load()
	if c != 1 {
		t.Errorf("expected 1 callback invocation due to debounce, got %d", c)
	}
}

func TestFileWatcher_MultipleChanges(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var count atomic.Int32
	debounce := 100 * time.Millisecond
	fw, err := NewFileWatcher(tmpFile, debounce, func() {
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	// Start the watcher in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- fw.Start()
	}()

	// Give the watcher time to start
	time.Sleep(50 * time.Millisecond)

	// First change
	if err := os.WriteFile(tmpFile, []byte("test1"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	time.Sleep(debounce + 100*time.Millisecond)

	// Second change after debounce
	if err := os.WriteFile(tmpFile, []byte("test2"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	time.Sleep(debounce + 100*time.Millisecond)

	// Stop the watcher
	fw.Stop()
	if err := <-done; err != nil {
		t.Fatalf("file watcher error: %v", err)
	}

	c := count.Load()
	if c < 2 {
		t.Errorf("expected at least 2 callback invocations, got %d", c)
	}
}

func TestFileWatcher_Stop(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fw, err := NewFileWatcher(tmpFile, 100*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	// Start the watcher in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- fw.Start()
	}()

	// Give the watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Stop the watcher
	fw.Stop()

	// Verify it stopped cleanly
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("file watcher error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("file watcher did not stop in time")
	}
}

func TestFileWatcher_StopIdempotent(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fw, err := NewFileWatcher(tmpFile, 100*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	// Stop twice should not panic
	fw.Stop()
	fw.Stop()
}
