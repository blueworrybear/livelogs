package core

import (
	"context"
	"io"
	"time"
)

type LogStore interface {
	Create(ctx context.Context, id int64, r io.Reader) error
	Find(ctx context.Context, id int64) (io.ReadCloser, error)
	Update(ctx context.Context, id int64, r io.Reader) error
	Delete(ctx context.Context, id int64) error
}

type LogLine struct {
	Number    int64     `json:"line"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"time"`
}
