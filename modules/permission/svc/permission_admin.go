package svc

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	dto "template-go/base/base"
	"template-go/core/telemetry/logger"
	"template-go/data/model"
	permissionConst "template-go/modules/permission/permission"
	permission "template-go/modules/permission/repository"
)

type PermissionAdminServiceImpl struct {
	Repository permission.PermissionStorage
	Logger     logger.Logger

	mu    sync.RWMutex
	cache map[string]map[string]bool
}

func NewPermissionAdminService(
	repository permission.PermissionStorage,
) PermissionAdminService {
	return &PermissionAdminServiceImpl{
		Repository: repository,
		Logger:     logger.NewZerologLogger("PermissionAdminService"),
	}
}

func (i *PermissionAdminServiceImpl) ChangeAdminPrivilegeRole(ctx context.Context, paramIn *ChangeAdminPrivilegeRoleIn) *ChangeAdminPrivilegeRoleOut {
	resp := &ChangeAdminPrivilegeRoleOut{}
	if paramIn.UserAdminId == 0 {
		resp.ErrorMessage = "UserAdminId empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.AdminPrivilegeGroup == "" {
		resp.ErrorMessage = "AdminPrivilegeGroup empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.TargetUserID == 0 {
		resp.ErrorMessage = "TargetUserID empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.NewPrivilegeGroup == "" {
		resp.ErrorMessage = "NewPrivilegeGroup empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	repo := i.Repository.NewPermissionRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	isAllowed, err := i.hasPrivilege(repo.PermissionRepositoryRead, paramIn.AdminPrivilegeGroup, permissionConst.PVChangeAdminPrivilegeRole)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPrivilegeRole(): fail to get check permission")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isAllowed == false {
		resp.ErrorMessage = "User has no permission"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	exists, err := repo.CheckPrivilegeGroupExists(paramIn.NewPrivilegeGroup)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPrivilegeRole(): fail to check if privilege group exists")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if !exists {
		resp.ErrorMessage = "PrivilegeGroup doesn't exists"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	err = repo.UpsertUserPrivilegeGroup(&model.UserPrivilegeGroup{
		UserID:             uint(paramIn.TargetUserID),
		PrivilegeGroupCode: paramIn.NewPrivilegeGroup,
	})
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPrivilegeRole(): fail to upsert user privilege group")
		resp.ErrorCode = http.StatusInternalServerError
		resp.ErrorMessage = err.Error()
		return resp
	}

	err = repo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("ChangeAdminPrivilegeRole(): fail to commit user privilege group")
		resp.ErrorCode = http.StatusInternalServerError
		resp.ErrorMessage = err.Error()
		return resp
	}

	resp.Success = true

	return resp
}

func (i *PermissionAdminServiceImpl) InsertNewPrivilegeGroup(ctx context.Context, paramIn *InsertNewPrivilegeGroupIn) *InsertNewPrivilegeGroupOut {
	resp := &InsertNewPrivilegeGroupOut{}
	if paramIn.UserAdminId == 0 {
		resp.ErrorMessage = "UserAdminId empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.AdminPrivilegeGroup == "" {
		resp.ErrorMessage = "AdminPrivilegeGroup empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.PrivilegeGroup == "" {
		resp.ErrorMessage = "NewPrivilegeGroup empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	repo := i.Repository.NewPermissionRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	isAllowed, err := i.hasPrivilege(repo.PermissionRepositoryRead, paramIn.AdminPrivilegeGroup, permissionConst.PVInsertPrivilegeGroup)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("InsertNewPrivilegeGroup(): fail to get check permission")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isAllowed == false {
		resp.ErrorMessage = "User has no permission"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	err = repo.InsertPrivilegeGroup(&model.PrivilegeGroup{
		Code:        paramIn.PrivilegeGroup,
		Description: paramIn.Description,
	})
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("InsertNewPrivilegeGroup(): fail to insert privilege group")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	err = repo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("InsertNewPrivilegeGroup(): fail to commit")
		resp.ErrorCode = http.StatusInternalServerError
		resp.ErrorMessage = err.Error()
		return resp
	}
	resp.Success = true

	return resp
}

