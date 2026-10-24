package sync

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher monitors a file for changes and triggers a callback with debounce.
type FileWatcher struct {
	watcher  *fsnotify.Watcher
	path     string
	debounce time.Duration
	onChange func()
	mu       sync.Mutex
	timer    *time.Timer
	stopped  chan struct{}
}

// NewFileWatcher creates a new file watcher for the given path.
// The onChange callback is invoked after the debounce period following a file modification.
func NewFileWatcher(path string, debounce time.Duration, onChange func()) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	return &FileWatcher{
		watcher:  watcher,
		path:     filepath.Clean(path),
		debounce: debounce,
		onChange: onChange,
		stopped:  make(chan struct{}),
	}, nil
}

// Start begins watching the file for changes. It blocks until Stop is called.
func (fw *FileWatcher) Start() error {
	err := fw.watcher.Add(fw.path)
	if err != nil {
		return fmt.Errorf("adding file to watcher: %w", err)
	}

	log.Printf("watching file for changes: %s (debounce: %s)", fw.path, fw.debounce)

	for {
		select {
		case <-fw.stopped:
			log.Printf("file watcher stopped: %s", fw.path)
			return nil
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("file changed: %s", event.Name)
				fw.debounceTrigger()
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("file watcher error: %v", err)
		}
	}
}

// Stop gracefully stops the file watcher.
func (fw *FileWatcher) Stop() {
	select {
	case <-fw.stopped:
		// Already stopped
		return
	default:
		close(fw.stopped)
		fw.watcher.Close()
	}
}

// debounceTrigger resets the debounce timer and fires onChange after the debounce period.
func (fw *FileWatcher) debounceTrigger() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.timer != nil {
		fw.timer.Stop()
	}
	fw.timer = time.AfterFunc(fw.debounce, func() {
		fw.mu.Lock()
		fw.timer = nil
		fw.mu.Unlock()
		fw.onChange()
	})
}
