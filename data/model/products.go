package model

import "time"

type Product struct {
	ID            uint       `gorm:"primaryKey;column:id"`
	Name          string     `gorm:"column:name;type:varchar(255)"`
	Timestamp     *Timestamp `gorm:"embedded"`
	CreatedByID   uint       `gorm:"column:created_by_id;type:int"`
	LastUpdatedBy uint       `gorm:"column:last_updated_by;type:int"`
	DeletedBy     *uint      `gorm:"column:deleted_by;type:int"`
	DeletedTs     *time.Time `gorm:"column:deleted_ts;type:datetime"`
}

func (Product) TableName() string {
	return "products"
}
