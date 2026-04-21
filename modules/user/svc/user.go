package svc

import (
	"context"
	"errors"
	"net/http"
	"template-go/base/helpers"
	"template-go/data/enum"
	"template-go/data/model"
	user "template-go/modules/user/repository"
	"template-go/util/config"
	"template-go/util/logger"
	"time"

	"gorm.io/gorm"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type UserServiceImpl struct {
	UserStorage user.UserStorage
	Config      config.Config
	Logger      logger.Logger
}

func NewUserService(
	UserStorage user.UserStorage,
	Config config.Config,

) UserService {
	return &UserServiceImpl{
		UserStorage: UserStorage,
		Config:      Config,
		Logger:      logger.NewZerologLogger("UserService"),
	}
}

func (service *UserServiceImpl) LoginUser(ctx context.Context, paramIn *LoginUserIn) *LoginUserOut {
	resp := &LoginUserOut{}
	userRepo := service.UserStorage.BeginTx(ctx)
	defer userRepo.Rollback()

	// Find user based on its username
	user, err := userRepo.FindUserByUsername(paramIn.Username)
	if err != nil {
		service.Logger.WarnErr(paramIn.Trace, err).Msg("LoginUser(): invalid username or password")
		resp.ErrorMessage = "invalid username or password"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	// Validate password
	decryptPassword, err := helpers.Decrypt(service.Config.AuthSecret, user.Password)
	if err != nil || decryptPassword != paramIn.Password {
		// If possible we don't want to throw invalid server error to user.
		if err != nil {
			service.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): error when decrypting")
		}
		resp.ErrorMessage = "invalid username or password"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	// Generate token.
	token, err := service.generateToken(user.ID, user.Email, user.FullName, string(user.Role))
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): failed to create token")
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
	userRepo := service.UserStorage.BeginTx(ctx)
	defer userRepo.Rollback()

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
	encryptPassword, err := helpers.Encrypt(service.Config.AuthSecret, paramIn.Password)
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
	user.Password = encryptPassword
	user.FullName = paramIn.FullName
	user.Role = enum.Role(paramIn.UserRole)
	user.DistributorID = paramIn.DistributorID
	user.AreaID = paramIn.AreaID
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

	err = userRepo.Commit()
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (service *UserServiceImpl) ChangePasswordByExternalID(ctx context.Context, paramIn *ChangePasswordIn) *ChangePasswordOut {
	resp := &ChangePasswordOut{}

	userRepo := service.UserStorage.BeginTx(ctx)
	defer userRepo.Rollback()

	// Fetch current user.
	currentUser, err := userRepo.FindUserByID(paramIn.UserID)
	if err != nil {
		// Panic because it's mean there's might be leak in JWTToken key.
		service.Logger.PanicErr(paramIn.Trace, err).Msg("ChangePasswordByExternalID(): current user not found")
		resp.ErrorMessage = "user not found"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	// Fetch user sales based on externalID.
	sales, err := userRepo.FindUserByExternalID(paramIn.ExternalID)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangePasswordByExternalID(): sales not found")
		resp.ErrorMessage = "sales not found"
		resp.ErrorCode = http.StatusNotFound
		return resp
	}

	// Lock user to be updated.
	sales, err = userRepo.LockUser(sales.ID)
	if err != nil {
		service.Logger.Warn(paramIn.Trace).Msg("ChangePasswordByExternalID(): failed to lock user")
		resp.ErrorMessage = "failed to lock user"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	// Validate access.
	if sales.Role != enum.Sales || sales.AreaID != currentUser.AreaID {
		service.Logger.Warn(paramIn.Trace).Msg("ChangePasswordByExternalID(): current user has no access")
		resp.ErrorMessage = "current user has no access"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	// Encrypt the password.
	encryptPassword, err := helpers.Encrypt(service.Config.AuthSecret, paramIn.Password)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangePasswordByExternalID(): encrypt password failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	sales.Password = encryptPassword

	// UpdateUser sales password.
	err = userRepo.UpdateUser(sales)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangePasswordByExternalID(): failed to update sales's data")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	// TODO: Add audit log.

	err = userRepo.Commit()
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangePasswordByExternalID(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (service *UserServiceImpl) GetUser(ctx context.Context, paramIn *GetUserIn) *GetUserOut {
	resp := &GetUserOut{}
	userRepo := service.UserStorage.BeginTx(ctx)
	defer userRepo.Rollback()

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
	userRepo := service.UserStorage.BeginTx(ctx)
	defer userRepo.Rollback()

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

	err = userRepo.Commit()
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
	userRepo := service.UserStorage.BeginTx(ctx)
	defer userRepo.Rollback()

	err := userRepo.DeleteFcmTokenBulk(paramIn.Tokens)
	if err != nil {
		service.Logger.ErrorErr(paramIn.Trace, err).Msg("DeleteFcmTokenBulk(): failed to create fcmToken")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.Commit()
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
	userRepo := service.UserStorage.BeginTx(ctx)
	defer userRepo.Rollback()

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
