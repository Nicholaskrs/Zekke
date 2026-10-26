package svc

import (
	"context"
	"errors"
	"net/http"
	"template-go/base/constants"
	"template-go/base/helpers"
	"template-go/core/telemetry/logger"
	"template-go/data/model"
	user "template-go/modules/user/repository"
	"template-go/util/config"

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

func (i *UserServiceImpl) LoginUser(ctx context.Context, paramIn *LoginUserIn) *LoginUserOut {
	resp := &LoginUserOut{}
	userRepo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer userRepo.Rollback(ctx)

	// Find user based on its username
	user, err := userRepo.FindUserByUsername(paramIn.Username)
	if err != nil {
		i.Logger.WarnErr(paramIn.Trace, err).Msg("LoginUser(): invalid username or password")
		resp.ErrorMessage = "invalid username or password"
		resp.ErrorCode = http.StatusUnauthorized
		return resp
	}

	// Validate password
	isVerified := helpers.VerifyPassword(paramIn.Password, user.Password)
	if !isVerified {
		resp.ErrorMessage = "invalid username or password"
		resp.ErrorCode = http.StatusUnauthorized
		return resp
	}

	// Generate token.
	token, err := i.generateToken(user.ID, user.Email, user.FullName, user.Role)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): failed to create token")
		resp.ErrorMessage = "generate token failed"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	refreshToken, err := helpers.RandomToken()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): failed to create random token")
		resp.ErrorMessage = "generate refresh token failed"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	now, err := helpers.TimeNow()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): fail get current time")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	expire := now.Add(constants.RefreshTTL)

	err = userRepo.InsertRefreshToken(&model.RefreshToken{
		UserID:    user.ID,
		FamilyID:  uuid.NewString(),
		TokenHash: helpers.HashToken(refreshToken),
		ExpireAt:  &expire,
		Timestamp: &model.Timestamp{
			CreatedTs:     now,
			LastUpdatedTs: now,
		},
	})
	if err != nil {
		i.Logger.WarnErr(paramIn.Trace, err).Msg("LoginUser(): failed to insert refresh token")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.Commit(ctx)
	if err != nil {
		i.Logger.WarnErr(paramIn.Trace, err).Msg("LoginUser(): failed to commit")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.Token = token
	resp.UserID = user.ID
	resp.UserRole = user.Role
	resp.FullName = user.FullName
	resp.Username = user.Username
	resp.RefreshToken = refreshToken
	return resp
}

// generateToken used to generate token that used in Login. It sets claims filled with user's id, email, name, and role.
func (i *UserServiceImpl) generateToken(id uint, email string, name string, role string) (string, error) {
	now, err := helpers.TimeNow()
	if err != nil {
		return "", err
	}
	claims := &AuthCustomClaims{
		id,
		email,
		name,
		role,
		jwt.StandardClaims{
			ExpiresAt: now.Add(constants.AccessTTL).Unix(),
			Issuer:    i.Config.JwtIssuer,
			IssuedAt:  now.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//encoded string
	tokenStr, err := token.SignedString([]byte(i.Config.JwtSecret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func (i *UserServiceImpl) Register(ctx context.Context, paramIn *UserRegisterIn) *UserRegisterOut {
	resp := &UserRegisterOut{}
	userRepo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer userRepo.Rollback(ctx)

	// Validate role
	isRoleExists, err := userRepo.CheckRoleExists(paramIn.UserRole)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): failed to check role exists")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if !isRoleExists {
		i.Logger.Warn(paramIn.Trace).Msg("Register(): role not found")
		resp.ErrorMessage = "role not found"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	// Check if user already exists based on its username. If yes, return error indicating that the username cannot be duplicated.
	isUserExists, err := userRepo.CheckUserExistsByUsername(paramIn.Username)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): failed to check user by username")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isUserExists {
		i.Logger.Warn(paramIn.Trace).Msg("Register(): username already exists")
		resp.ErrorMessage = "username already exists"
		resp.ErrorCode = http.StatusUnprocessableEntity
		return resp
	}

	// Encrypt the password.
	hashedPassword, err := helpers.HashPassword(paramIn.Password)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): encrypt password failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	// Set user entity.
	now, err := helpers.TimeNow()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): fail get current time")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	var user model.User
	user.ExternalID = uuid.New().String()
	user.Username = paramIn.Username
	user.Email = paramIn.Email
	user.Password = hashedPassword
	user.FullName = paramIn.FullName
	user.Role = paramIn.UserRole
	user.Timestamp = &model.Timestamp{
		CreatedTs:     now,
		LastUpdatedTs: now,
	}

	_, err = userRepo.CreateUser(&user)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): failed to create user")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (i *UserServiceImpl) GetUser(ctx context.Context, paramIn *GetUserIn) *GetUserOut {
	resp := &GetUserOut{}
	userRepo := i.UserStorage.NewUserRepositoryRead(ctx)

	user, err := userRepo.FindUserByID(paramIn.UserID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			i.Logger.ErrorErr(paramIn.Trace, err).
				Int("UserID", int(paramIn.UserID)).
				Msg("GetUser(): UserID Not found")
			resp.ErrorMessage = err.Error()
			resp.ErrorCode = http.StatusNotFound
			return resp
		}
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("GetUser(): failed to FindUserByID")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusNotFound
		return resp
	}

	resp.Success = true
	resp.User = user
	return resp
}

