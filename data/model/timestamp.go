package model

import "time"

// Currently all timestamp is using time.now() (depends on machine locale).

type Timestamp struct {
	CreatedTs     time.Time `gorm:"column:created_ts;type:timestamp"`
	LastUpdatedTs time.Time `gorm:"column:last_updated_ts;type:timestamp"`
}
