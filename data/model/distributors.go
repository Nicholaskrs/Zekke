package model

type Distributor struct {
	ID        uint       `gorm:"primaryKey;column:id"`
	Name      string     `gorm:"column:name;type:varchar(255)"`
	Timestamp *Timestamp `gorm:"embedded"`
}

func (Distributor) TableName() string {
	return "distributors"
}
