package logs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/blueworrybear/livelogs/core"
)

type mockStreamer struct {
	*streamer
}

func TestStream(t *testing.T) {
	var w sync.WaitGroup
	s := newStreamer()
	s.write(&core.LogLine{})
	s.write(&core.LogLine{})
	s.write(&core.LogLine{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := s.watch(ctx)
	w.Add(6)
	go func() {
		s.write(&core.LogLine{})
		s.write(&core.LogLine{})
		w.Done()
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-stream:
				w.Done()
			}
		}
	}()
	closed := make(chan struct{})
	go func() {
		w.Wait()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(1000 * time.Millisecond):
		t.Log("Timeout, unable to restore logs")
		t.Fail()
	}
}
