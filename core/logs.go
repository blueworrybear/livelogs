package core

import (
	"context"
	"io"
	"time"
)

//go:generate mockgen -package core -destination logs_mock.go . LogStore

// LogLine holds line information in log
type LogLine struct {
	Number    int64     `json:"line"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"time"`
}

// LogStore is a storage to save log output
type LogStore interface {
	Create(r io.Reader) (int64, error)
	Find(id int64) (io.ReadCloser, error)
	Update(id int64, r io.Reader) error
	Delete(id int64) error
}

// LogStream provides log streaming
type LogStream interface {
	Create(ctx context.Context, id int64) error
	Write(ctx context.Context, id int64, line *LogLine) error
	Delete(ctx context.Context, id int64) error
	Tail(ctx context.Context, id int64) (<-chan *LogLine, error)
}

// Log is a log file
type Log interface {
	ID() int64
	Write(line *LogLine) error
	Save() error
	Remove() error
	Tail() <-chan *LogLine
}

// LogManager manages log files
type LogManager interface {
	Create() Log
	Find(id int64) Log
}
