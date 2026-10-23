package svc

import (
	"context"
	dto "template-go/base/base"
	"template-go/data/model"
	"template-go/util/logtrace"

	"github.com/golang-jwt/jwt"
)

type UserAdminService interface {
	// LoginUserAdmin used to authorize and authenticate admin. It checks into admin's username and password.
	// If user do exists, it returns generated token.
	LoginUserAdmin(ctx context.Context, request *LoginUserAdminIn) *LoginUserAdminOut

	// FilterUser used to get all active users.
	FilterUser(ctx context.Context, paramIn *AdminFilterUserIn) *AdminFilterUserOut

	// DeleteUser used to delete active users.
	DeleteUser(ctx context.Context, paramIn *AdminDeleteUserIn) *AdminDeleteUserOut

	// Refresh exchanges a valid refresh token for a brand-new token pair. The presented
	// refresh token is consumed and can never be used again (rotation).
	Refresh(ctx context.Context, paramIn *AdminRefreshIn) *AdminRefreshOut

	// ChangeAdminPassword used to update admin's password.
	// this will revoke all previous refresh token and will return new refresh token.
	ChangeAdminPassword(ctx context.Context, paramIn *AdminChangePasswordIn) *AdminChangePasswordOut

	// RegisterNewAdmin used to register a new administrator in the system. It takes the registration parameters
	// and creates a new admin account, returning the registration result.
	RegisterNewAdmin(ctx context.Context, paramIn *RegisterNewAdminIn) *RegisterNewAdminOut
}

type LoginUserAdminIn struct {
	Trace    *logtrace.LogTrace
	Username string
	Password string
}

type LoginUserAdminOut struct {
	dto.BaseOut
	Token               string
	UserID              uint
	AdminGroupPrivilege string
	FullName            string
	Username            string
	ExpireIn            int
	RefreshToken        string
}

type AdminAuthCustomClaims struct {
	AdminID        uint
	Email          string
	FullName       string
	PrivilegeGroup string
	jwt.StandardClaims
}

type AdminFilterUserIn struct {
	Trace *logtrace.LogTrace
	dto.PaginateIn
	AdminUserId         int
	AdminPrivilegeGroup string
}

type AdminFilterUserOut struct {
	dto.BaseOut
	Users     []*model.User
	TotalRows int
}

type AdminDeleteUserIn struct {
	Trace               *logtrace.LogTrace
	AdminUserId         int
	AdminPrivilegeGroup string
	UserId              int
}

type AdminDeleteUserOut struct {
	dto.BaseOut
}

type AdminRefreshIn struct {
	Trace        *logtrace.LogTrace
	AdminUserId  int
	RefreshToken string
}

type AdminRefreshOut struct {
	dto.BaseOut
	AccessToken  string
	RefreshToken string
	ExpireIn     int
}

type AdminChangePasswordIn struct {
	Trace       *logtrace.LogTrace
	UserAdminId int
	OldPassword string
	NewPassword string
}

type AdminChangePasswordOut struct {
	dto.BaseOut
	RefreshToken string
	ExpireIn     int
}

type RegisterNewAdminIn struct {
	Trace               *logtrace.LogTrace
	UserAdminId         int
	AdminPrivilegeGroup string
	Username            string
	Email               string
	Password            string
	FullName            string
	PrivilegeGroup      string
}
type RegisterNewAdminOut struct {
	dto.BaseOut
}
