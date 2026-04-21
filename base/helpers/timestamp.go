package helpers

import (
	"template-go/data/model"
	"time"
)

// SetTimestampModel used as helper to set timestamp model.
func SetTimestampModel(timeNow *time.Time) *model.Timestamp {
	return &model.Timestamp{
		CreatedTs:     *timeNow,
		LastUpdatedTs: *timeNow,
	}
}