func (i *PermissionAdminServiceImpl) AssignPrivilegeToGroup(ctx context.Context, paramIn *AssignPrivilegeToGroupIn) *AssignPrivilegeToGroupOut {
	resp := &AssignPrivilegeToGroupOut{}

	if paramIn.AdminPrivilegeGroup == "" {
		resp.ErrorMessage = "AdminPrivilegeGroup empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.UserAdminId == 0 {
		resp.ErrorMessage = "UserAdminId empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.PrivilegeCode == "" {
		resp.ErrorMessage = "PrivilegeCode empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.TargetPrivilegeGroup == "" {
		resp.ErrorMessage = "TargetPrivilegeGroup empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	repo := i.Repository.NewPermissionRepositoryWrite(ctx)
	defer repo.Rollback(ctx)

	allowed, err := i.hasPrivilege(repo.PermissionRepositoryRead, paramIn.AdminPrivilegeGroup, permissionConst.PVAssignPrivilegeToGroup)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("AssignPrivilegeToGroup(): fail to get check permission")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	if allowed == false {
		resp.ErrorMessage = "Unauthorized"
		resp.ErrorCode = http.StatusUnauthorized
		return resp
	}

	exists, err := repo.CheckPrivilegeGroupExists(paramIn.TargetPrivilegeGroup)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("AssignPrivilegeToGroup(): fail to check if privilege group exists")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if !exists {
		resp.ErrorMessage = "PrivilegeGroup doesn't exists"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	err = repo.AssignPrivilegeToPrivilegeGroup(&model.PrivilegeGroupPrivilege{
		PrivilegeGroupCode: paramIn.TargetPrivilegeGroup,
		Privilege:          paramIn.PrivilegeCode,
	})
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("AssignPrivilegeToGroup(): fail to assign privilege group")
		resp.ErrorCode = http.StatusInternalServerError
		resp.ErrorMessage = err.Error()
		return resp
	}

	err = repo.Commit(ctx)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("AssignPrivilegeToGroup(): fail to commit")
		resp.ErrorCode = http.StatusInternalServerError
		resp.ErrorMessage = err.Error()
		return resp
	}

	resp.Success = true

	// Clearing privilege cache could fire and forget because it might be used so in this case we'll just fire and forget.
	go func() {
		i.clearPrivilegeCache()
	}()

	return resp
}

func (i *PermissionAdminServiceImpl) HasPrivilege(ctx context.Context, paramIn *HasPrivilegeIn) *HasPrivilegeOut {
	resp := &HasPrivilegeOut{}

	if paramIn.PrivilegeGroup == "" {
		resp.ErrorMessage = "PrivilegeGroup Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.Privilege == "" {
		resp.ErrorMessage = "Privilege Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	permissionReadRepo := i.Repository.NewPermissionRepositoryRead(ctx)

	isAllowed, err := i.hasPrivilege(permissionReadRepo, paramIn.PrivilegeGroup, paramIn.Privilege)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("HasPrivilege(): fail to get permission map")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}
	resp.IsAllowed = isAllowed
	resp.Success = true
	return resp
}

