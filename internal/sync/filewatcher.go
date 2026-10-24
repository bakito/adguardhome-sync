package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"

	"github.com/bakito/adguardhome-sync/internal/log"
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
	l        *zap.SugaredLogger
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
		l:        log.GetLogger("filewatcher").With("file", path),
	}, nil
}

// Start begins watching the file for changes. It blocks until Stop is called.
// If the path is a file or symlink, it watches the parent directory to handle
// Kubernetes ConfigMap updates that remove and recreate the file.
// If the path is a symlink, it also watches the target file's parent directory.
func (fw *FileWatcher) Start() error {
	fw.l.Info("starting file watcher")

	// Watch the parent directory if the path is a file or symlink, to handle
	// ConfigMap updates that remove and recreate the file (common in Kubernetes).
	// Use Lstat to detect symlinks without following them.
	watchPath := fw.path
	var symlinkTarget string
	if info, err := os.Lstat(fw.path); err == nil && !info.IsDir() {
		watchPath = filepath.Dir(fw.path)
		// If it's a symlink, also watch the target file
		if info.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Readlink(fw.path); err == nil {
				// Resolve relative symlinks
				if !filepath.IsAbs(target) {
					target = filepath.Join(filepath.Dir(fw.path), target)
				}
				symlinkTarget = filepath.Clean(target)
			}
		}
	}

	err := fw.watcher.Add(watchPath)
	if err != nil {
		return fmt.Errorf("adding to watcher: %w", err)
	}

	// If watching a symlink, also watch the target's parent directory
	var targetWatchPath string
	if symlinkTarget != "" {
		targetWatchPath = filepath.Dir(symlinkTarget)
		if targetWatchPath != watchPath {
			if err := fw.watcher.Add(targetWatchPath); err != nil {
				fw.l.With("target", symlinkTarget, "error", err).Warn("failed to watch symlink target")
			}
		}
	}

	if symlinkTarget != "" {
		fw.l.With("watch", watchPath, "target", symlinkTarget, "targetWatch", targetWatchPath, "debounce", fw.debounce).
			Info("watching file for changes")
	} else {
		fw.l.With("watch", watchPath, "debounce", fw.debounce).Info("watching file for changes")
	}

	for {
		select {
		case <-fw.stopped:
			fw.l.Info("file watcher stopped: %s", fw.path)
			return nil
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}
			// If watching a directory, only trigger for events matching the watched file
			// or its symlink target. When watching a symlink target's parent directory,
			// accept any event, as Kubernetes atomic writes (temp file + RENAME) may
			// fire events for temporary filenames that still affect the watched file.
			// Also accept events in the symlink's parent directory (watchPath), since
			// kubelet may create temp files there during atomic symlink replacement.
			if watchPath != fw.path {
				eventDir := filepath.Dir(event.Name)
				if event.Name != fw.path && event.Name != symlinkTarget && eventDir != targetWatchPath &&
					eventDir != watchPath {
					fw.l.With("event", event.Name, "operation", event.Op).Debug("ignoring event for unrelated path")
					continue
				}
			}
			// Accept all event types to handle Kubernetes ConfigMap updates
			// that remove and recreate files (REMOVE + CREATE) instead of writing in-place.
			if event.Op != 0 {
				fw.l.With("event", event.Name, "operation", event.Op).Debug("file changed")
				fw.debounceTrigger()
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			fw.l.With("error", err).Error("file watcher error")
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
		fw.l.Info("file change debounce elapsed, triggering sync")
		fw.onChange()
	})
}
