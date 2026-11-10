package service_test

import (
	"context"
	"net/http"
	"testing"

	permissionConst "template-go/modules/permission/permission"
	permissionSvc "template-go/modules/permission/svc"
	"template-go/test"
	"template-go/test/fixture"
)

func seedPermissionAdminFixture(t *testing.T, f *fixture.PermissionFixture, privileges ...string) {
	t.Helper()
	f.SeedPrivilegeGroup(t, "SuperAdmin", "super administrator")
	for _, privilege := range privileges {
		f.SeedGroupPrivilege(t, "SuperAdmin", privilege)
	}
}

func TestPermissionAdminService_ChangeAdminPrivilegeRole(t *testing.T) {
	t.Parallel()

	type input struct {
		UserAdminId         int
		AdminPrivilegeGroup string
		TargetUserID        int
		NewPrivilegeGroup   string
	}
	type expected struct {
		success   bool
		errorCode int
		groupCode string
	}
	type testCase struct {
		name     string
		input    input
		expected expected
		allow    bool
	}

	tests := []testCase{
		{
			name:     "assigns requested group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetUserID: 20, NewPrivilegeGroup: "Auditor"},
			expected: expected{success: true, groupCode: "Auditor"},
			allow:    true,
		},
		{
			name:     "rejects empty target user ID",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", NewPrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty admin user ID",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", TargetUserID: 20, NewPrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty admin privilege group",
			input:    input{UserAdminId: 10, TargetUserID: 20, NewPrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty new privilege group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetUserID: 20},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects unknown target group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetUserID: 20, NewPrivilegeGroup: "Missing"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects caller without privilege",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetUserID: 20, NewPrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusForbidden},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			if tt.allow {
				seedPermissionAdminFixture(t, f, permissionConst.PVChangeAdminPrivilegeRole)
			} else {
				seedPermissionAdminFixture(t, f)
			}
			f.SeedPrivilegeGroup(t, "Auditor", "auditor")

			got := f.Service.ChangeAdminPrivilegeRole(context.Background(), &permissionSvc.ChangeAdminPrivilegeRoleIn{
				Trace: fixture.TestTrace(), UserAdminId: tt.input.UserAdminId,
				AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup, TargetUserID: tt.input.TargetUserID,
				NewPrivilegeGroup: tt.input.NewPrivilegeGroup,
			})
			test.AssertEqual(t, "ChangeAdminPrivilegeRole success", tt.expected.success, got.Success)
			test.AssertEqual(t, "ChangeAdminPrivilegeRole error code", tt.expected.errorCode, got.ErrorCode)
			group := f.FindUserPrivilegeGroup(uint(tt.input.TargetUserID))
			if tt.expected.groupCode == "" {
				test.AssertNil(t, "user privilege group", group)
			} else {
				test.AssertEqual(t, "assigned privilege group", tt.expected.groupCode, group.PrivilegeGroupCode)
			}
		})
	}
}

func TestPermissionAdminService_InsertNewPrivilegeGroup(t *testing.T) {
	t.Parallel()

	type input struct {
		UserAdminId         int
		AdminPrivilegeGroup string
		PrivilegeGroup      string
		Description         string
	}
	type expected struct {
		success   bool
		errorCode int
		groupRows int64
	}
	type testCase struct {
		name     string
		input    input
		expected expected
		allow    bool
	}

	tests := []testCase{
		{
			name:     "inserts privilege group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", PrivilegeGroup: "Auditor", Description: "audit permissions"},
			expected: expected{success: true, groupRows: 2},
			allow:    true,
		},
		{
			name:     "rejects empty admin user ID",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", PrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusBadRequest, groupRows: 1},
			allow:    true,
		},
		{
			name:     "rejects empty admin privilege group",
			input:    input{UserAdminId: 10, PrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusBadRequest, groupRows: 1},
			allow:    true,
		},
		{
			name:     "rejects empty new privilege group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin"},
			expected: expected{errorCode: http.StatusBadRequest, groupRows: 1},
			allow:    true,
		},
		{
			name:     "rejects caller without privilege",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", PrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusForbidden, groupRows: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			if tt.allow {
				seedPermissionAdminFixture(t, f, permissionConst.PVInsertPrivilegeGroup)
			} else {
				seedPermissionAdminFixture(t, f)
			}

			got := f.Service.InsertNewPrivilegeGroup(context.Background(), &permissionSvc.InsertNewPrivilegeGroupIn{
				Trace: fixture.TestTrace(), UserAdminId: tt.input.UserAdminId,
				AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup,
				PrivilegeGroup:      tt.input.PrivilegeGroup, Description: tt.input.Description,
			})
			test.AssertEqual(t, "InsertNewPrivilegeGroup success", tt.expected.success, got.Success)
			test.AssertEqual(t, "InsertNewPrivilegeGroup error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "privilege group count", tt.expected.groupRows, f.CountPrivilegeGroups())
		})
	}
}

