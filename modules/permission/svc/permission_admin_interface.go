package svc

import (
	"context"
	dto "template-go/base/base"
	"template-go/data/model"
	"template-go/util/logtrace"
)

type PermissionAdminService interface {

	// ChangeAdminPrivilegeRole used to update or change the privilege role assigned to an existing administrator.
	// It modifies the admin's permissions and returns the operation result.
	ChangeAdminPrivilegeRole(ctx context.Context, paramIn *ChangeAdminPrivilegeRoleIn) *ChangeAdminPrivilegeRoleOut

	// InsertNewPrivilegeGroup used to create and insert a new privilege group into the system.
	// It defines a new grouping of permissions and returns the creation result.
	InsertNewPrivilegeGroup(ctx context.Context, paramIn *InsertNewPrivilegeGroupIn) *InsertNewPrivilegeGroupOut

	// AssignPrivilegeToGroup used to assign specific privileges to a given privilege group.
	// It associates the privileges with the group and returns the assignment result.
	AssignPrivilegeToGroup(ctx context.Context, paramIn *AssignPrivilegeToGroupIn) *AssignPrivilegeToGroupOut

	// HasPrivilege checks whether a privilege group possesses a specific privilege.
	HasPrivilege(ctx context.Context, paramIn *HasPrivilegeIn) *HasPrivilegeOut

	// GetAdminGroupPrivilege retrieves the privilege group assigned to a specific admin user.
	GetAdminGroupPrivilege(ctx context.Context, paramIn *GetAdminGroupPrivilegeIn) *GetAdminGroupPrivilegeOut

	// CheckAdminGroupPrivilegeExists verifies whether a given privilege group exists in the system.
	CheckAdminGroupPrivilegeExists(ctx context.Context, paramIn *CheckAdminGroupPrivilegeExistsIn) *CheckAdminGroupPrivilegeExistsOut

	// FilterPrivilegeGroup filter admin's privilege groups.
	FilterPrivilegeGroup(ctx context.Context, paramIn *FilterPrivilegeGroupIn) *FilterPrivilegeGroupOut

	// FilterPrivilegeGroupPrivileges filter privilege group's privilege list.
	FilterPrivilegeGroupPrivileges(ctx context.Context, paramIn *FilterPrivilegeGroupPrivilegesIn) *FilterPrivilegeGroupPrivilegesOut
}

type ChangeAdminPrivilegeRoleIn struct {
	Trace               *logtrace.LogTrace
	UserAdminId         int
	AdminPrivilegeGroup string
	TargetUserID        int
	NewPrivilegeGroup   string
}
type ChangeAdminPrivilegeRoleOut struct {
	dto.BaseOut
}
type InsertNewPrivilegeGroupIn struct {
	Trace               *logtrace.LogTrace
	UserAdminId         int
	AdminPrivilegeGroup string
	PrivilegeGroup      string
	Description         string
}
type InsertNewPrivilegeGroupOut struct {
	dto.BaseOut
}
type AssignPrivilegeToGroupIn struct {
	Trace               *logtrace.LogTrace
	UserAdminId         int
	AdminPrivilegeGroup string

	TargetPrivilegeGroup string
	PrivilegeCode        string
}

type AssignPrivilegeToGroupOut struct {
	dto.BaseOut
}

type HasPrivilegeIn struct {
	Trace          *logtrace.LogTrace
	PrivilegeGroup string
	Privilege      string
}

type HasPrivilegeOut struct {
	dto.BaseOut
	IsAllowed bool
}

type GetAdminGroupPrivilegeIn struct {
	Trace       *logtrace.LogTrace
	UserAdminId int
}

type GetAdminGroupPrivilegeOut struct {
	dto.BaseOut
	AdminPrivilegeGroup string
}

type CheckAdminGroupPrivilegeExistsIn struct {
	Trace               *logtrace.LogTrace
	AdminPrivilegeGroup string
}

type CheckAdminGroupPrivilegeExistsOut struct {
	dto.BaseOut
	IsExists bool
}

type FilterPrivilegeGroupIn struct {
	Trace               *logtrace.LogTrace
	AdminPrivilegeGroup string
	AdminUserId         int
	Page                int
	Limit               int
}

type FilterPrivilegeGroupOut struct {
	dto.BaseOut
	List      []*model.PrivilegeGroup
	TotalRows int
}
type FilterPrivilegeGroupPrivilegesIn struct {
	Trace               *logtrace.LogTrace
	AdminPrivilegeGroup string
	AdminUserId         int
	PrivilegeGroup      string
	Page                int
	Limit               int
}

type FilterPrivilegeGroupPrivilegesOut struct {
	dto.BaseOut
	List      []*model.PrivilegeGroupPrivilege
	TotalRows int
}