func (i *UserServiceImpl) InsertFcmToken(ctx context.Context, paramIn *InsertFcmTokenIn) *InsertFcmTokenOut {
	resp := &InsertFcmTokenOut{}
	userRepo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer userRepo.Rollback(ctx)

	// Check if user already exists based on its username. If yes, return error indicating that the username cannot be duplicated.
	usr, err := userRepo.FindUserByID(paramIn.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			i.Logger.ErrorErr(paramIn.Trace, err).
				Int("UserID", int(paramIn.UserID)).
				Msg("InsertFcmToken(): UserID Not found")
			resp.ErrorMessage = err.Error()
			resp.ErrorCode = http.StatusNotFound
			return resp
		}
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("CreateFcmToken(): failed to check userID exists")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if usr == nil {
		i.Logger.Warn(paramIn.Trace).Msg("CreateFcmToken(): user not found")
		resp.ErrorMessage = "user not found"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	// Set user entity.
	now, err := helpers.TimeNow()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Register(): fail get current time")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.CreateFcmToken(&model.FcmToken{
		UserID:   paramIn.UserID,
		FcmToken: paramIn.Token,
		Timestamp: &model.Timestamp{
			CreatedTs:     now,
			LastUpdatedTs: now,
		},
	})
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("CreateFcmToken(): failed to create fcmToken")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("CreateFcmToken(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (i *UserServiceImpl) DeleteFcmTokenBulk(ctx context.Context, paramIn *DeleteFcmTokenBulkIn) *DeleteFcmTokenBulkOut {
	resp := &DeleteFcmTokenBulkOut{}
	userRepo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer userRepo.Rollback(ctx)

	err := userRepo.DeleteFcmTokenBulk(paramIn.Tokens)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("DeleteFcmTokenBulk(): failed to create fcmToken")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = userRepo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("DeleteFcmTokenBulk(): commit failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (i *UserServiceImpl) GetUserFcmToken(ctx context.Context, paramIn *GetUserFcmTokenIn) *GetUserFcmTokenOut {
	resp := &GetUserFcmTokenOut{}
	userRepo := i.UserStorage.NewUserRepositoryRead(ctx)

	// Check if user already exists based on its username. If yes, return error indicating that the username cannot be duplicated.
	isUserExists, err := userRepo.FindUserByID(paramIn.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.ErrorCode = http.StatusNotFound
		} else {
			resp.ErrorCode = http.StatusInternalServerError
		}
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("GetUserFcmToken(): failed to check userID exists")
		resp.ErrorMessage = err.Error()
		return resp
	}

	if isUserExists == nil {
		i.Logger.Warn(paramIn.Trace).Msg("GetUserFcmToken(): user not found")
		resp.ErrorMessage = "user not found"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	fcmTokens, err := userRepo.GetUserFcmToken(paramIn.UserID)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("GetUserFcmToken(): failed to create fcmToken")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.FcmTokens = fcmTokens
	resp.Success = true
	return resp
}

func (i *UserServiceImpl) Refresh(ctx context.Context, paramIn *RefreshIn) *RefreshOut {
	resp := &RefreshOut{}
	h := helpers.HashToken(paramIn.RefreshToken)

	// Atomic: only one request can consume a given refresh token.
	repo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	refreshToken, err := repo.LockRefreshToken(paramIn.UserId, h)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.ErrorMessage = "Refresh token not found"
			resp.ErrorCode = http.StatusNotFound
			i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to lock refresh token")
			return resp
		}
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to lock refresh token")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	now, err := helpers.TimeNow()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to read refresh token")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if refreshToken.ExpireAt.Before(now) || refreshToken.RevokedAt != nil {
		resp.ErrorMessage = "Token is expired"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	if refreshToken.UsedAt != nil {
		// Token is reused possibly stolen revoke everything that generated by this familyID.
		err = repo.RevokeRefreshTokenByFamilyID(refreshToken.FamilyID)
		if err != nil {
			i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to revoke refresh token")
			resp.ErrorMessage = err.Error()
			resp.ErrorCode = http.StatusInternalServerError
			return resp
		}
		if err = repo.Commit(ctx); err != nil {
			i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to commit token revocation")
			resp.ErrorMessage = err.Error()
			resp.ErrorCode = http.StatusInternalServerError
			return resp
		}

		resp.ErrorMessage = "Token is already used"
		resp.ErrorCode = http.StatusForbidden
		return resp

	}

	err = repo.UseRefreshToken(int(refreshToken.ID))
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to use refresh token")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	usr, err := repo.FindUserByID(refreshToken.UserID)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to use get userid")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	token, err := i.generateToken(usr.ID, usr.Email, usr.FullName, usr.Role)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to create token")
		resp.ErrorMessage = "generate token failed"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	expire := now.Add(constants.RefreshTTL)
	newRefreshToken, err := helpers.RandomToken()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to generate random token")
		resp.ErrorMessage = "generate refresh token failed"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = repo.InsertRefreshToken(&model.RefreshToken{
		UserID:    usr.ID,
		FamilyID:  refreshToken.FamilyID,
		TokenHash: helpers.HashToken(newRefreshToken),
		ExpireAt:  &expire,
		Timestamp: &model.Timestamp{
			CreatedTs:     now,
			LastUpdatedTs: now,
		},
	})
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to insert refresh token")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = repo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to commit")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.RefreshToken = newRefreshToken
	resp.AccessToken = token
	resp.ExpireIn = int(constants.RefreshTTL.Seconds())
	return resp
}
