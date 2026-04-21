package model

type Area struct {
	ID        uint       `gorm:"primaryKey;column:id"`
	Name      string     `gorm:"column:name;type:varchar(255)"`
	Code      string     `gorm:"column:code;type:varchar(255)"`
	Timestamp *Timestamp `gorm:"embedded"`
}

func (Area) TableName() string {
	return "areas"
}
