package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/blueworrybear/livelogs/core"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func testCreate(store core.LogStore, t *testing.T) {
	r := bytes.NewBuffer([]byte("test"))
	id, err := store.Create(r)
	if err != nil {
		t.Error(err)
	}
	if id <= 0 {
		t.Log("No ID return")
		t.Fail()
	}
}

func TestLogStore(t *testing.T) {
	dir, _ := os.Getwd()
	tempFile, err := ioutil.TempFile(dir, "****.db")
	tempFile.Close()
	db, err := gorm.Open("sqlite3", tempFile.Name())
	if err != nil {
		t.Error(err)
	}
	store := NewLogStore(db)
	defer func() {
		db.Close()
		os.Remove(tempFile.Name())
	}()
	t.Run("Test Create", func(t *testing.T) {
		testCreate(store, t)
	})
	t.Run("Test Multiple Create", func(t *testing.T) {
		for i := 1; i <= 10; i++ {
			t.Run(fmt.Sprintf("Create %d", i), func(t *testing.T) {
				t.Parallel()
				testCreate(store, t)
			})
		}
	})
	var count int
	session := db.New()
	session.Model(&logModel{}).Count(&count)
	if count != 11 {
		t.Logf("Count: %d", count)
		t.Fail()
	}
	t.Run("Test Find", func(t *testing.T) {
		r := bytes.NewBuffer([]byte("find me!"))
		id, err := store.Create(r)
		if err != nil {
			t.Error(err)
			return
		}
		rc, err := store.Find(id)
		if err != nil {
			t.Error(err)
			return
		}
		defer rc.Close()
		data, err := ioutil.ReadAll(rc)
		if err != nil {
			t.Error(err)
			return
		}
		if string(data) != "find me!" {
			t.Fail()
		}
	})
}