// GetAdminGroupPrivilege has no privilege group check because it's used to get user's permission group.
func (i *PermissionAdminServiceImpl) GetAdminGroupPrivilege(ctx context.Context, paramIn *GetAdminGroupPrivilegeIn) *GetAdminGroupPrivilegeOut {
	resp := &GetAdminGroupPrivilegeOut{}

	if paramIn.UserAdminId == 0 {
		resp.ErrorMessage = "UserAdminId Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	permissionRepo := i.Repository.NewPermissionRepositoryRead(ctx)

	userPrivilegeGroup, err := permissionRepo.GetUserPrivilegeGroup(paramIn.UserAdminId)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("GetAdminGroupPrivilege(): failed to get admin group privileges")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.AdminPrivilegeGroup = userPrivilegeGroup
	return resp
}

func (i *PermissionAdminServiceImpl) CheckAdminGroupPrivilegeExists(ctx context.Context, paramIn *CheckAdminGroupPrivilegeExistsIn) *CheckAdminGroupPrivilegeExistsOut {
	resp := &CheckAdminGroupPrivilegeExistsOut{}

	if paramIn.AdminPrivilegeGroup == "" {
		resp.ErrorMessage = "AdminPrivilegeGroup Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	permissionRepo := i.Repository.NewPermissionRepositoryRead(ctx)

	isAllowed, err := i.hasPrivilege(permissionRepo, paramIn.AdminPrivilegeGroup, permissionConst.PVCheckAdminGroupPrivilegeExists)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("CheckAdminGroupPrivilegeExists(): fail to get check permission")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isAllowed == false {
		fmt.Printf("%v \n", paramIn.AdminPrivilegeGroup)
		resp.ErrorMessage = "User has no permission"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	isExists, err := permissionRepo.CheckPrivilegeGroupExists(paramIn.AdminPrivilegeGroup)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("CheckAdminGroupPrivilegeExists(): failed to get check group privileges")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.IsExists = isExists
	return resp
}

func (i *PermissionAdminServiceImpl) FilterPrivilegeGroup(ctx context.Context, paramIn *FilterPrivilegeGroupIn) *FilterPrivilegeGroupOut {
	resp := &FilterPrivilegeGroupOut{}

	if paramIn.AdminPrivilegeGroup == "" {
		resp.ErrorMessage = "AdminPrivilegeGroup Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.AdminUserId == 0 {
		resp.ErrorMessage = "AdminUserId Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.Page <= 0 {
		resp.ErrorMessage = "Page Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.Limit <= 0 {
		resp.ErrorMessage = "Limit Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	permissionRepo := i.Repository.NewPermissionRepositoryRead(ctx)

	isAllowed, err := i.hasPrivilege(permissionRepo, paramIn.AdminPrivilegeGroup, permissionConst.PVFilterPrivilegeGroup)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterPrivilegeGroup(): fail to get check permission")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isAllowed == false {
		resp.ErrorMessage = "User has no permission"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	filter := permission.FilterPrivilegeGroup{
		PaginateIn: dto.PaginateIn{
			Limit: paramIn.Limit,
			Page:  paramIn.Page,
		},
	}

	privilegeGroups, err := permissionRepo.FilterPrivilegeGroup(filter)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterPrivilegeGroup(): failed to filter group")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	count, err := permissionRepo.FilterPrivilegeGroupCount(filter)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterPrivilegeGroup(): failed to count group filter")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.List = privilegeGroups
	resp.TotalRows = count
	return resp
}

// hasPrivilege reports whether the privilege is granted to the group.
// It reads from the cache and loads it with the given repo on a cache miss.
func (i *PermissionAdminServiceImpl) hasPrivilege(repo *permission.PermissionRepositoryRead, group, privilege string) (bool, error) {
	// Fast path: read from the cache.
	i.mu.RLock()
	if i.cache != nil {
		allowed := i.cache[group][privilege]
		i.mu.RUnlock()
		return allowed, nil
	}
	i.mu.RUnlock()

	// Cache miss: load from the repository.
	privileges, err := repo.GetPrivilegeGroupPrivileges()
	if err != nil {
		return false, err
	}

	// Only one concurrent initializer publishes the cache.
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.cache == nil {
		i.cache = privileges
	}
	return i.cache[group][privilege], nil
}

// hasPrivilege reports whether the privilege is granted to the group.
// It reads from the cache and loads it with the given repo on a cache miss.
func (i *PermissionAdminServiceImpl) clearPrivilegeCache() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.cache = nil
}

func (i *PermissionAdminServiceImpl) FilterPrivilegeGroupPrivileges(
	ctx context.Context,
	paramIn *FilterPrivilegeGroupPrivilegesIn,
) *FilterPrivilegeGroupPrivilegesOut {
	resp := &FilterPrivilegeGroupPrivilegesOut{}

	if paramIn.AdminPrivilegeGroup == "" {
		resp.ErrorMessage = "AdminPrivilegeGroup Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.AdminUserId == 0 {
		resp.ErrorMessage = "AdminUserId Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.PrivilegeGroup == "" {
		resp.ErrorMessage = "PrivilegeGroup Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.Page <= 0 {
		resp.ErrorMessage = "Page Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	if paramIn.Limit <= 0 {
		resp.ErrorMessage = "Limit Empty"
		resp.ErrorCode = http.StatusBadRequest
		return resp
	}

	permissionRepo := i.Repository.NewPermissionRepositoryRead(ctx)

	isAllowed, err := i.hasPrivilege(permissionRepo, paramIn.AdminPrivilegeGroup, permissionConst.PVFilterPrivilege)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterPrivilege(): fail to get check permission")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	if isAllowed == false {
		resp.ErrorMessage = "User has no permission"
		resp.ErrorCode = http.StatusForbidden
		return resp
	}

	filter := permission.FilterPrivilegeGroupPrivilege{
		PaginateIn: dto.PaginateIn{
			Limit: paramIn.Limit,
			Page:  paramIn.Page,
		},
		PrivilegeGroup: paramIn.PrivilegeGroup,
	}

	privilegeGroups, err := permissionRepo.FilterPrivilegeGroupPrivilege(filter)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterPrivilege(): failed to filter privilege")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	count, err := permissionRepo.FilterPrivilegeGroupPrivilegeCount(filter)
	if err != nil {
		i.Logger.ErrorErr(paramIn.Trace, err).Msg("FilterPrivilege(): failed to count filter")
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = http.StatusInternalServerError
		return resp
	}

	resp.Success = true
	resp.List = privilegeGroups
	resp.TotalRows = count
	return resp
}
