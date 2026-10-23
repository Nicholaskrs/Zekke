package user

import (
	"context"
	"net/http"
	"template-go/api"
	adminApi "template-go/api/admin"
	dto "template-go/base/base"
	"template-go/base/constants"
	userSvc "template-go/modules/user/svc"
	"template-go/util/logtrace"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserAdminHandler struct {
	UserAdminService userSvc.UserAdminService
}

func NewUserAdminHandler(
	UserAdminService userSvc.UserAdminService,
) *UserAdminHandler {
	return &UserAdminHandler{
		UserAdminService: UserAdminService,
	}
}

// Login used to authorize user
func (handler *UserAdminHandler) Login(c *gin.Context) {
	response := &adminApi.AuthResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.AuthReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "login", constants.MethodNull, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "login", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}

	// Call service
	authOut := handler.UserAdminService.LoginUserAdmin(c,
		&userSvc.LoginUserAdminIn{
			Trace:    trace,
			Username: request.Username,
			Password: request.Password,
		})
	if !authOut.BaseOut.Success {
		api.ReturnResponse(c, "login", constants.MethodNull, response, authOut.BaseOut.ErrorMessage, authOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	response.Auth = &adminApi.Auth{
		Token:        authOut.Token,
		UserID:       authOut.UserID,
		UserRole:     authOut.AdminGroupPrivilege,
		FullName:     authOut.FullName,
		Username:     authOut.Username,
		ExpireIn:     authOut.ExpireIn,
		RefreshToken: authOut.RefreshToken,
	}

	// Send response
	api.ReturnResponse(c, "login", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)
}

// FilterUser used to filter user.
func (handler *UserAdminHandler) FilterUser(c *gin.Context) {
	response := &adminApi.FilterUserResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.FilterUserReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "filter user", constants.MethodNull, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "filter user", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}
	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	// Get current user login role and id.
	userID := c.GetInt(constants.ContextAdminID)
	PrivilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service.
	svcOut := handler.UserAdminService.FilterUser(ctx,
		&userSvc.AdminFilterUserIn{
			Trace: trace,
			PaginateIn: dto.PaginateIn{
				Limit: request.GetLimit(),
				Page:  request.Page,
			},
			AdminUserId:         userID,
			AdminPrivilegeGroup: PrivilegeGroup,
		},
	)

	if !svcOut.Success {
		api.ReturnResponse(c, "filter user", constants.MethodGet, response, svcOut.ErrorMessage, svcOut.ErrorCode, trace.TraceId)
		return
	}

	// Set response.
	response.Users = adminApi.ParseUsers(svcOut.Users)
	response.Limit = request.GetLimit()
	response.Page = request.Page
	response.TotalRows = svcOut.TotalRows

	// Send response
	api.ReturnResponse(c, "filter user", constants.MethodGet, response, "", http.StatusOK, trace.TraceId)
}

// DeleteUser used to delete user.
func (handler *UserAdminHandler) DeleteUser(c *gin.Context) {
	trace := logtrace.GetLogTrace(c)
	request := adminApi.DeleteUserReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "user", constants.MethodDelete, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "user", constants.MethodDelete, &request, ve, trace.TraceId)
		return
	}
	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	// Get current user login role and id.
	userID := c.GetInt(constants.ContextAdminID)
	PrivilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service.
	svcOut := handler.UserAdminService.DeleteUser(ctx,
		&userSvc.AdminDeleteUserIn{
			Trace:               trace,
			AdminUserId:         userID,
			AdminPrivilegeGroup: PrivilegeGroup,
			UserId:              request.UserId,
		},
	)

	if !svcOut.Success {
		api.ReturnResponse(c, "user", constants.MethodDelete, nil, svcOut.ErrorMessage, svcOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "user", constants.MethodDelete, nil, "", http.StatusOK, trace.TraceId)
}

// Refresh used to get new access token.
func (handler *UserAdminHandler) Refresh(c *gin.Context) {
	response := &adminApi.RefreshResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.RefreshReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "refresh", constants.MethodNull, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "refresh", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}
	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)

	// Call service.
	svcOut := handler.UserAdminService.Refresh(ctx,
		&userSvc.AdminRefreshIn{
			Trace:        trace,
			RefreshToken: request.RefreshToken,
			AdminUserId:  userID,
		},
	)

	if !svcOut.Success {
		api.ReturnResponse(c, "refresh", constants.MethodNull, nil, svcOut.ErrorMessage, svcOut.ErrorCode, trace.TraceId)
		return
	}

	response = &adminApi.RefreshResp{
		RefreshToken: svcOut.RefreshToken,
		ExpiresIn:    svcOut.ExpireIn,
	}

	// Send response
	api.ReturnResponse(c, "refresh", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)
}

// ChangePassword used to get change new password and revoke all refresh token.
func (handler *UserAdminHandler) ChangePassword(c *gin.Context) {
	response := &adminApi.ChangePasswordResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.ChangePasswordReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "refresh", constants.MethodNull, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "refresh", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}
	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)

	// Call service.
	svcOut := handler.UserAdminService.ChangeAdminPassword(ctx,
		&userSvc.AdminChangePasswordIn{
			Trace:       trace,
			UserAdminId: userID,
			OldPassword: request.OldPassword,
			NewPassword: request.NewPassword,
		},
	)

	if !svcOut.Success {
		api.ReturnResponse(c, "refresh", constants.MethodNull, nil, svcOut.ErrorMessage, svcOut.ErrorCode, trace.TraceId)
		return
	}

	response = &adminApi.ChangePasswordResp{
		RefreshToken: svcOut.RefreshToken,
		ExpiresIn:    svcOut.ExpireIn,
	}

	// Send response
	api.ReturnResponse(c, "refresh", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)
}

// RegisterNewAdmin used to register a new admin.
func (handler *UserAdminHandler) RegisterNewAdmin(c *gin.Context) {
	response := &adminApi.RegisterNewAdminResp{}
	trace := logtrace.GetLogTrace(c)
	request := adminApi.RegisterNewAdminReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "register admin", constants.MethodNull, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "register admin", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}
	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	// Get current user login id.
	userID := c.GetInt(constants.ContextAdminID)
	privilegeGroup := c.GetString(constants.ContextAdminPrivilegeGroup)

	// Call service.
	svcOut := handler.UserAdminService.RegisterNewAdmin(ctx,
		&userSvc.RegisterNewAdminIn{
			Trace:               trace,
			UserAdminId:         userID,
			AdminPrivilegeGroup: privilegeGroup,
			Username:            request.Username,
			Email:               request.Email,
			Password:            request.Password,
			FullName:            request.FullName,
			PrivilegeGroup:      request.PrivilegeGroup,
		},
	)

	if !svcOut.Success {
		api.ReturnResponse(c, "register admin", constants.MethodNull, nil, svcOut.ErrorMessage, svcOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "register admin", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)
}
