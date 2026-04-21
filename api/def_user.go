package api

import (
	"template-go/data/enum"
	"template-go/data/model"
)

type User struct {
	ID            uint      `json:"id"`
	ExternalId    string    `json:"external_id"`
	Role          enum.Role `json:"role"`
	FullName      string    `json:"full_name"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	DistributorID uint      `json:"distributor_id"`
	AreaID        uint      `json:"area_id"`
}

func ParseUser(user *model.User) *User {
	return &User{
		ID:            user.ID,
		ExternalId:    user.ExternalID,
		Role:          user.Role,
		FullName:      user.FullName,
		Username:      user.Username,
		Email:         user.Email,
		DistributorID: user.DistributorID,
		AreaID:        user.AreaID,
	}
}
