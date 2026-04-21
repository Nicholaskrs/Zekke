package user

import (
	"context"
	"net/http"
	"template-go/api"
	"template-go/base/constants"
	"template-go/data/enum"
	userSvc "template-go/modules/user/svc"
	"template-go/util/trace"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UserHandler struct {
	UserService userSvc.UserService
}

func NewUserHandler(
	UserService userSvc.UserService,
) *UserHandler {
	return &UserHandler{
		UserService: UserService,
	}
}

// Login used to authorize user
func (handler *UserHandler) Login(c *gin.Context) {
	response := &api.ApiAuthResp{}
	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}
	request := api.ApiAuthReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "login", constants.MethodNull, request, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "login", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}

	// Call service
	authOut := handler.UserService.LoginUser(c,
		&userSvc.LoginUserIn{
			Trace:    trace,
			Username: request.Username,
			Password: request.Password,
		})
	if !authOut.BaseOut.Success {
		api.ReturnResponse(c, "login", constants.MethodNull, response, authOut.BaseOut.ErrorMessage, authOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	auth := &api.Auth{
		Token:    authOut.Token,
		UserID:   authOut.UserID,
		UserRole: authOut.UserRole,
		FullName: authOut.FullName,
		Username: authOut.Username,
	}
	response.Auth = auth

	// Send response
	api.ReturnResponse(c, "login", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)
}

// Register used to register new user
func (handler *UserHandler) Register(c *gin.Context) {
	response := &api.UserRegisterResp{}

	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}

	request := api.UserRegisterReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "register", constants.MethodNull, request, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "register", constants.MethodNull, &request, ve, trace.TraceId)
		return
	}

	sanitizedRequest := request
	sanitizedRequest.Password = ""
	trace.Request = sanitizedRequest

	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	// Call service
	authOut := handler.UserService.Register(ctx,
		&userSvc.UserRegisterIn{
			Trace:         trace,
			Username:      request.Username,
			Email:         request.Email,
			Password:      request.Password,
			FullName:      request.FullName,
			UserRole:      request.UserRole,
			DistributorID: request.DistributorID,
			AreaID:        request.AreaID,
		},
	)
	if !authOut.BaseOut.Success {
		api.ReturnResponse(c, "register", constants.MethodNull, response, authOut.BaseOut.ErrorMessage, authOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "register", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)
}

// ChangePasswordByExternalID used to change sales password. Note that only the area manager who oversees the sales has permission to change their password.
func (handler *UserHandler) ChangePasswordByExternalID(c *gin.Context) {
	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	request := api.ChangePasswordReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "change password", constants.MethodUpdate, request, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "change password", constants.MethodUpdate, &request, ve, trace.TraceId)
		return
	}

	// Get current user login role and id
	userID := c.GetInt("ID")
	userRole := c.GetString("role")

	if enum.Role(userRole) != enum.AreaManager {
		api.ReturnResponse(
			c,
			"change password",
			constants.MethodUpdate,
			&request,
			"forbidden access",
			http.StatusForbidden,
			trace.TraceId,
		)
		return
	}

	// Call service
	serviceOut := handler.UserService.ChangePasswordByExternalID(ctx,
		&userSvc.ChangePasswordIn{
			Trace:      trace,
			ExternalID: request.ExternalID,
			Password:   request.Password,
			UserID:     uint(userID),
		},
	)

	if !serviceOut.BaseOut.Success {
		api.ReturnResponse(c, "change password", constants.MethodUpdate, nil, serviceOut.BaseOut.ErrorMessage, serviceOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "change password", constants.MethodUpdate, nil, "", http.StatusOK, trace.TraceId)

}

// GetUserProfile used to get user profile based on current logged user.
func (handler *UserHandler) GetUserProfile(c *gin.Context) {
	response := &api.GetUserProfileResp{}

	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(c, constants.TransactionTimeOut)
	defer cancel()

	// Get current user login role and id.
	userID := c.GetInt("ID")

	// Call service.
	svcOut := handler.UserService.GetUser(ctx,
		&userSvc.GetUserIn{
			Trace:  trace,
			UserID: uint(userID),
		},
	)

	if !svcOut.Success {
		api.ReturnResponse(c, "user", constants.MethodGet, response, svcOut.ErrorMessage, svcOut.ErrorCode, trace.TraceId)
		return
	}

	// Set response.
	response.User = api.ParseUser(svcOut.User)

	// Send response
	api.ReturnResponse(c, "user", constants.MethodGet, response, "", http.StatusOK, trace.TraceId)
}

// InsertFcmToken used to insert user's fcm token for firebase purpose.
func (handler *UserHandler) InsertFcmToken(c *gin.Context) {
	response := &api.InsertFcmTokenResp{}
	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}
	request := api.InsertFcmTokenReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "fcm token", constants.MethodCreate, request, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "fcm token", constants.MethodCreate, &request, ve, trace.TraceId)
		return
	}

	// Get current user login role and id
	userID := c.GetInt("ID")

	// Call service
	authOut := handler.UserService.InsertFcmToken(c,
		&userSvc.InsertFcmTokenIn{
			Trace:  trace,
			UserID: uint(userID),
			Token:  request.Token,
		})
	if !authOut.BaseOut.Success {
		api.ReturnResponse(c, "fcm token", constants.MethodCreate, response, authOut.BaseOut.ErrorMessage, authOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "fcm token", constants.MethodCreate, response, "", http.StatusOK, trace.TraceId)
}

// DeleteFcmToken used to insert user's fcm token for firebase purpose.
func (handler *UserHandler) DeleteFcmToken(c *gin.Context) {
	response := &api.DeleteFcmTokenResp{}
	trace := &trace.Trace{
		TraceId: uuid.New().String(),
	}
	request := api.DeleteFcmTokenReq{}
	if err := c.ShouldBind(&request); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ReturnInternalServerError(c, "fcm token", constants.MethodDelete, request, err, trace.TraceId)
			return
		}
		api.ReturnValidationError(c, "fcm token", constants.MethodDelete, &request, ve, trace.TraceId)
		return
	}

	// Call service
	authOut := handler.UserService.DeleteFcmTokenBulk(c,
		&userSvc.DeleteFcmTokenBulkIn{
			Trace:  trace,
			Tokens: []string{request.Token},
		})
	if !authOut.BaseOut.Success {
		api.ReturnResponse(c, "login", constants.MethodNull, response, authOut.BaseOut.ErrorMessage, authOut.BaseOut.ErrorCode, trace.TraceId)
		return
	}

	// Send response
	api.ReturnResponse(c, "fcm token", constants.MethodNull, response, "", http.StatusOK, trace.TraceId)
}
