package service_test

import (
	"context"
	"net/http"
	"testing"

	"template-go/data/model"
	permissionConst "template-go/modules/permission/permission"
	permissionSvc "template-go/modules/permission/svc"
)

func TestPermissionAdminService_ChangeAdminPrivilegeRole(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		group   string
	}

	type testCase struct {
		name                string
		adminPrivilegeGroup string
		newPrivilegeGroup   string
		expected            expected
	}

	// Seed data: privilege PVChangeAdminPrivilegeRole granted to "Root",
	// plus an existing "Operator" privilege group.
	//
	// +--------+-----------------------------+
	// | Group  | Privilege                   |
	// +--------+-----------------------------+
	// | Root   | PVChangeAdminPrivilegeRole  |
	// +--------+-----------------------------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedGroupPrivilege(t, f.db, "Root", permissionConst.PVChangeAdminPrivilegeRole)
		seedPrivilegeGroup(t, f.db, "Operator")
	}

	// Test cases:
	//
	// +--------------------------------------+---------------+------------------+-------------------+----------------+
	// | Scenario                             | Admin Group   | New Group        | Expected Code     | Assigned Group |
	// +--------------------------------------+---------------+------------------+-------------------+----------------+
	// | assigns an existing privilege group  | Root          | Operator         | 0 (success)        | Operator       |
	// | rejects callers without privilege    | NoPrivilege   | Operator         | 403 Forbidden      | -              |
	// | rejects an unknown target group      | Root          | Missing          | 400 Bad Request    | -              |
	// +--------------------------------------+---------------+------------------+-------------------+----------------+
	tests := []testCase{
		{
			name: "assigns an existing privilege group", adminPrivilegeGroup: "Root", newPrivilegeGroup: "Operator",
			expected: expected{success: true, group: "Operator"},
		},
		{
			name: "rejects callers without privilege", adminPrivilegeGroup: "NoPrivilege", newPrivilegeGroup: "Operator",
			expected: expected{code: http.StatusForbidden},
		},
		{
			name: "rejects an unknown target group", adminPrivilegeGroup: "Root", newPrivilegeGroup: "Missing",
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.ChangeAdminPrivilegeRole(context.Background(), &permissionSvc.ChangeAdminPrivilegeRoleIn{
				Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: tt.adminPrivilegeGroup,
				TargetUserID: 7, NewPrivilegeGroup: tt.newPrivilegeGroup,
			})
			assertEqual(t, "ChangeAdminPrivilegeRole() success", tt.expected.success, got.Success)
			assertEqual(t, "ChangeAdminPrivilegeRole() error code", tt.expected.code, got.ErrorCode)
			var mapping model.UserPrivilegeGroup
			err := f.db.Where("user_id = ?", 7).First(&mapping).Error
			if tt.expected.success {
				if err != nil {
					t.Fatal(err)
				}
				if mapping.PrivilegeGroupCode != tt.expected.group {
					t.Fatalf("assigned group = %q, want %q", mapping.PrivilegeGroupCode, tt.expected.group)
				}
			} else if err == nil {
				t.Fatal("failed operation unexpectedly persisted a user privilege group")
			}
		})
	}
}

