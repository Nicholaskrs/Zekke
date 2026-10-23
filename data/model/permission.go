package model

type PrivilegeGroup struct {
	ID          uint   `gorm:"primaryKey;column:id"`
	Code        string `gorm:"column:code;type:varchar(255)"`
	Description string `gorm:"column:description;type:varchar(255)"`
}

func (PrivilegeGroup) TableName() string {
	return "privilege_groups"
}

type UserPrivilegeGroup struct {
	UserID             uint   `gorm:"column:user_id;type:int;uniqueIndex"`
	PrivilegeGroupCode string `gorm:"column:privilege_group_code;type:varchar(255)"`
}

func (UserPrivilegeGroup) TableName() string {
	return "user_privilege_groups"
}

type PrivilegeGroupPrivilege struct {
	ID                 uint   `gorm:"primaryKey;column:id"`
	PrivilegeGroupCode string `gorm:"column:privilege_group_code;type:varchar(255)"`
	Privilege          string `gorm:"column:privilege;type:varchar(255)"`
}

func (PrivilegeGroupPrivilege) PrivilegeGroupPrivilege() string {
	return "privilege_group_privileges"
}

type Privilege struct {
	ID        uint   `gorm:"primaryKey;column:id"`
	Privilege string `gorm:"column:privilege;type:varchar(255)"`
}

func (Privilege) TableName() string {
	return "privileges"
}
