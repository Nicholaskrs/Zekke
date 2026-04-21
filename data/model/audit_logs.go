package model

import (
	"time"
)

type AuditLog struct {
	ID            uint      `gorm:"primaryKey;column:id"`
	EntityType    string    `gorm:"column:entity_type;type:varchar(255)"`
	EntityKey     string    `gorm:"column:entity_key;type:varchar(255)"`
	TraceID       string    `gorm:"column:trace_id;type:varchar(255)"`
	Diff          string    `gorm:"column:diff;type:text"`
	UserID        uint      `gorm:"column:user_id;type:int"`
	ServiceCaller string    `gorm:"column:service_caller;type:varchar(255)"`
	CreatedTs     time.Time `gorm:"column:created_ts;type:datetime"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