func TestPermissionAdminService_InsertNewPrivilegeGroup(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		groups  int64
	}

	type testCase struct {
		name                string
		adminPrivilegeGroup string
		privilegeGroup      string
		expected            expected
	}

	// Seed data: privilege PVInsertPrivilegeGroup granted to "Root".
	//
	// +--------+---------------------------+
	// | Group  | Privilege                 |
	// +--------+---------------------------+
	// | Root   | PVInsertPrivilegeGroup    |
	// +--------+---------------------------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedGroupPrivilege(t, f.db, "Root", permissionConst.PVInsertPrivilegeGroup)
	}

	// Test cases:
	//
	// +----------------------------------+---------------+---------------+-------------------+--------------+
	// | Scenario                         | Admin Group   | New Group     | Expected Code     | Groups Count |
	// +----------------------------------+---------------+---------------+-------------------+--------------+
	// | inserts group when authorized    | Root          | Operator      | 0 (success)        | 1            |
	// | rejects caller without privilege | NoPrivilege   | Operator      | 403 Forbidden      | 0            |
	// | validates required group         | Root          | (empty)       | 400 Bad Request    | 0            |
	// +----------------------------------+---------------+---------------+-------------------+--------------+
	tests := []testCase{
		{
			name: "inserts group when authorized", adminPrivilegeGroup: "Root", privilegeGroup: "Operator",
			expected: expected{success: true, groups: 1},
		},
		{
			name: "rejects caller without privilege", adminPrivilegeGroup: "NoPrivilege", privilegeGroup: "Operator",
			expected: expected{code: http.StatusForbidden},
		},
		{
			name: "validates required group", adminPrivilegeGroup: "Root", privilegeGroup: "",
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.InsertNewPrivilegeGroup(context.Background(), &permissionSvc.InsertNewPrivilegeGroupIn{
				Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: tt.adminPrivilegeGroup,
				PrivilegeGroup: tt.privilegeGroup, Description: "Operations team",
			})
			assertEqual(t, "InsertNewPrivilegeGroup() success", tt.expected.success, got.Success)
			assertEqual(t, "InsertNewPrivilegeGroup() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.PrivilegeGroup{}).Where("code = ?", "Operator").Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.groups {
				t.Fatalf("persisted group count = %d, want %d", count, tt.expected.groups)
			}
		})
	}
}

func TestPermissionAdminService_AssignPrivilegeToGroup(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		links   int64
	}

	type testCase struct {
		name                 string
		adminPrivilegeGroup  string
		targetPrivilegeGroup string
		expected             expected
	}

	// Seed data: privilege PVAssignPrivilegeToGroup granted to "Root" (this
	// grant itself is a PrivilegeGroupPrivilege row, hence the baseline of
	// 1 link), plus an existing "Operator" privilege group.
	//
	// +--------+---------------------------+
	// | Group  | Privilege                 |
	// +--------+---------------------------+
	// | Root   | PVAssignPrivilegeToGroup  |
	// +--------+---------------------------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedGroupPrivilege(t, f.db, "Root", permissionConst.PVAssignPrivilegeToGroup)
		seedPrivilegeGroup(t, f.db, "Operator")
	}

	// Test cases:
	//
	// +-------------------------------------+---------------+-----------------------+---------------------+-------+
	// | Scenario                            | Admin Group   | Target Group          | Expected Code       | Links |
	// +-------------------------------------+---------------+-----------------------+---------------------+-------+
	// | assigns privilege to target group   | Root          | Operator              | 0 (success)         | 2     |
	// | rejects caller without privilege    | NoPrivilege   | Operator              | 401 Unauthorized    | 1     |
	// | rejects missing target group        | Root          | Missing               | 400 Bad Request     | 1     |
	// +-------------------------------------+---------------+-----------------------+---------------------+-------+
	tests := []testCase{
		{
			name: "assigns privilege to target group", adminPrivilegeGroup: "Root", targetPrivilegeGroup: "Operator",
			expected: expected{success: true, links: 2},
		},
		{
			name: "rejects caller without privilege", adminPrivilegeGroup: "NoPrivilege", targetPrivilegeGroup: "Operator",
			expected: expected{code: http.StatusUnauthorized, links: 1},
		},
		{
			name: "rejects missing target group", adminPrivilegeGroup: "Root", targetPrivilegeGroup: "Missing",
			expected: expected{code: http.StatusBadRequest, links: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.AssignPrivilegeToGroup(context.Background(), &permissionSvc.AssignPrivilegeToGroupIn{
				Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: tt.adminPrivilegeGroup,
				TargetPrivilegeGroup: tt.targetPrivilegeGroup, PrivilegeCode: "ReadReports",
			})
			assertEqual(t, "AssignPrivilegeToGroup() success", tt.expected.success, got.Success)
			assertEqual(t, "AssignPrivilegeToGroup() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.PrivilegeGroupPrivilege{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.links {
				t.Fatalf("persisted group privilege count = %d, want %d", count, tt.expected.links)
			}
		})
	}
}

