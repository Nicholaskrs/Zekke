package fixture

import (
	"testing"

	"gorm.io/gorm"

	"template-go/data/model"
	permissionRepo "template-go/modules/permission/repository"
	permissionSvc "template-go/modules/permission/svc"
)

type PermissionFixture struct {
	t       *testing.T
	db      *gorm.DB
	Service permissionSvc.PermissionAdminService
}

func NewPermissionFixture(t *testing.T) *PermissionFixture {
	t.Helper()
	db, _ := NewTestDatabase(t)
	storage := permissionRepo.NewPermissionStorage(db)
	return &PermissionFixture{
		t:       t,
		db:      db,
		Service: permissionSvc.NewPermissionAdminService(*storage),
	}
}

func (p *PermissionFixture) SeedPrivilegeGroup(t *testing.T, code, description string) *model.PrivilegeGroup {
	t.Helper()
	group := &model.PrivilegeGroup{Code: code, Description: description}
	mustCreate(t, p.db, group, "privilege group")
	return group
}

func (p *PermissionFixture) SeedGroupPrivilege(t *testing.T, group, privilege string) *model.PrivilegeGroupPrivilege {
	t.Helper()
	groupPrivilege := &model.PrivilegeGroupPrivilege{PrivilegeGroupCode: group, Privilege: privilege}
	mustCreate(t, p.db, groupPrivilege, "privilege group privilege")
	return groupPrivilege
}

func (p *PermissionFixture) SeedUserPrivilegeGroup(t *testing.T, userID uint, group string) *model.UserPrivilegeGroup {
	t.Helper()
	userPrivilegeGroup := &model.UserPrivilegeGroup{UserID: userID, PrivilegeGroupCode: group}
	mustCreate(t, p.db, userPrivilegeGroup, "user privilege group")
	return userPrivilegeGroup
}

func (p *PermissionFixture) FindUserPrivilegeGroup(userID uint) *model.UserPrivilegeGroup {
	p.t.Helper()
	var group model.UserPrivilegeGroup
	result := p.db.Where("user_id = ?", userID).Find(&group)
	if result.Error != nil {
		p.t.Fatal(result.Error)
		return nil
	}
	if result.RowsAffected == 0 {
		return nil
	}
	return &group
}

func (p *PermissionFixture) FindGroupPrivileges(group string) []*model.PrivilegeGroupPrivilege {
	p.t.Helper()
	var privileges []*model.PrivilegeGroupPrivilege
	if err := p.db.Where("privilege_group_code = ?", group).Find(&privileges).Error; err != nil {
		p.t.Fatal(err)
		return nil
	}
	return privileges
}

func (p *PermissionFixture) CountPrivilegeGroups() int64 {
	p.t.Helper()
	var count int64
	if err := p.db.Model(&model.PrivilegeGroup{}).Count(&count).Error; err != nil {
		p.t.Fatal(err)
	}
	return count
}

func (p *PermissionFixture) CountGroupPrivileges(group string) int64 {
	p.t.Helper()
	var count int64
	if err := p.db.Model(&model.PrivilegeGroupPrivilege{}).Where("privilege_group_code = ?", group).Count(&count).Error; err != nil {
		p.t.Fatal(err)
	}
	return count
}
