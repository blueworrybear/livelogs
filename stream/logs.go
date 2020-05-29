package stream

import (
	"context"
	"errors"
	"sync"

	"github.com/blueworrybear/livelogs/core"
)

var errStreamerNotFound = errors.New("streamer not found")

type logStream struct {
	sync.Mutex

	streamers map[int64]*streamer
}

// New returns a new log stream
func New() core.LogStream {
	return &logStream{
		streamers: make(map[int64]*streamer),
	}
}

func (s *logStream) Create(ctx context.Context, id int64) error {
	s.Lock()
	defer s.Unlock()
	s.streamers[id] = newStreamer()
	return nil
}

func (s *logStream) Write(ctx context.Context, id int64, line *core.LogLine) error {
	s.Lock()
	streamer, ok := s.streamers[id]
	s.Unlock()
	if !ok {
		return errStreamerNotFound
	}
	return streamer.write(line)
}
func (s *logStream) Delete(ctx context.Context, id int64) error {
	s.Lock()
	streamer, ok := s.streamers[id]
	if ok {
		delete(s.streamers, id)
	}
	s.Unlock()
	if !ok {
		return errStreamerNotFound
	}
	return streamer.close()
}

func (s *logStream) Tail(ctx context.Context, id int64) (<-chan *core.LogLine, error) {
	s.Lock()
	streamer, ok := s.streamers[id]
	s.Unlock()
	if !ok {
		return nil, errStreamerNotFound
	}
	return streamer.watch(ctx), nil
}
