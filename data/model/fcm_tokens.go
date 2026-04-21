package model

type FcmToken struct {
	ID        uint       `gorm:"primaryKey;column:id"`
	UserID    uint       `gorm:"column:user_id;type:int"`
	FcmToken  string     `gorm:"column:fcm_token;type:varchar(255)"`
	Timestamp *Timestamp `gorm:"embedded"`
}

func (FcmToken) TableName() string {
	return "fcm_tokens"
}
