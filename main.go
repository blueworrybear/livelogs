package livelogs

import (
	"github.com/blueworrybear/livelogs/core"
	"github.com/blueworrybear/livelogs/manager"
	"github.com/jinzhu/gorm"
)

// NewLogManager manages open logs
func NewLogManager(db *gorm.DB) core.LogManager {
	return manager.NewLiveLogManager(db)
}
