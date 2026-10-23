package api

import "template-go/api"

type CreatePrivilegeGroupReq struct {
	PrivilegeGroup string `json:"privilege_group"`
	Description    string `json:"description"`
}
type CreatePrivilegeGroupResp struct {
}

type AssignPrivilegeToGroupReq struct {
	PrivilegeGroup string `json:"privilege_group"`
	PrivilegeCode  string `json:"privilege_code"`
}
type AssignPrivilegeToGroupResp struct {
}

type AssignAdminUserToPrivilegeGroupReq struct {
	TargetUserID      int    `json:"target_user_id"`
	NewPrivilegeGroup string `json:"new_privilege_group"`
}
type AssignAdminUserToPrivilegeGroupResp struct{}

type FilterAdminGroupReq struct {
	api.PaginateRequest
}

type FilterAdminGroupResp struct {
	PrivilegeGroups []*PrivilegeGroup `json:"privilege_groups"`
	api.PaginationMeta
}

type FilterPrivilegeGroupPrivilegesReq struct {
	api.PaginateRequest
	PrivilegeGroup string `json:"privilege_group"`
}

type FilterPrivilegeGroupPrivilegesResp struct {
	Privileges []*PrivilegeGroupPrivilege `json:"privilege"`
	api.PaginationMeta
}
