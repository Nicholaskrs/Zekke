package fixture

import (
	"template-go/data/model"
	"testing"

	"gorm.io/gorm"
)

type UserQuery struct {
	t  *testing.T
	db *gorm.DB
}

func NewUserQuery(t *testing.T, db *gorm.DB) *UserQuery {
	return &UserQuery{
		t:  t,
		db: db,
	}
}

func (q *UserQuery) FindAll() []model.User {
	var users []model.User
	if err := q.db.Find(&users).Error; err != nil {
		q.t.Fatal(err)
		return nil
	}
	return users
}

func (q *UserQuery) FindByID(id uint) *model.User {
	var user model.User
	if err := q.db.First(&user, id).Error; err != nil {
		q.t.Fatal(err)
		return nil
	}
	return &user
}

func (q *UserQuery) CountUsers() int64 {
	var count int64
	if err := q.db.Model(&model.User{}).Count(&count).Error; err != nil {
		q.t.Fatal(err)
		return 0
	}
	return count
}
