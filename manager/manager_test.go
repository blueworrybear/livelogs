package manager

import (
	"os"
	"io/ioutil"
	"testing"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestManager (t *testing.T) {
	dir, _ := os.Getwd()
	tempFile, err := ioutil.TempFile(dir, `*.db`)
	tempFile.Close()
	db, err := gorm.Open("sqlite3", tempFile.Name())
	if err != nil {
		t.Error(err)
	}
	defer func() {
		db.Close()
		os.Remove(tempFile.Name())
	}()
	m := NewLiveLogManager(db)
	log, err := m.Create()
	if err != nil {
		t.Error(err)
		return
	}
	if log.ID() <= 0 {
		t.Fail()
		return
	}
	id := log.ID()

	log, err = m.Open(id)
	if err != nil {
		t.Error(err)
		return
	}
	if log.ID() != id {
		t.Fail()
		return
	}

	if err := m.Close(id); err != nil {
		t.Error(err)
		return
	}
	log, err = m.Open(id)
	if err != nil {
		t.Error(err)
		return
	}
	if log.ID() != id {
		t.Fail()
		return
	}
}