package api

import (
	"template-go/api"
)

type AuthReq struct {
	Username string `json:"username" binding:"required,ascii"`
	Password string `json:"password" binding:"required,ascii"`
}

type AuthResp struct {
	*api.Response
	*Auth
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResp struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
