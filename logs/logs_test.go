package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/blueworrybear/livelogs/core"
	"github.com/blueworrybear/livelogs/mock"
	"github.com/golang/mock/gomock"
)

type readerMatcher struct {
	x interface{}
}

func (r readerMatcher) Matches(x interface{}) bool {
	reader, ok := x.(io.Reader)
	if !ok {
		return false
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return false
	}
	expect, ok := r.x.([]byte)
	if !ok {
		return false
	}
	if bytes.Compare(expect, data) != 0 {
		return false
	}
	return true
}

func (r readerMatcher) String() string {
	return "reader not match"
}

func TestLiveLogWrite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockLogStore(ctrl)
	stream := mock.NewMockLogStream(ctrl)

	emptyR := ioutil.NopCloser(bytes.NewReader([]byte("[]")))
	lines := make([]*core.LogLine, 0)
	lines = append(lines, &core.LogLine{})
	expectLines, _ := json.Marshal(lines)
	store.EXPECT().Find(gomock.Eq(int64(1))).Return(emptyR, nil)
	store.EXPECT().Update(gomock.Eq(int64(1)), readerMatcher{expectLines}).Return(nil)
	stream.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	ctx := context.Background()
	bufferSize = 0
	log := NewLiveLog(int64(1), stream, store)
	log.Write(ctx, &core.LogLine{})
}
