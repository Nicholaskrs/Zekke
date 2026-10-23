package api

import (
	"template-go/api/common"
	"template-go/data/model"
)

type User struct {
	ExternalId string `json:"external_id"`
	FullName   string `json:"full_name"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func ParseUser(user *model.User) *User {
	return &User{
		ExternalId: user.ExternalID,
		FullName:   user.FullName,
		Username:   user.Username,
		Email:      user.Email,
		CreatedAt:  common.ToDateString(user.Timestamp.CreatedTs),
		UpdatedAt:  common.ToDateString(user.Timestamp.LastUpdatedTs),
	}
}

func ParseUsers(users []*model.User) []*User {
	result := make([]*User, len(users))
	for i, user := range users {
		result[i] = ParseUser(user)
	}
	return result

}
