package stream

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/blueworrybear/livelogs/core"
)

func TestLosStream(t *testing.T) {
	i := New()
	s, ok := i.(*logStream)
	if !ok {
		t.Log("Fail to new logStream")
		t.Fail()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Create(ctx, 1); err != nil {
		t.Error(err)
		return
	}
	var w sync.WaitGroup
	w.Add(6)
	s.Write(ctx, 1, &core.LogLine{})
	s.Write(ctx, 1, &core.LogLine{})
	s.Write(ctx, 1, &core.LogLine{})
	go func ()  {
		s.Write(ctx, 1, &core.LogLine{})
		s.Write(ctx, 1, &core.LogLine{})
		w.Done()
	}()
	tail, err := s.Tail(ctx, 1)
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		for {
			select{
			case <-ctx.Done():
				return
			case <-tail:
				w.Done()
			}
		}
	}()
	finish := make(chan struct{})
	go func ()  {
		w.Wait()
		close(finish)
	}()
	select {
	case <-finish:
	case <-time.After(3 * time.Second):
		t.Log("Time out!")
		t.Fail()
		return
	}
}

func TestLogStreamError(t *testing.T) {
	s := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	err = s.Write(ctx, 1234, &core.LogLine{})
	if err == nil || err != errStreamerNotFound {
		t.Fail()
		return
	}
	err = s.Delete(ctx, 1234)
	if err == nil || err != errStreamerNotFound {
		t.Fail()
		return
	}
	_, err = s.Tail(ctx, 1234)
	if err == nil || err != errStreamerNotFound {
		t.Fail()
		return
	}
}

func TestLosStreamDelete(t *testing.T) {
	s := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Create(ctx, 1); err != nil {
		t.Error(err)
		return
	}
	var w sync.WaitGroup
	tail, err := s.Tail(ctx, 1)
	if err != nil {
		t.Error(err)
		return
	}
	w.Add(1)
	go func ()  {
		defer w.Done()
		for {
			select{
			case _, ok := <-tail:
				if !ok {
					return
				}
			case <- ctx.Done():
				return
			}
		}
	}()
	if err := s.Delete(ctx, 1);err != nil {
		t.Error(err)
		return
	}
	finish := make(chan struct{})
	go func ()  {
		w.Wait()
		close(finish)
	}()
	select {
	case <-finish:
	case <-time.After(1 * time.Second):
		t.Log("Time out")
		t.Fail()
	}
}