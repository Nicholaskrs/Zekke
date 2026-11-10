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

func (q *UserQuery) FindByUsername(username string) *model.User {
	var user model.User
	if err := q.db.Where("username = ?", username).First(&user).Error; err != nil {
		q.t.Fatal(err)
		return nil
	}
	return &user
}

func (q *UserQuery) FindFcmTokens(userID uint) []*model.FcmToken {
	var tokens []*model.FcmToken
	if err := q.db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		q.t.Fatal(err)
		return nil
	}
	return tokens
}

func (q *UserQuery) FindRefreshTokens(userID uint) []*model.RefreshToken {
	var tokens []*model.RefreshToken
	if err := q.db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		q.t.Fatal(err)
		return nil
	}
	return tokens
}

func (q *UserQuery) FindRefreshTokenByHash(tokenHash string) *model.RefreshToken {
	var token model.RefreshToken
	if err := q.db.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		q.t.Fatal(err)
		return nil
	}
	return &token
}

func (q *UserQuery) CountUsers() int64 {
	var count int64
	if err := q.db.Model(&model.User{}).Count(&count).Error; err != nil {
		q.t.Fatal(err)
		return 0
	}
	return count
}
