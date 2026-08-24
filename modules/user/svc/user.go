package svc

import (
	"context"
	"errors"
	"net/http"
	"template-go/base/helpers"
	logger2 "template-go/core/telemetry/logger"
	"template-go/data/enum"
	"template-go/data/model"
	user "template-go/modules/user/repository"
	"template-go/util/config"
	"time"

	"gorm.io/gorm"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type UserServiceImpl struct {
	UserStorage user.UserStorage
	Config      config.Config
	Logger      logger2.Logger
}

func NewUserService(
	UserStorage user.UserStorage,
	Config config.Config,

) UserService {
	return &UserServiceImpl{
		UserStorage: UserStorage,
		Config:      Config,
		Logger:      logger2.NewZerologLogger("UserService"),
	}
}

func (service *UserServiceImpl) LoginUser(ctx context.Context, paramIn *LoginUserIn) *LoginUserOut {
	resp := &LoginUserOut{}
	userRepo := service.UserStorage.NewUserRepositoryRead(ctx)

	// Find user based on its username
	user, err := userRepo.FindUserByUsername(paramIn.Username)
	if err != nil {
		service.Logger.WarnErr(paramIn.Trace, err).Msg("LoginUser(): invalid username or password")
		resp.ErrorMessage = "invalid username or password"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	// Validate password
	isVerified := helpers.VerifyPassword(paramIn.Password, user.Password)
	if !isVerified {
		resp.ErrorMessage = "invalid username or password"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	// Generate token.
	token, err := service.generateToken(user.ID, user.Email, user.FullName, string(user.Role))
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): failed to create token")
		resp.ErrorMessage = "generate token failed"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.Token = token
	resp.UserID = user.ID
	resp.UserRole = string(user.Role)
	resp.FullName = user.FullName
	resp.Username = user.Username
	return resp
}

// generateToken used to generate token that used in Login. It sets claims filled with user's id, email, name, and role.
func (service *UserServiceImpl) generateToken(id uint, email string, name string, role string) (string, error) {
	claims := &AuthCustomClaims{
		id,
		email,
		name,
		role,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    service.Config.JwtIssuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//encoded string
	tokenStr, err := token.SignedString([]byte(service.Config.JwtSecret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func (service *UserServiceImpl) Register(ctx context.Context, paramIn *UserRegisterIn) *UserRegisterOut {
	resp := &UserRegisterOut{}
	userRepo := service.UserStorage.NewUserRepositoryWrite(ctx)
	defer userRepo.Rollback(ctx)

	// Validate role
	if !helpers.InArray(enum.SliceRole, paramIn.UserRole) {
		service.Logger.Warn(paramIn.Trace).Msg("Register(): invalid role type")
		resp.ErrorMessage = "invalid role type"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	// Check if user already exists based on its username. If yes, return error indicating that the username cannot be duplicated.
	isUserExists, err := userRepo.CheckUserExistsByUsername(paramIn.Username)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): failed to check user by username")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isUserExists {
		service.Logger.Warn(paramIn.Trace).Msg("Register(): username already exists")
		resp.ErrorMessage = "username already exists"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	// Encrypt the password.
	hashedPassword, err := helpers.HashPassword(paramIn.Password)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): encrypt password failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	// Set user entity.
	now := time.Now().Local()
	var user model.User
	user.ExternalID = uuid.New().String()
	user.Username = paramIn.Username
	user.Email = paramIn.Email
	user.Password = hashedPassword
	user.FullName = paramIn.FullName
	user.Role = enum.Role(paramIn.UserRole)
	user.Timestamp = &model.Timestamp{
		CreatedTs:     now,
		LastUpdatedTs: now,
	}

	_, err = userRepo.CreateUser(&user)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): failed to create user")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	// TODO: Add audit log.

	err = userRepo.Commit(ctx)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (service *UserServiceImpl) GetUser(ctx context.Context, paramIn *GetUserIn) *GetUserOut {
	resp := &GetUserOut{}
	userRepo := service.UserStorage.NewUserRepositoryRead(ctx)

	user, err := userRepo.FindUserByID(paramIn.UserID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			service.Logger.ErrorErr(paramIn.Trace, err).
				Int("UserID", int(paramIn.UserID)).
				Msg("GetUser(): UserID Not found")
			resp.ErrorMessage = err.Error()
			resp.ErrorCode = http.StatusNotFound
			return resp
		}
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("GetUser(): failed to FindUserByID")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusNotFound
		return resp
	}

	resp.Success = true
	resp.User = user
	return resp
}

func (service *UserServiceImpl) InsertFcmToken(ctx context.Context, paramIn *InsertFcmTokenIn) *InsertFcmTokenOut {
	resp := &InsertFcmTokenOut{}
	userRepo := service.UserStorage.NewUserRepositoryWrite(ctx)
	defer userRepo.Rollback(ctx)

	// Check if user already exists based on its username. If yes, return error indicating that the username cannot be duplicated.
	usr, err := userRepo.FindUserByID(paramIn.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			service.Logger.ErrorErr(paramIn.Trace, err).
				Int("UserID", int(paramIn.UserID)).
				Msg("InsertFcmToken(): UserID Not found")
			resp.ErrorMessage = err.Error()
			resp.ErrorCode = http.StatusNotFound
			return resp
		}
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("CreateFcmToken(): failed to check userID exists")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if usr == nil {
		service.Logger.Warn(paramIn.Trace).Msg("CreateFcmToken(): user not found")
		resp.ErrorMessage = "user not found"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	// Set user entity.
	now := time.Now()

	err = userRepo.CreateFcmToken(&model.FcmToken{
		UserID:   paramIn.UserID,
		FcmToken: paramIn.Token,
		Timestamp: &model.Timestamp{
			CreatedTs:     now,
			LastUpdatedTs: now,
		},
	})
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("CreateFcmToken(): failed to create fcmToken")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.Commit(ctx)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("CreateFcmToken(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (service *UserServiceImpl) DeleteFcmTokenBulk(ctx context.Context, paramIn *DeleteFcmTokenBulkIn) *DeleteFcmTokenBulkOut {
	resp := &DeleteFcmTokenBulkOut{}
	userRepo := service.UserStorage.NewUserRepositoryWrite(ctx)
	defer userRepo.Rollback(ctx)

	err := userRepo.DeleteFcmTokenBulk(paramIn.Tokens)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("DeleteFcmTokenBulk(): failed to create fcmToken")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.Commit(ctx)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("DeleteFcmTokenBulk(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (service *UserServiceImpl) GetUserFcmToken(ctx context.Context, paramIn *GetUserFcmTokenIn) *GetUserFcmTokenOut {
	resp := &GetUserFcmTokenOut{}
	userRepo := service.UserStorage.NewUserRepositoryRead(ctx)

	// Check if user already exists based on its username. If yes, return error indicating that the username cannot be duplicated.
	isUserExists, err := userRepo.FindUserByID(paramIn.UserID)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("GetUserFcmToken(): failed to check userID exists")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isUserExists == nil {
		service.Logger.Warn(paramIn.Trace).Msg("GetUserFcmToken(): user not found")
		resp.ErrorMessage = "user not found"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	fcmTokens, err := userRepo.GetUserFcmToken(paramIn.UserID)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("GetUserFcmToken(): failed to create fcmToken")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.FcmTokens = fcmTokens
	resp.Success = true
	return resp
}
