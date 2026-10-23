package model

import "time"

type Timestamp struct {
	CreatedTs     time.Time `gorm:"column:created_ts;type:timestamptz"`
	LastUpdatedTs time.Time `gorm:"column:last_updated_ts;type:timestamptz"`
}
