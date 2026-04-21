package model

import "template-go/data/enum"

type VisitationImage struct {
	ID           uint           `gorm:"primaryKey;column:id"`
	VisitationID uint           `gorm:"visitation_id;type:int"`
	Type         enum.ImageType `gorm:"column:type;type:ENUM('Store', 'Freezer', 'FreezerCode', 'ProductDisplayClose', 'ProductDisplayOpen', 'LowerFreezerPhoto', 'BackupPhoto', 'FreezerThermometer')"`
	ImageUrl     string         `gorm:"column:image_url;type:varchar(255)"`
	Timestamp    *Timestamp     `gorm:"embedded"`
}

func (VisitationImage) TableName() string {
	return "visitation_images"
}
