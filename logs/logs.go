package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/blueworrybear/livelogs/core"
)

var bufferSize = 5000

// LiveLog is the based log object
type LiveLog struct {
	sync.Mutex
	id     int64
	buffer []*core.LogLine
	stream core.LogStream
	store  core.LogStore
}

// NewLiveLog returns a live log implements core.Log
func NewLiveLog(id int64, stream core.LogStream, store core.LogStore) *LiveLog {
	return &LiveLog{
		id:     id,
		buffer: make([]*core.LogLine, 0),
		stream: stream,
		store:  store,
	}
}

func (l *LiveLog) flush() error {
	r, err := l.store.Find(l.id)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var lines []*core.LogLine
	json.Unmarshal(data, &lines)
	lines = append(lines, l.buffer...)
	data, err = json.Marshal(lines)
	if err != nil {
		return err
	}
	br := bytes.NewReader(data)
	return l.store.Update(l.id, br)
}

// ID is log's identity
func (l *LiveLog) ID() int64 {
	return l.id
}

// Write log line
func (l *LiveLog) Write(ctx context.Context, line *core.LogLine) error {
	l.Lock()
	l.buffer = append(l.buffer, line)
	if size := len(l.buffer); size > bufferSize {
		if err := l.flush(); err != nil {
			return err
		}
		l.buffer = l.buffer[:0]
	}
	l.Unlock()
	return l.stream.Write(ctx, l.id, line)
}

// Remove log from the database and disable streaming
func (l *LiveLog) Remove(ctx context.Context) error {
	l.Lock()
	defer l.Unlock()
	if err := l.stream.Delete(ctx, l.id); err != nil {
		return err
	}
	return l.store.Delete(l.id)
}

// Close the log and disable streaming
func (l *LiveLog) Close(ctx context.Context) error {
	l.Lock()
	defer l.Unlock()
	return l.stream.Delete(ctx, l.id)
}

// Cat the log lines.
//
// **NOTE** The content may not be complete before Save()
func (l *LiveLog) Cat(ctx context.Context) ([]*core.LogLine, error) {
	rc, err := l.store.Find(l.id)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	var lines []*core.LogLine
	err = json.Unmarshal(data, &lines)
	return lines, err
}

// Tail follows the log
func (l *LiveLog) Tail(ctx context.Context) (<-chan *core.LogLine, error) {
	l.Lock()
	defer l.Unlock()
	return l.stream.Tail(ctx, l.id)
}

// Save log lines into database
func (l *LiveLog) Save(ctx context.Context) error {
	l.Lock()
	defer l.Unlock()
	return l.flush()
}