func TestPermissionAdminService_HasPrivilege(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		allowed bool
		code    int
	}

	type testCase struct {
		name           string
		privilegeGroup string
		privilege      string
		expected       expected
	}

	// Seed data: "Operator" group granted the "ReadReports" privilege.
	//
	// +----------+---------------+
	// | Group    | Privilege     |
	// +----------+---------------+
	// | Operator | ReadReports   |
	// +----------+---------------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedGroupPrivilege(t, f.db, "Operator", "ReadReports")
	}

	// Test cases:
	//
	// +---------------------------------------+---------------+----------------+-------------------+---------+
	// | Scenario                              | Group          | Privilege      | Expected Code      | Allowed |
	// +---------------------------------------+---------------+----------------+-------------------+---------+
	// | returns true for assigned privilege   | Operator       | ReadReports    | 0 (success)        | true    |
	// | returns false for unassigned privilege| Operator       | DeleteReports  | 0 (success)        | false   |
	// | rejects empty privilege group         | (empty)        | ReadReports    | 400 Bad Request    | false   |
	// +---------------------------------------+---------------+----------------+-------------------+---------+
	tests := []testCase{
		{
			name: "returns true for assigned privilege", privilegeGroup: "Operator", privilege: "ReadReports",
			expected: expected{success: true, allowed: true},
		},
		{
			name: "returns false for unassigned privilege", privilegeGroup: "Operator", privilege: "DeleteReports",
			expected: expected{success: true},
		},
		{
			name: "rejects empty privilege group", privilegeGroup: "", privilege: "ReadReports",
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.HasPrivilege(context.Background(), &permissionSvc.HasPrivilegeIn{
				Trace: testTrace(), PrivilegeGroup: tt.privilegeGroup, Privilege: tt.privilege,
			})
			assertEqual(t, "HasPrivilege() success", tt.expected.success, got.Success)
			assertEqual(t, "HasPrivilege() allowed", tt.expected.allowed, got.IsAllowed)
			assertEqual(t, "HasPrivilege() error code", tt.expected.code, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_GetAdminGroupPrivilege(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		group   string
	}

	type testCase struct {
		name        string
		userAdminId int
		expected    expected
	}

	// Seed data: user 9 is mapped to the "Root" privilege group.
	//
	// +----------+-------+
	// | UserID   | Group |
	// +----------+-------+
	// | 9        | Root  |
	// +----------+-------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedUserPrivilegeGroup(t, f.db, 9, "Root")
	}

	// Test cases:
	//
	// +----------------------------------------------+-------------+-------------------+-------+
	// | Scenario                                      | UserAdminId | Expected Code     | Group |
	// +----------------------------------------------+-------------+-------------------+-------+
	// | returns mapped admin group                    | 9           | 0 (success)       | Root  |
	// | returns empty group when user has no mapping  | 10          | 0 (success)       | ""    |
	// | rejects zero user ID                          | 0           | 400 Bad Request   | ""    |
	// +----------------------------------------------+-------------+-------------------+-------+
	tests := []testCase{
		{name: "returns mapped admin group", userAdminId: 9, expected: expected{success: true, group: "Root"}},
		{name: "returns empty group when user has no mapping", userAdminId: 10, expected: expected{success: true}},
		{name: "rejects zero user ID", userAdminId: 0, expected: expected{code: http.StatusBadRequest}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.GetAdminGroupPrivilege(context.Background(), &permissionSvc.GetAdminGroupPrivilegeIn{
				Trace: testTrace(), UserAdminId: tt.userAdminId,
			})
			assertEqual(t, "GetAdminGroupPrivilege() success", tt.expected.success, got.Success)
			assertEqual(t, "GetAdminGroupPrivilege() group", tt.expected.group, got.AdminPrivilegeGroup)
			assertEqual(t, "GetAdminGroupPrivilege() error code", tt.expected.code, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_CheckAdminGroupPrivilegeExists(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		exists  bool
		code    int
	}

	type testCase struct {
		name                string
		adminPrivilegeGroup string
		expected            expected
	}

	// Seed data: "Root" and "Missing" are both granted the
	// PVCheckAdminGroupPrivilegeExists privilege, but only "Root" exists as
	// an actual privilege group row.
	//
	// +---------+--------------------------------+----------------+
	// | Group   | Privilege                      | Group Exists   |
	// +---------+--------------------------------+----------------+
	// | Root    | PVCheckAdminGroupPrivilegeExists | yes          |
	// | Missing | PVCheckAdminGroupPrivilegeExists | no           |
	// +---------+--------------------------------+----------------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedGroupPrivilege(t, f.db, "Root", permissionConst.PVCheckAdminGroupPrivilegeExists)
		seedPrivilegeGroup(t, f.db, "Root")
		seedGroupPrivilege(t, f.db, "Missing", permissionConst.PVCheckAdminGroupPrivilegeExists)
	}

	// Test cases:
	//
	// +---------------------------------------+---------------+-------------------+--------+
	// | Scenario                               | Admin Group    | Expected Code     | Exists |
	// +---------------------------------------+---------------+-------------------+--------+
	// | returns true for existing group        | Root           | 0 (success)       | true   |
	// | returns false when group is missing    | Missing        | 0 (success)       | false  |
	// | rejects caller without privilege       | NoPrivilege    | 403 Forbidden     | false  |
	// +---------------------------------------+---------------+-------------------+--------+
	tests := []testCase{
		{name: "returns true for existing group", adminPrivilegeGroup: "Root", expected: expected{success: true, exists: true}},
		{name: "returns false when group is missing", adminPrivilegeGroup: "Missing", expected: expected{success: true}},
		{name: "rejects caller without privilege", adminPrivilegeGroup: "NoPrivilege", expected: expected{code: http.StatusForbidden}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.CheckAdminGroupPrivilegeExists(context.Background(), &permissionSvc.CheckAdminGroupPrivilegeExistsIn{
				Trace: testTrace(), AdminPrivilegeGroup: tt.adminPrivilegeGroup,
			})
			assertEqual(t, "CheckAdminGroupPrivilegeExists() success", tt.expected.success, got.Success)
			assertEqual(t, "CheckAdminGroupPrivilegeExists() exists", tt.expected.exists, got.IsExists)
			assertEqual(t, "CheckAdminGroupPrivilegeExists() error code", tt.expected.code, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroup(t *testing.T) {
	t.Parallel()

	type expected struct {
		success   bool
		code      int
		groups    int
		totalRows int
	}

	type testCase struct {
		name                string
		adminPrivilegeGroup string
		page                int
		limit               int
		expected            expected
	}

	// Seed data: privilege PVFilterPrivilegeGroup granted to "Root", plus
	// two existing privilege groups.
	//
	// +----------+
	// | Group    |
	// +----------+
	// | Operator |
	// | Auditor  |
	// +----------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedGroupPrivilege(t, f.db, "Root", permissionConst.PVFilterPrivilegeGroup)
		seedPrivilegeGroup(t, f.db, "Operator")
		seedPrivilegeGroup(t, f.db, "Auditor")
	}

	// Test cases:
	//
	// +---------------------------------------+---------------+------+-------+-------------------+--------+-----------+
	// | Scenario                               | Admin Group    | Page | Limit | Expected Code     | Groups | TotalRows |
	// +---------------------------------------+---------------+------+-------+-------------------+--------+-----------+
	// | returns groups with total count        | Root           | 1    | 10    | 0 (success)       | 2      | 2         |
	// | rejects invalid pagination             | Root           | 0    | 10    | 400 Bad Request   | 0      | 0         |
	// | rejects callers without privilege       | NoPrivilege    | 1    | 10    | 403 Forbidden     | 0      | 0         |
	// +---------------------------------------+---------------+------+-------+-------------------+--------+-----------+
	tests := []testCase{
		{
			name: "returns groups with total count", adminPrivilegeGroup: "Root", page: 1, limit: 10,
			expected: expected{success: true, groups: 2, totalRows: 2},
		},
		{
			name: "rejects invalid pagination", adminPrivilegeGroup: "Root", page: 0, limit: 10,
			expected: expected{code: http.StatusBadRequest},
		},
		{
			name: "rejects callers without privilege", adminPrivilegeGroup: "NoPrivilege", page: 1, limit: 10,
			expected: expected{code: http.StatusForbidden},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.FilterPrivilegeGroup(context.Background(), &permissionSvc.FilterPrivilegeGroupIn{
				Trace: testTrace(), AdminPrivilegeGroup: tt.adminPrivilegeGroup, AdminUserId: 1, Page: tt.page, Limit: tt.limit,
			})
			assertEqual(t, "FilterPrivilegeGroup() success", tt.expected.success, got.Success)
			assertEqual(t, "FilterPrivilegeGroup() error code", tt.expected.code, got.ErrorCode)
			assertEqual(t, "FilterPrivilegeGroup() group count", tt.expected.groups, len(got.List))
			assertEqual(t, "FilterPrivilegeGroup() total rows", tt.expected.totalRows, got.TotalRows)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroupPrivileges(t *testing.T) {
	t.Parallel()

	type expected struct {
		success   bool
		code      int
		links     int
		totalRows int
	}

	type testCase struct {
		name                string
		adminPrivilegeGroup string
		limit               int
		expected            expected
	}

	// Seed data: privilege PVFilterPrivilege granted to "Root", plus
	// "Operator" granted two privileges.
	//
	// +----------+---------------+
	// | Group    | Privilege     |
	// +----------+---------------+
	// | Operator | ReadReports   |
	// | Operator | WriteReports  |
	// +----------+---------------+
	seedSharedData := func(t *testing.T, f *permissionFixture) {
		seedGroupPrivilege(t, f.db, "Root", permissionConst.PVFilterPrivilege)
		seedGroupPrivilege(t, f.db, "Operator", "ReadReports")
		seedGroupPrivilege(t, f.db, "Operator", "WriteReports")
	}

	// Test cases:
	//
	// +---------------------------------------+---------------+-------+-------------------+-------+-----------+
	// | Scenario                               | Admin Group    | Limit | Expected Code     | Links | TotalRows |
	// +---------------------------------------+---------------+-------+-------------------+-------+-----------+
	// | returns group privilege links          | Root           | 10    | 0 (success)       | 2     | 2         |
	// | rejects missing privilege group        | NoPrivilege    | 10    | 403 Forbidden     | 0     | 0         |
	// | rejects invalid page size               | Root           | 0     | 400 Bad Request   | 0     | 0         |
	// +---------------------------------------+---------------+-------+-------------------+-------+-----------+
	tests := []testCase{
		{
			name: "returns group privilege links", adminPrivilegeGroup: "Root", limit: 10,
			expected: expected{success: true, links: 2, totalRows: 2},
		},
		{
			name: "rejects missing privilege group", adminPrivilegeGroup: "NoPrivilege", limit: 10,
			expected: expected{code: http.StatusForbidden},
		},
		{
			name: "rejects invalid page size", adminPrivilegeGroup: "Root", limit: 0,
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			seedSharedData(t, f)
			got := f.service.FilterPrivilegeGroupPrivileges(context.Background(), &permissionSvc.FilterPrivilegeGroupPrivilegesIn{
				Trace: testTrace(), AdminPrivilegeGroup: tt.adminPrivilegeGroup, AdminUserId: 1,
				PrivilegeGroup: "Operator", Page: 1, Limit: tt.limit,
			})
			assertEqual(t, "FilterPrivilegeGroupPrivileges() success", tt.expected.success, got.Success)
			assertEqual(t, "FilterPrivilegeGroupPrivileges() error code", tt.expected.code, got.ErrorCode)
			assertEqual(t, "FilterPrivilegeGroupPrivileges() link count", tt.expected.links, len(got.List))
			assertEqual(t, "FilterPrivilegeGroupPrivileges() total rows", tt.expected.totalRows, got.TotalRows)
		})
	}
}
