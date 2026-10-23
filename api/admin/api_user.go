package api

import (
	"template-go/api"
)

type FilterUserReq struct {
	api.PaginateRequest
}
type FilterUserResp struct {
	Users []*User `json:"users"`
	api.PaginationMeta
}

type DeleteUserReq struct {
	UserId int `json:"user_id"`
}
