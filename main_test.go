package livelogs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/blueworrybear/livelogs/core"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

func TestLogManager(t *testing.T) {
	m := NewLogManager(db)
	for i := 0; i < 50; i++ {
		t.Run(fmt.Sprintf("Run time %d", i), func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			log, err := m.Create()
			if err != nil {
				t.Error(err)
				return
			}
			tail, err := log.Tail(ctx)
			if err != nil {
				t.Error(err)
				return
			}
			go func ()  {
				for i := 0; i < 10; i++ {
					log.Write(ctx, &core.LogLine{Number: int64(i)})
				}
				if err := m.Close(log.ID()); err != nil {
					cancel()
				}
			}()
			for {
				select{
				case <-ctx.Done():
					t.Log("Race condition...")
					t.Fail()
					return
				case _, ok := <-tail:
					if !ok {
						return
					}
				case <- time.After(1 * time.Second):
					t.Logf("Deadlock at %d run", log.ID())
					t.Fail()
					return
				}
			}
		})
	}
}

func TestMain(m *testing.M) {
	dir, _ := os.Getwd()
	tempFile, err := ioutil.TempFile(dir, `*.db`)
	tempFile.Close()
	db, err = gorm.Open("sqlite3", tempFile.Name())
	if err != nil {
		os.Exit(1)
	}
	ok := m.Run()
	db.Close()
	os.Remove(tempFile.Name())
	os.Exit(ok)
}