package model

type Attendance struct {
	ID              uint       `gorm:"primaryKey;column:id"`
	UserID          uint       `gorm:"column:user_id;type:int"`
	VisitationCount int        `gorm:"column:visitation_count;type:int"`
	Timestamp       *Timestamp `gorm:"embedded"`

	// Date is used to store attendance date, we put it on date and time to make it easier to filter.
	Date string `gorm:"column:date;type:date"`
}

func (Attendance) TableName() string {
	return "attendances"
}
