package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"

	"github.com/blueworrybear/livelogs/core"
	"github.com/blueworrybear/livelogs/logs"
	"github.com/blueworrybear/livelogs/store"
	"github.com/blueworrybear/livelogs/stream"
	"github.com/jinzhu/gorm"
)

var errLogNotFound = errors.New("Log not found")

type LiveLogManager struct {
	sync.Mutex
	store  core.LogStore
	stream core.LogStream
	logs   map[int64]core.Log
}

func NewLiveLogManager(db *gorm.DB) *LiveLogManager {
	store := store.NewLogStore(db)
	stream := stream.New()
	return &LiveLogManager{
		store:  store,
		stream: stream,
		logs:   make(map[int64]core.Log),
	}
}

func (m *LiveLogManager) Create() (core.Log, error) {
	m.Lock()
	defer m.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lines, err := json.Marshal(make([]*core.LogLine, 0))
	if err != nil {
		return nil, err
	}
	r := ioutil.NopCloser(bytes.NewReader(lines))
	id, err := m.store.Create(r)
	if err != nil {
		return nil, err
	}
	if err := m.stream.Create(ctx, id); err != nil {
		return nil, err
	}
	log := logs.NewLiveLog(id, m.stream, m.store)
	m.logs[id] = log
	return log, nil
}

func (m *LiveLogManager) Open(id int64) (core.Log, error) {
	m.Lock()
	defer m.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if log, ok := m.logs[id]; ok {
		return log, nil
	}
	if !m.store.Exists(id) {
		return nil, errLogNotFound
	}
	if err := m.stream.Create(ctx, id); err != nil {
		return nil, err
	}
	m.logs[id] = logs.NewLiveLog(id, m.stream, m.store)
	return m.logs[id], nil
}

func (m *LiveLogManager) Close(id int64) error {
	m.Lock()
	defer m.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log, ok := m.logs[id]
	if !ok {
		return errLogNotFound
	}
	if err := log.Close(ctx); err != nil {
		return err
	}
	delete(m.logs, id)
	return nil
}
