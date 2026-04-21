package svc

import (
	"context"
	dto "template-go/base/base"
	"template-go/data/model"
	"template-go/util/trace"

	"github.com/golang-jwt/jwt"
)

type UserService interface {
	// LoginUser used to authorize and authenticate user. It checks into user's username and password.
	// If user do exists, it returns generated token.
	LoginUser(ctx context.Context, request *LoginUserIn) *LoginUserOut

	// Register used to register new user. Note that register is for internal usage only.
	Register(ctx context.Context, paramIn *UserRegisterIn) *UserRegisterOut

	// ChangePasswordByExternalID used to change sales password.
	// Note that only the area manager who oversees the sales has permission to change their password.
	ChangePasswordByExternalID(ctx context.Context, paramIn *ChangePasswordIn) *ChangePasswordOut

	// GetUser will return user based on ID it'll return success false if user not exists.
	GetUser(ctx context.Context, request *GetUserIn) *GetUserOut

	InsertFcmToken(ctx context.Context, paramIn *InsertFcmTokenIn) *InsertFcmTokenOut

	DeleteFcmTokenBulk(ctx context.Context, paramIn *DeleteFcmTokenBulkIn) *DeleteFcmTokenBulkOut

	GetUserFcmToken(ctx context.Context, paramIn *GetUserFcmTokenIn) *GetUserFcmTokenOut
}

type LoginUserIn struct {
	Trace    *trace.Trace
	Username string
	Password string
}

type LoginUserOut struct {
	dto.BaseOut
	Token    string
	UserID   uint
	UserRole string
	FullName string
	Username string
}

type AuthCustomClaims struct {
	ID       uint
	Email    string
	FullName string
	Role     string
	jwt.StandardClaims
}

type UserRegisterIn struct {
	Trace         *trace.Trace
	Username      string
	Email         string
	Password      string
	FullName      string
	UserRole      string
	DistributorID uint
	AreaID        uint
}

type UserRegisterOut struct {
	dto.BaseOut
}

type ChangePasswordIn struct {
	Trace      *trace.Trace
	ExternalID string
	Password   string
	UserID     uint
}

type ChangePasswordOut struct {
	dto.BaseOut
}

type GetUserIn struct {
	Trace  *trace.Trace
	UserID uint
}

type GetListUsersByRoleIn struct {
	Trace  *trace.Trace
	UserID uint
}
type GetListuserByRoleOut struct {
	dto.BaseOut
	UserIDs []uint
}

type GetUserOut struct {
	dto.BaseOut
	User *model.User
}

type GetDistributorIDByUserIDIn struct {
	Trace  *trace.Trace
	UserID uint
}

type GetDistributorIDByUserIDOut struct {
	dto.BaseOut
	DistributorID uint
}

type GetSalesUserIn struct {
	Trace  *trace.Trace
	UserID uint
}

type GetSalesUserOut struct {
	dto.BaseOut
	Users []*model.User
}

type GetSalesUsersByAreaIDIn struct {
	Trace  *trace.Trace
	AreaID uint
}

type GetSalesUsersByAreaIDOut struct {
	dto.BaseOut
	Users []*model.User
}

type GetUsersByDistributorIDIn struct {
	Trace         *trace.Trace
	DistributorID uint
}

type GetUsersByDistributorIDOut struct {
	dto.BaseOut
	Users []*model.User
}

type InsertFcmTokenIn struct {
	Trace  *trace.Trace
	UserID uint
	Token  string
}

type InsertFcmTokenOut struct {
	dto.BaseOut
}

type DeleteFcmTokenBulkIn struct {
	Trace  *trace.Trace
	Tokens []string
}

type DeleteFcmTokenBulkOut struct {
	dto.BaseOut
}

type GetUserFcmTokenIn struct {
	Trace  *trace.Trace
	UserID uint
}

type GetUserFcmTokenOut struct {
	dto.BaseOut
	FcmTokens []*model.FcmToken
}
