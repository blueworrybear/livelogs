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

type LiveLog struct {
	sync.Mutex
	id     int64
	buffer []*core.LogLine
	stream core.LogStream
	store  core.LogStore
}

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

func (l *LiveLog) ID() int64 {
	return l.id
}

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

func (l *LiveLog) Save(ctx context.Context) error {
	return l.flush()
}
