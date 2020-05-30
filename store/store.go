package store

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/blueworrybear/livelogs/core"
	"github.com/jinzhu/gorm"
)

type logModel struct {
	ID   int64
	Data []byte
}

type logStore struct {
	db *gorm.DB
}

func NewLogStore(db *gorm.DB) core.LogStore {
	db.AutoMigrate(&logModel{})
	return &logStore{
		db: db,
	}
}

func (s *logStore) Create(r io.Reader) (int64, error) {
	session := s.db.New()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	m := &logModel{
		Data: data,
	}
	if err := session.Create(m).Error; err != nil {
		return m.ID, err
	}
	return m.ID, nil
}

func (s *logStore) Find(id int64) (io.ReadCloser, error) {
	session := s.db.New()
	m := &logModel{}
	if err := session.First(m, id).Error; err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(m.Data)), nil
}

func (s *logStore) Update(id int64, r io.Reader) error {
	session := s.db.New()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	m := &logModel{}
	if err := session.First(m, id).Error; err != nil {
		return err
	}
	m.Data = data
	if err := session.Save(m).Error; err != nil {
		return err
	}
	return nil
}

func (s *logStore) Delete(id int64) error {
	session := s.db.New()
	m := &logModel{}
	if err := session.First(m, id).Error; err != nil {
		return err
	}
	if err := session.Delete(m).Error; err != nil {
		return err
	}
	return nil
}
