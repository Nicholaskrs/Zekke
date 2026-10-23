package api

import (
	"template-go/api"
)

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,ascii"`
	Password string `json:"password" binding:"required,ascii"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required,ascii"`
	UserRole string `json:"user_role" binding:"required,ascii"`
}

type UserRegisterResp struct {
	*api.Response
}

type GetUserProfileResp struct {
	*api.Response
	User *User `json:"user"`
}

type InsertFcmTokenReq struct {
	Token string `json:"token" binding:"required"`
}
type InsertFcmTokenResp struct {
	*api.Response
}

type DeleteFcmTokenReq struct {
	Token string `json:"token" binding:"required"`
}
type DeleteFcmTokenResp struct {
	*api.Response
}
