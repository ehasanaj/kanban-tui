// Package watcher provides file system watching capabilities for the kanban board.
package watcher

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event represents a file system event.
type Event struct {
	Path string
	Op   fsnotify.Op
}

// Watcher watches directories for file changes with debouncing.
type Watcher struct {
	watcher     *fsnotify.Watcher
	Events      chan Event
	Errors      chan error
	debounce    time.Duration
	pending     map[string]*time.Timer
	pendingLock sync.Mutex
	done        chan struct{}
}

// New creates a new Watcher with the specified debounce duration.
func New(debounce time.Duration) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:  fsWatcher,
		Events:   make(chan Event, 100),
		Errors:   make(chan error, 10),
		debounce: debounce,
		pending:  make(map[string]*time.Timer),
		done:     make(chan struct{}),
	}

	go w.run()

	return w, nil
}

// Add adds a directory to watch.
func (w *Watcher) Add(path string) error {
	return w.watcher.Add(path)
}

// Remove stops watching a directory.
func (w *Watcher) Remove(path string) error {
	return w.watcher.Remove(path)
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	close(w.done)
	return w.watcher.Close()
}

// run processes file system events.
func (w *Watcher) run() {
	for {
		select {
		case <-w.done:
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only process markdown files
			if filepath.Ext(event.Name) != ".md" {
				continue
			}

			w.debounceEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			select {
			case w.Errors <- err:
			default:
				// Drop error if channel is full
			}
		}
	}
}

// debounceEvent debounces file events to avoid rapid-fire updates.
func (w *Watcher) debounceEvent(event fsnotify.Event) {
	w.pendingLock.Lock()
	defer w.pendingLock.Unlock()

	// Cancel any pending timer for this file
	if timer, exists := w.pending[event.Name]; exists {
		timer.Stop()
	}

	// Create new timer
	w.pending[event.Name] = time.AfterFunc(w.debounce, func() {
		w.pendingLock.Lock()
		delete(w.pending, event.Name)
		w.pendingLock.Unlock()

		select {
		case w.Events <- Event{Path: event.Name, Op: event.Op}:
		case <-w.done:
		}
	})
}
