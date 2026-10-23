package model

import "time"

type User struct {
	ID uint `gorm:"primaryKey;column:id"`
	// ExternalID filled with uuid. This is used to view user and change password.
	ExternalID string     `gorm:"column:external_id;type:char(36)"`
	Username   string     `gorm:"column:username;type:varchar(255);uniqueIndex"`
	Email      string     `gorm:"column:email;type:varchar(255);uniqueIndex"`
	Password   string     `gorm:"column:password;type:varchar(255)"`
	FullName   string     `gorm:"column:full_name;type:varchar(255)"`
	Role       string     `gorm:"column:role;type:varchar(255)"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;type:timestamptz"`
	Timestamp  *Timestamp `gorm:"embedded"`
}

func (User) TableName() string {
	return "users"
}

type Role struct {
	Code        string     `gorm:"column:code;type:varchar(255);uniqueIndex"`
	Description string     `gorm:"column:description;type:varchar(255)"`
	Timestamp   *Timestamp `gorm:"embedded"`
}

func (Role) TableName() string {
	return "roles"
}

type RefreshToken struct {
	ID        uint       `gorm:"primaryKey;column:id"`
	UserID    uint       `gorm:"column:user_id;type:int"`
	FamilyID  string     `gorm:"column:family_id;type:varchar(36)"`
	TokenHash string     `gorm:"column:token_hash;type:text"`
	ExpireAt  *time.Time `gorm:"column:expire_at;type:timestamptz"`
	UsedAt    *time.Time `gorm:"column:used_at;type:timestamptz"`
	RevokedAt *time.Time `gorm:"column:revoked_at;type:timestamptz"`
	Timestamp *Timestamp `gorm:"embedded"`
}

func (RefreshToken) TableName() string {
	return "refresh_token"
}
