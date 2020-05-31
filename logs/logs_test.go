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
	log := NewLiveLog(1, stream, store)
	log.Write(ctx, &core.LogLine{})
}

func TestLiveLogCat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockLogStore(ctrl)
	stream := mock.NewMockLogStream(ctrl)

	expectLines := make([]*core.LogLine, 0)
	expectLines = append(expectLines, &core.LogLine{})
	expectLines = append(expectLines, &core.LogLine{})
	data, err := json.Marshal(expectLines)
	if err != nil {
		t.Error(err)
		return
	}
	mockR := ioutil.NopCloser(bytes.NewReader(data))
	store.EXPECT().Find(gomock.Eq(int64(1))).Return(mockR, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := NewLiveLog(1, stream, store)
	lines, err := log.Cat(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	if len(lines) != 2 {
		t.Fail()
		return
	}

	emptyLines, err := json.Marshal(make([]*core.LogLine, 0))
	empty := ioutil.NopCloser(bytes.NewReader(emptyLines))
	store.EXPECT().Find(gomock.Eq(int64(2))).Return(empty, nil)
	log = NewLiveLog(2, stream, store)
	lines, err = log.Cat(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	if len(lines) != 0 {
		t.Fail()
		return
	}
}