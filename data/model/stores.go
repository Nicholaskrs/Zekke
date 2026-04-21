package model

type Store struct {
	ID          uint       `gorm:"primaryKey;column:id"`
	Code        string     `gorm:"column:code;unique;not null"`
	Name        string     `gorm:"column:name;type:varchar(255)"`
	PhoneNumber string     `gorm:"column:phone_number;type:varchar(255)"`
	FreezerIdn  string     `gorm:"column:freezer_idn;unique;type:varchar(255);"` // freezer identity number (used as freezer serial number)
	CreatedByID uint       `gorm:"column:created_by_id;type:int"`
	Timestamp   *Timestamp `gorm:"embedded"`
}

func (Store) TableName() string {
	return "stores"
}
