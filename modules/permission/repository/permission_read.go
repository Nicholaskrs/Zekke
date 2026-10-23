package permission

import (
	"template-go/base/helpers"
	"template-go/data/model"

	"gorm.io/gorm"
)

type PermissionRepositoryRead struct {
	transaction *gorm.DB
}

// GetPrivilegeGroupPrivileges returns every role and its associated privileges.
func (r *PermissionRepositoryRead) GetPrivilegeGroupPrivileges() (
	map[string]map[string]bool,
	error,
) {

	rows, err := r.transaction.Model(&model.PrivilegeGroupPrivilege{}).
		Select("privilege_group_code, privilege").
		Rows()

	if err != nil {
		return nil, err
	}

	privilegesByRole := make(map[string]map[string]bool)

	for rows.Next() {
		var privilegeGroupCode, privilegeCode string
		if err := rows.Scan(&privilegeGroupCode, &privilegeCode); err != nil {
			return nil, err
		}
		if privilegesByRole[privilegeGroupCode] == nil {
			privilegesByRole[privilegeGroupCode] = make(map[string]bool)
		}

		privilegesByRole[privilegeGroupCode][privilegeCode] = true
	}

	return privilegesByRole, nil
}

// GetUserPrivilegeGroup returns given admin user privilege group.
func (r *PermissionRepositoryRead) GetUserPrivilegeGroup(userId int) (
	string,
	error,
) {
	var privilegeGroupCode string
	err := r.transaction.
		Model(&model.UserPrivilegeGroup{}).
		Select("privilege_group_code").
		Where("user_id = ?", userId).
		Scan(&privilegeGroupCode).Error

	if err != nil {
		return "", err
	}

	return privilegeGroupCode, nil
}

// CheckPrivilegeGroupExists returns true if the privilege group exists.
func (r *PermissionRepositoryRead) CheckPrivilegeGroupExists(privilegeGroup string) (
	bool,
	error,
) {
	var count int64
	err := r.transaction.
		Model(&model.PrivilegeGroup{}).
		Where("code = ?", privilegeGroup).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CheckPrivilegeExists returns true if the privilege  exists.
func (r *PermissionRepositoryRead) CheckPrivilegeExists(privilegeGroup string) (
	bool,
	error,
) {
	var count int64
	err := r.transaction.
		Model(&model.Privilege{}).
		Where("privilege = ?", privilegeGroup).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *PermissionRepositoryRead) applyFilterPrivilegeGroup(filter FilterPrivilegeGroup, query *gorm.DB) *gorm.DB {
	if filter.CodeStartWith != "" {
		query = query.Where("code ", "LIKE", filter.CodeStartWith+"%")
	}
	return query

}

// FilterPrivilegeGroupCount return total data of given filter.
func (r *PermissionRepositoryRead) FilterPrivilegeGroupCount(filter FilterPrivilegeGroup) (int, error) {
	var count int64

	query := r.transaction.Model(model.PrivilegeGroup{})

	// Where Condition.
	query = r.applyFilterPrivilegeGroup(filter, query)

	// Find model
	if err := query.
		Count(&count).Error; err != nil {
		return 0, helpers.Wrap(err)
	}
	return int(count), nil
}

// FilterPrivilegeGroup retrieves privilegeGroup based on given filter. It return array privilege and error.
func (r *PermissionRepositoryRead) FilterPrivilegeGroup(filter FilterPrivilegeGroup) ([]*model.PrivilegeGroup, error) {
	var privilegeGroups []*model.PrivilegeGroup

	query := r.transaction

	// Where Condition.
	query = r.applyFilterPrivilegeGroup(filter, query)

	// Find model
	if err := query.
		Offset(filter.GetOffset()).
		Limit(filter.Limit).
		Find(&privilegeGroups).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return privilegeGroups, nil
}

func (r *PermissionRepositoryRead) applyFilterPrivilegeGroupPrivilege(filter FilterPrivilegeGroupPrivilege, query *gorm.DB) *gorm.DB {
	if filter.PrivilegeGroup != "" {
		query = query.Where("privilege_group_code = ?", filter.PrivilegeGroup)
	}
	return query
}

func (r *PermissionRepositoryRead) FilterPrivilegeGroupPrivilege(filter FilterPrivilegeGroupPrivilege) ([]*model.PrivilegeGroupPrivilege, error) {
	var privileges []*model.PrivilegeGroupPrivilege

	query := r.transaction

	// Where Condition.
	query = r.applyFilterPrivilegeGroupPrivilege(filter, query)

	// Find model
	if err := query.
		Offset(filter.GetOffset()).
		Limit(filter.Limit).
		Find(&privileges).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return privileges, nil
}

func (r *PermissionRepositoryRead) FilterPrivilegeGroupPrivilegeCount(filter FilterPrivilegeGroupPrivilege) (int, error) {
	var count int64

	query := r.transaction.Model(model.PrivilegeGroupPrivilege{})

	// Where Condition.
	query = r.applyFilterPrivilegeGroupPrivilege(filter, query)

	// Find model
	if err := query.
		Count(&count).Error; err != nil {
		return 0, helpers.Wrap(err)
	}
	return int(count), nil
}
