package model

import (
	"template-go/data/enum"
)

type User struct {
	ID uint `gorm:"primaryKey;column:id"`
	// ExternalID filled with uuid. This is used to view user and change password.
	ExternalID    string     `gorm:"column:external_id;type:char(36)"`
	Username      string     `gorm:"column:username;type:varchar(255)"`
	Email         string     `gorm:"column:email;type:varchar(255)"`
	Password      string     `gorm:"column:password;type:varchar(255)"`
	FullName      string     `gorm:"column:full_name;type:varchar(255)"`
	Role          enum.Role  `gorm:"column:role;type:ENUM('National', 'Area Manager', 'Distributor', 'Operator', 'Sales')"`
	DistributorID uint       `gorm:"column:distributor_id;type:int"`
	AreaID        uint       `gorm:"column:area_id;type:int"`
	Timestamp     *Timestamp `gorm:"embedded"`
}

func (User) TableName() string {
	return "users"
}
