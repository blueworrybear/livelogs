package stream

import (
	"context"
	"sync"

	"github.com/blueworrybear/livelogs/core"
)

const bufferSize = 5000

type streamer struct {
	sync.Mutex
	lines    []*core.LogLine
	watchers map[*watcher]struct{}
}

func newStreamer() *streamer {
	s := &streamer{
		lines:    make([]*core.LogLine, 0),
		watchers: make(map[*watcher]struct{}),
	}
	return s
}

func (s *streamer) watch(ctx context.Context) <-chan *core.LogLine {
	w := newWatcher()
	s.Lock()
	for _, line := range s.lines {
		w.notify((line))
	}
	s.watchers[w] = struct{}{}
	s.Unlock()
	go func() {
		select {
		case <-w.closed:
		case <-ctx.Done():
			w.close()
		}
	}()
	return w.buffer
}

func (s *streamer) write(line *core.LogLine) error {
	s.Lock()
	defer s.Unlock()
	s.lines = append(s.lines, line)
	for w := range s.watchers {
		w.notify(line)
	}
	if size := len(s.lines); size > bufferSize {
		s.lines = s.lines[size-bufferSize:]
	}
	return nil
}

func (s *streamer) close() error {
	s.Lock()
	defer s.Unlock()
	for w := range s.watchers {
		delete(s.watchers, w)
		w.close()
	}
	return nil
}

type watcher struct {
	sync.Mutex
	buffer chan *core.LogLine
	closed chan struct{}
}

func newWatcher() *watcher {
	return &watcher{
		buffer: make(chan *core.LogLine, bufferSize),
		closed: make(chan struct{}),
	}
}

func (w *watcher) notify(line *core.LogLine) {
	select {
	case <-w.closed:
	case w.buffer <- line:
	default:
	}
}

func (w *watcher) close() {
	w.Lock()
	defer w.Unlock()
	select {
	case <-w.closed:
		// Already closed
	default:
		close(w.closed)
	}
}
