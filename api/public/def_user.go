package api

import (
	"template-go/data/model"
)

type User struct {
	ExternalId string `json:"external_id"`
	Role       string `json:"role"`
	FullName   string `json:"full_name"`
	Username   string `json:"username"`
	Email      string `json:"email"`
}

func ParseUser(user *model.User) *User {
	return &User{
		ExternalId: user.ExternalID,
		Role:       user.Role,
		FullName:   user.FullName,
		Username:   user.Username,
		Email:      user.Email,
	}
}