func TestPermissionAdminService_AssignPrivilegeToGroup(t *testing.T) {
	t.Parallel()

	type input struct {
		UserAdminId          int
		AdminPrivilegeGroup  string
		TargetPrivilegeGroup string
		PrivilegeCode        string
	}
	type expected struct {
		success   bool
		errorCode int
		rows      int64
	}
	type testCase struct {
		name     string
		input    input
		expected expected
		allow    bool
	}

	tests := []testCase{
		{
			name:     "assigns privilege to group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetPrivilegeGroup: "Auditor", PrivilegeCode: "ReadUsers"},
			expected: expected{success: true, rows: 1},
			allow:    true,
		},
		{
			name:     "rejects unknown target group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetPrivilegeGroup: "Missing", PrivilegeCode: "ReadUsers"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty admin privilege group",
			input:    input{UserAdminId: 10, TargetPrivilegeGroup: "Auditor", PrivilegeCode: "ReadUsers"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty admin user ID",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", TargetPrivilegeGroup: "Auditor", PrivilegeCode: "ReadUsers"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty privilege code",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetPrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty target privilege group",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", PrivilegeCode: "ReadUsers"},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects caller without privilege",
			input:    input{UserAdminId: 10, AdminPrivilegeGroup: "SuperAdmin", TargetPrivilegeGroup: "Auditor", PrivilegeCode: "ReadUsers"},
			expected: expected{errorCode: http.StatusUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			if tt.allow {
				seedPermissionAdminFixture(t, f, permissionConst.PVAssignPrivilegeToGroup)
			} else {
				seedPermissionAdminFixture(t, f)
			}
			f.SeedPrivilegeGroup(t, "Auditor", "auditor")

			got := f.Service.AssignPrivilegeToGroup(context.Background(), &permissionSvc.AssignPrivilegeToGroupIn{
				Trace: fixture.TestTrace(), UserAdminId: tt.input.UserAdminId,
				AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup, TargetPrivilegeGroup: tt.input.TargetPrivilegeGroup,
				PrivilegeCode: tt.input.PrivilegeCode,
			})
			test.AssertEqual(t, "AssignPrivilegeToGroup success", tt.expected.success, got.Success)
			test.AssertEqual(t, "AssignPrivilegeToGroup error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "assigned privilege count", tt.expected.rows, f.CountGroupPrivileges("Auditor"))
			if tt.expected.success {
				privileges := f.FindGroupPrivileges("Auditor")
				test.AssertEqual(t, "assigned privilege", tt.input.PrivilegeCode, privileges[0].Privilege)
			}
		})
	}
}

func TestPermissionAdminService_HasPrivilege(t *testing.T) {
	t.Parallel()

	type input struct {
		PrivilegeGroup string
		Privilege      string
	}
	type expected struct {
		success   bool
		errorCode int
		allowed   bool
	}
	type testCase struct {
		name     string
		input    input
		expected expected
		seed     bool
	}

	tests := []testCase{
		{
			name:     "returns true when group has privilege",
			input:    input{PrivilegeGroup: "SuperAdmin", Privilege: "ReadUsers"},
			expected: expected{success: true, allowed: true},
			seed:     true,
		},
		{
			name:     "returns false when group lacks privilege",
			input:    input{PrivilegeGroup: "SuperAdmin", Privilege: "ReadUsers"},
			expected: expected{success: true},
		},
		{
			name:     "rejects empty privilege group",
			input:    input{Privilege: "ReadUsers"},
			expected: expected{errorCode: http.StatusBadRequest},
		},
		{
			name:     "rejects empty privilege",
			input:    input{PrivilegeGroup: "SuperAdmin"},
			expected: expected{errorCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			f.SeedPrivilegeGroup(t, "SuperAdmin", "super administrator")
			if tt.seed {
				f.SeedGroupPrivilege(t, tt.input.PrivilegeGroup, tt.input.Privilege)
			}

			got := f.Service.HasPrivilege(context.Background(), &permissionSvc.HasPrivilegeIn{
				Trace: fixture.TestTrace(), PrivilegeGroup: tt.input.PrivilegeGroup, Privilege: tt.input.Privilege,
			})
			test.AssertEqual(t, "HasPrivilege success", tt.expected.success, got.Success)
			test.AssertEqual(t, "HasPrivilege error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "HasPrivilege allowed", tt.expected.allowed, got.IsAllowed)
		})
	}
}

func TestPermissionAdminService_GetAdminGroupPrivilege(t *testing.T) {
	t.Parallel()

	type input struct {
		UserAdminId int
	}
	type expected struct {
		success   bool
		errorCode int
		group     string
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}

	tests := []testCase{
		{
			name:     "returns assigned group",
			input:    input{UserAdminId: 10},
			expected: expected{success: true, group: "SuperAdmin"},
		},
		{
			name:     "returns empty group for unassigned user",
			input:    input{UserAdminId: 20},
			expected: expected{success: true},
		},
		{
			name:     "rejects empty user ID",
			expected: expected{errorCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			if tt.expected.group != "" {
				f.SeedUserPrivilegeGroup(t, uint(tt.input.UserAdminId), tt.expected.group)
			}
			got := f.Service.GetAdminGroupPrivilege(context.Background(), &permissionSvc.GetAdminGroupPrivilegeIn{
				Trace: fixture.TestTrace(), UserAdminId: tt.input.UserAdminId,
			})
			test.AssertEqual(t, "GetAdminGroupPrivilege success", tt.expected.success, got.Success)
			test.AssertEqual(t, "GetAdminGroupPrivilege error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "GetAdminGroupPrivilege group", tt.expected.group, got.AdminPrivilegeGroup)
		})
	}
}

func TestPermissionAdminService_CheckAdminGroupPrivilegeExists(t *testing.T) {
	t.Parallel()

	type input struct {
		AdminPrivilegeGroup string
	}
	type expected struct {
		success   bool
		errorCode int
		exists    bool
	}
	type testCase struct {
		name        string
		input       input
		expected    expected
		grant       bool
		createGroup bool
	}

	tests := []testCase{
		{
			name:     "returns existing group",
			input:    input{AdminPrivilegeGroup: "SuperAdmin"},
			expected: expected{success: true, exists: true},
			grant:    true, createGroup: true,
		},
		{
			name:     "returns false for nonexistent group",
			input:    input{AdminPrivilegeGroup: "Missing"},
			expected: expected{success: true},
			grant:    true,
		},
		{
			name:        "rejects caller without privilege",
			input:       input{AdminPrivilegeGroup: "SuperAdmin"},
			expected:    expected{errorCode: http.StatusForbidden},
			createGroup: true,
		},
		{
			name:     "rejects empty privilege group",
			expected: expected{errorCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			if tt.createGroup {
				f.SeedPrivilegeGroup(t, "SuperAdmin", "super administrator")
			}
			if tt.grant {
				f.SeedGroupPrivilege(t, tt.input.AdminPrivilegeGroup, permissionConst.PVCheckAdminGroupPrivilegeExists)
			}
			got := f.Service.CheckAdminGroupPrivilegeExists(context.Background(), &permissionSvc.CheckAdminGroupPrivilegeExistsIn{
				Trace: fixture.TestTrace(), AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup,
			})
			test.AssertEqual(t, "CheckAdminGroupPrivilegeExists success", tt.expected.success, got.Success)
			test.AssertEqual(t, "CheckAdminGroupPrivilegeExists error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "CheckAdminGroupPrivilegeExists exists", tt.expected.exists, got.IsExists)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroup(t *testing.T) {
	t.Parallel()

	type input struct {
		AdminPrivilegeGroup string
		AdminUserId         int
		Page                int
		Limit               int
	}
	type expected struct {
		success   bool
		errorCode int
		rows      int
		total     int
	}
	type testCase struct {
		name     string
		input    input
		expected expected
		allow    bool
	}

	tests := []testCase{
		{
			name:     "filters privilege groups",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, Page: 1, Limit: 1},
			expected: expected{success: true, rows: 1, total: 2},
			allow:    true,
		},
		{
			name:     "rejects invalid page",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, Limit: 10},
			expected: expected{errorCode: http.StatusBadRequest, total: 0},
			allow:    true,
		},
		{
			name:     "rejects empty admin privilege group",
			input:    input{AdminUserId: 10, Page: 1, Limit: 10},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty admin user ID",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", Page: 1, Limit: 10},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty limit",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, Page: 1},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects caller without privilege",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, Page: 1, Limit: 10},
			expected: expected{errorCode: http.StatusForbidden},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			f.SeedPrivilegeGroup(t, "SuperAdmin", "super administrator")
			f.SeedPrivilegeGroup(t, "Auditor", "auditor")
			if tt.allow {
				f.SeedGroupPrivilege(t, "SuperAdmin", permissionConst.PVFilterPrivilegeGroup)
			}
			got := f.Service.FilterPrivilegeGroup(context.Background(), &permissionSvc.FilterPrivilegeGroupIn{
				Trace: fixture.TestTrace(), AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup,
				AdminUserId: tt.input.AdminUserId, Page: tt.input.Page, Limit: tt.input.Limit,
			})
			test.AssertEqual(t, "FilterPrivilegeGroup success", tt.expected.success, got.Success)
			test.AssertEqual(t, "FilterPrivilegeGroup error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "FilterPrivilegeGroup rows", tt.expected.rows, len(got.List))
			test.AssertEqual(t, "FilterPrivilegeGroup total", tt.expected.total, got.TotalRows)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroupPrivileges(t *testing.T) {
	t.Parallel()

	type input struct {
		AdminPrivilegeGroup string
		AdminUserId         int
		PrivilegeGroup      string
		Page                int
		Limit               int
	}
	type expected struct {
		success   bool
		errorCode int
		rows      int
		total     int
	}
	type testCase struct {
		name     string
		input    input
		expected expected
		allow    bool
	}

	tests := []testCase{
		{
			name:     "filters group privileges",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, PrivilegeGroup: "Auditor", Page: 1, Limit: 10},
			expected: expected{success: true, rows: 2, total: 2},
			allow:    true,
		},
		{
			name:     "rejects empty privilege group",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, Page: 1, Limit: 10},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty admin privilege group",
			input:    input{AdminUserId: 10, PrivilegeGroup: "Auditor", Page: 1, Limit: 10},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty admin user ID",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", PrivilegeGroup: "Auditor", Page: 1, Limit: 10},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty page",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, PrivilegeGroup: "Auditor", Limit: 10},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects empty limit",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, PrivilegeGroup: "Auditor", Page: 1},
			expected: expected{errorCode: http.StatusBadRequest},
			allow:    true,
		},
		{
			name:     "rejects caller without privilege",
			input:    input{AdminPrivilegeGroup: "SuperAdmin", AdminUserId: 10, PrivilegeGroup: "Auditor", Page: 1, Limit: 10},
			expected: expected{errorCode: http.StatusForbidden},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewPermissionFixture(t)
			f.SeedPrivilegeGroup(t, "SuperAdmin", "super administrator")
			f.SeedPrivilegeGroup(t, "Auditor", "auditor")
			f.SeedGroupPrivilege(t, "Auditor", "ReadUsers")
			f.SeedGroupPrivilege(t, "Auditor", "UpdateUsers")
			if tt.allow {
				f.SeedGroupPrivilege(t, "SuperAdmin", permissionConst.PVFilterPrivilege)
			}
			got := f.Service.FilterPrivilegeGroupPrivileges(context.Background(), &permissionSvc.FilterPrivilegeGroupPrivilegesIn{
				Trace: fixture.TestTrace(), AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup,
				AdminUserId: tt.input.AdminUserId, PrivilegeGroup: tt.input.PrivilegeGroup,
				Page: tt.input.Page, Limit: tt.input.Limit,
			})
			test.AssertEqual(t, "FilterPrivilegeGroupPrivileges success", tt.expected.success, got.Success)
			test.AssertEqual(t, "FilterPrivilegeGroupPrivileges error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "FilterPrivilegeGroupPrivileges rows", tt.expected.rows, len(got.List))
			test.AssertEqual(t, "FilterPrivilegeGroupPrivileges total", tt.expected.total, got.TotalRows)
		})
	}
}
