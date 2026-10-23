package api

import "template-go/data/model"

type PrivilegeGroup struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func ParsePrivilegeGroups(privilegeGroups []*model.PrivilegeGroup) []*PrivilegeGroup {
	result := make([]*PrivilegeGroup, 0)
	for _, privilegeGroup := range privilegeGroups {
		result = append(result, ParsePrivilegeGroup(privilegeGroup))
	}
	return result
}

func ParsePrivilegeGroup(privilegeGroup *model.PrivilegeGroup) *PrivilegeGroup {
	return &PrivilegeGroup{
		Code:        privilegeGroup.Code,
		Description: privilegeGroup.Description,
	}
}

type PrivilegeGroupPrivilege struct {
	PrivilegeCode string `json:"privilege_code"`
}

func ParsePrivilegeGroupPrivilege(privilege *model.PrivilegeGroupPrivilege) *PrivilegeGroupPrivilege {
	return &PrivilegeGroupPrivilege{
		PrivilegeCode: privilege.Privilege,
	}
}

func ParsePrivilegeGroupPrivileges(privileges []*model.PrivilegeGroupPrivilege) []*PrivilegeGroupPrivilege {
	result := make([]*PrivilegeGroupPrivilege, 0)
	for _, privilege := range privileges {
		result = append(result, ParsePrivilegeGroupPrivilege(privilege))
	}
	return result
}
