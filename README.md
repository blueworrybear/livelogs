[![Maintainability](https://api.codeclimate.com/v1/badges/542dcba773db3daf5fe6/maintainability)](https://codeclimate.com/github/blueworrybear/livelogs/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/542dcba773db3daf5fe6/test_coverage)](https://codeclimate.com/github/blueworrybear/livelogs/test_coverage)
[![Go Test](https://github.com/blueworrybear/livelogs/workflows/Go%20Test/badge.svg)](https://github.com/blueworrybear/livelogs/actions)
# Live Logs to Go

Livelogs is a library to stream logs to multiple watchers.

The logs could also be save into database with [gorm](https://github.com/jinzhu/gorm).

![livelogs](https://i.imgur.com/k4EO02H.gif)

# Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/blueworrybear/livelogs"
	"github.com/blueworrybear/livelogs/core"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	db, _ := gorm.Open("sqlite3", "core.db")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := livelogs.NewLogManager(db)
	log, _ := m.Create()

	go func() {
		tail, _ := log.Tail(ctx)
		for {
			select {
			case line := <-tail:
				fmt.Println(line.Text)
			}
		}
	}()

	for {
		log.Write(ctx, &core.LogLine{Text: "log"})
		time.Sleep(1 * time.Second)
	}
}
```

Above code will keep print out `log`.

# Documentation

[Godoc](https://pkg.go.dev/github.com/blueworrybear/livelogs)

# License

© Benno, 2020

Released under the [MIT](https://github.com/blueworrybear/livelogs/blob/master/LICENSE) License