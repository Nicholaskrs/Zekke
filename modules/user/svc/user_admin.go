package svc

import (
	"context"
	"errors"
	"net/http"
	dto "template-go/base/base"
	"template-go/base/constants"
	"template-go/base/helpers"
	"template-go/core/telemetry/logger"
	"template-go/data/model"
	permissionSvc "template-go/modules/permission/svc"
	"template-go/modules/user/permission"
	userRepo "template-go/modules/user/repository"
	"template-go/util/config"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserAdminServiceImpl struct {
	UserStorage            userRepo.UserStorage
	PermissionAdminService permissionSvc.PermissionAdminService
	Config                 config.Config
	Logger                 logger.Logger
}

func NewUserAdminService(
	UserStorage userRepo.UserStorage,
	PermissionAdminService permissionSvc.PermissionAdminService,
	Config config.Config,
) UserAdminService {
	return &UserAdminServiceImpl{
		UserStorage:            UserStorage,
		PermissionAdminService: PermissionAdminService,
		Config:                 Config,
		Logger:                 logger.NewZerologLogger("UserService"),
	}
}

func (i *UserAdminServiceImpl) LoginUserAdmin(ctx context.Context, paramIn *LoginUserAdminIn) *LoginUserAdminOut {
	resp := &LoginUserAdminOut{}
	repo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	// Find user based on its username
	user, err := repo.FindUserAdminByUsername(paramIn.Username)
	if err != nil {
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

	getAdminGroupPrivilegeResp := i.PermissionAdminService.GetAdminGroupPrivilege(ctx, &permissionSvc.GetAdminGroupPrivilegeIn{
		Trace:       paramIn.Trace,
		UserAdminId: int(user.ID),
	})
	if getAdminGroupPrivilegeResp.Success == false {
		i.Logger.Error(paramIn.Trace).Msg("LoginUserAdmin(): get admin privilege failed")
		resp.ErrorMessage = getAdminGroupPrivilegeResp.ErrorMessage
		resp.ErrorCode = getAdminGroupPrivilegeResp.ErrorCode
		return resp
	}

	// Generate token.
	token, err := i.generateToken(user.ID, user.Email, user.FullName, getAdminGroupPrivilegeResp.AdminPrivilegeGroup)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUserAdmin(): failed to create token")
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
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): fail get current time")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	expire := now.Add(constants.RefreshTTL)

	err = repo.InsertRefreshToken(&model.RefreshToken{
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
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): failed to insert refresh token")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = repo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("LoginUser(): failed to commit")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.Token = token
	resp.UserID = user.ID
	resp.AdminGroupPrivilege = getAdminGroupPrivilegeResp.AdminPrivilegeGroup
	resp.FullName = user.FullName
	resp.Username = user.Username
	resp.ExpireIn = int(constants.AccessTTL.Seconds())
	resp.RefreshToken = refreshToken
	return resp
}

// generateToken used to generate token that used in Login. It sets claims filled with user's id, email, name, and privilege group.
func (i *UserAdminServiceImpl) generateToken(id uint, email string, name string, privilegeGroup string) (string, error) {
	claims := &AdminAuthCustomClaims{
		id,
		email,
		name,
		privilegeGroup,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    i.Config.JwtIssuer,
			IssuedAt:  time.Now().Unix(),
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

func (i *UserAdminServiceImpl) FilterUser(ctx context.Context, paramIn *AdminFilterUserIn) *AdminFilterUserOut {
	resp := &AdminFilterUserOut{}

	hasPrivilegeOut := i.PermissionAdminService.HasPrivilege(ctx, &permissionSvc.HasPrivilegeIn{
		Trace:          paramIn.Trace,
		PrivilegeGroup: paramIn.AdminPrivilegeGroup,
		Privilege:      permission.PVFilterUser,
	})

	if hasPrivilegeOut.Success == false {
		i.Logger.Error(paramIn.Trace).Msg("FilterUser(): failed to check user privilege")
		resp.ErrorMessage = hasPrivilegeOut.ErrorMessage
		resp.ErrorCode = hasPrivilegeOut.ErrorCode
		return resp
	}

	if hasPrivilegeOut.IsAllowed == false {
		resp.ErrorMessage = "User has no privilege"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	repo := i.UserStorage.NewUserRepositoryRead(ctx)

	filterUser := userRepo.FilterUser{
		PaginateIn: dto.PaginateIn{
			Limit: paramIn.Limit,
			Page:  paramIn.Page,
		},
		IsDeleted: false,
	}

	users, err := repo.FilterUser(filterUser)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterUser(): failed to filter user")
		resp.ErrorMessage = "failed to filter user"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	totalRows, err := repo.FilterUserCount(filterUser)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterUser(): failed to filter user count")
		resp.ErrorMessage = "failed to filter user count"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.Users = users
	resp.TotalRows = totalRows
	return resp
}

func (i *UserAdminServiceImpl) DeleteUser(ctx context.Context, paramIn *AdminDeleteUserIn) *AdminDeleteUserOut {
	resp := &AdminDeleteUserOut{}

	if paramIn.UserId <= 0 {
		resp.ErrorMessage = "UserId empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	hasPrivilegeOut := i.PermissionAdminService.HasPrivilege(ctx, &permissionSvc.HasPrivilegeIn{
		Trace:          paramIn.Trace,
		PrivilegeGroup: paramIn.AdminPrivilegeGroup,
		Privilege:      permission.PVDeleteUser,
	})
	if !hasPrivilegeOut.Success {
		resp.ErrorMessage = hasPrivilegeOut.ErrorMessage
		resp.ErrorCode = hasPrivilegeOut.ErrorCode
		return resp
	}
	if !hasPrivilegeOut.IsAllowed {
		resp.ErrorMessage = "User has no privilege"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	repo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	if err := repo.DeleteUserByID(uint(paramIn.UserId)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.ErrorMessage = err.Error()
			resp.ErrorCode = http.StatusNotFound
			return resp
		}
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("DeleteUser(): failed to delete user")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if err := repo.Commit(ctx); err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("DeleteUser(): failed to commit")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	return resp
}

func (i *UserAdminServiceImpl) Refresh(ctx context.Context, paramIn *AdminRefreshIn) *AdminRefreshOut {
	resp := &AdminRefreshOut{}
	h := helpers.HashToken(paramIn.RefreshToken)

	// Atomic: only one request can consume a given refresh token.
	repo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	refreshToken, err := repo.LockRefreshToken(paramIn.AdminUserId, h)
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

	user, err := repo.FindUserByID(refreshToken.UserID)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to use get userid")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	privilegeGroup, err := repo.GetAdminUserPrivilegeGroup(user.ID)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("Refresh(): failed to use get privilege group")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	token, err := i.generateToken(user.ID, user.Email, user.FullName, privilegeGroup)
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
		UserID:    user.ID,
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
	resp.ExpireIn = int(constants.AccessTTL.Seconds())
	return resp
}

func (i *UserAdminServiceImpl) ChangeAdminPassword(ctx context.Context, paramIn *AdminChangePasswordIn) *AdminChangePasswordOut {
	resp := &AdminChangePasswordOut{}

	if paramIn.UserAdminId == 0 {
		resp.ErrorMessage = "UserAdminId empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.NewPassword == "" {
		resp.ErrorMessage = "NewPassword empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.OldPassword == "" {
		resp.ErrorMessage = "OldPassword empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	repo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	user, err := repo.FindUserByID(uint(paramIn.UserAdminId))
	if err != nil {
		i.Logger.WarnErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): invalid userid")
		resp.ErrorMessage = "invalid username or password"
		resp.ErrorCode = http.StatusUnauthorized
		return resp
	}

	// Validate password
	isVerified := helpers.VerifyPassword(paramIn.OldPassword, user.Password)
	if !isVerified {
		resp.ErrorMessage = "invalid old password"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	user.Password, err = helpers.HashPassword(paramIn.NewPassword)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): hash password failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = repo.UpdateUser(user)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): update user failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	getAdminGroupPrivilegeResp := i.PermissionAdminService.GetAdminGroupPrivilege(ctx, &permissionSvc.GetAdminGroupPrivilegeIn{
		Trace:       paramIn.Trace,
		UserAdminId: int(user.ID),
	})
	if getAdminGroupPrivilegeResp.Success == false {
		i.Logger.Error(paramIn.Trace).Msg("ChangeAdminPassword(): get admin privilege failed")
		resp.ErrorMessage = getAdminGroupPrivilegeResp.ErrorMessage
		resp.ErrorCode = getAdminGroupPrivilegeResp.ErrorCode
		return resp
	}

	// Token is reused possibly stolen revoke everything that generated by this familyID.
	err = repo.RevokeRefreshTokenByUserID(uint(paramIn.UserAdminId))
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): failed to revoke refresh token by user id")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	refreshToken, err := helpers.RandomToken()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): failed to create random token")
		resp.ErrorMessage = "generate refresh token failed"
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	now, err := helpers.TimeNow()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): fail get current time")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	expire := now.Add(constants.RefreshTTL)

	err = repo.InsertRefreshToken(&model.RefreshToken{
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
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): failed to insert refresh token")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = repo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPassword(): failed to commit")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.RefreshToken = refreshToken
	resp.ExpireIn = int(constants.AccessTTL.Seconds())

	return resp
}

func (i *UserAdminServiceImpl) RegisterNewAdmin(ctx context.Context, paramIn *RegisterNewAdminIn) *RegisterNewAdminOut {
	resp := &RegisterNewAdminOut{}

	getAdminGroupPrivilegeResp := i.PermissionAdminService.CheckAdminGroupPrivilegeExists(ctx, &permissionSvc.CheckAdminGroupPrivilegeExistsIn{
		Trace:               paramIn.Trace,
		AdminPrivilegeGroup: paramIn.PrivilegeGroup,
	})
	if getAdminGroupPrivilegeResp.Success == false {
		i.Logger.Error(paramIn.Trace).Msg("RegisterNewAdmin(): get admin privilege failed")
		resp.ErrorMessage = getAdminGroupPrivilegeResp.ErrorMessage
		resp.ErrorCode = getAdminGroupPrivilegeResp.ErrorCode
		return resp
	}
	if getAdminGroupPrivilegeResp.IsExists == false {
		resp.ErrorMessage = "PrivilegeGroup doesn't exists"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	hasPrivilegeOut := i.PermissionAdminService.HasPrivilege(ctx, &permissionSvc.HasPrivilegeIn{
		Trace:          paramIn.Trace,
		PrivilegeGroup: paramIn.AdminPrivilegeGroup,
		Privilege:      permission.PVRegisterAdminUser,
	})
	if hasPrivilegeOut.Success == false {
		i.Logger.Error(paramIn.Trace).Msg("RegisterNewAdmin(): check admin privilege failed")
		resp.ErrorMessage = getAdminGroupPrivilegeResp.ErrorMessage
		resp.ErrorCode = getAdminGroupPrivilegeResp.ErrorCode
		return resp
	}
	if hasPrivilegeOut.IsAllowed == false {
		resp.ErrorMessage = "User has no privilege"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	repo := i.UserStorage.NewUserRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	passwordHash, err := helpers.HashPassword(paramIn.Password)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("RegisterNewAdmin(): hash password failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	now, err := helpers.TimeNow()
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("RegisterNewAdmin(): fail get current time")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	isExists, err := repo.CheckUsernameExists(paramIn.Username)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("RegisterNewAdmin(): check username failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	if isExists {
		resp.ErrorMessage = "Username is already exists"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	isExists, err = repo.CheckEmailExists(paramIn.Email)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("RegisterNewAdmin(): check email failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	if isExists {
		resp.ErrorMessage = "Email is already exists"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	newAdminId, err := repo.CreateUser(&model.User{
		ID:         0,
		ExternalID: uuid.New().String(),
		Username:   paramIn.Username,
		Email:      paramIn.Email,
		Password:   passwordHash,
		FullName:   paramIn.FullName,
		Role:       "Admin",
		Timestamp: &model.Timestamp{
			CreatedTs:     now,
			LastUpdatedTs: now,
		},
	})
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("RegisterNewAdmin(): insert user failed")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	assignPermissionToAdminUserOut := i.PermissionAdminService.ChangeAdminPrivilegeRole(ctx, &permissionSvc.ChangeAdminPrivilegeRoleIn{
		Trace:               paramIn.Trace,
		UserAdminId:         paramIn.UserAdminId,
		AdminPrivilegeGroup: paramIn.AdminPrivilegeGroup,
		TargetUserID:        int(newAdminId),
		NewPrivilegeGroup:   paramIn.PrivilegeGroup,
	})

	if assignPermissionToAdminUserOut.Success == false {
		resp.ErrorMessage = assignPermissionToAdminUserOut.ErrorMessage
		resp.ErrorCode = assignPermissionToAdminUserOut.ErrorCode
		return resp
	}

	err = repo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("RegisterNewAdmin(): failed to commit")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true

	return resp
}
