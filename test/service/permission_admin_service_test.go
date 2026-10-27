package service_test

import (
	"context"
	"net/http"
	"testing"

	"template-go/data/model"
	permissionConst "template-go/modules/permission/permission"
	permissionSvc "template-go/modules/permission/svc"
)

type permissionAdminServiceChangeAdminPrivilegeRoleExpected struct {
	success bool
	code    int
	group   string
}

type permissionAdminServiceChangeAdminPrivilegeRoleTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.ChangeAdminPrivilegeRoleIn
	expected permissionAdminServiceChangeAdminPrivilegeRoleExpected
}

type permissionAdminServiceInsertNewPrivilegeGroupExpected struct {
	success bool
	code    int
	groups  int64
}

type permissionAdminServiceInsertNewPrivilegeGroupTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn
	expected permissionAdminServiceInsertNewPrivilegeGroupExpected
}

type permissionAdminServiceAssignPrivilegeToGroupExpected struct {
	success bool
	code    int
	links   int64
}

type permissionAdminServiceAssignPrivilegeToGroupTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.AssignPrivilegeToGroupIn
	expected permissionAdminServiceAssignPrivilegeToGroupExpected
}

type permissionAdminServiceHasPrivilegeExpected struct {
	success bool
	allowed bool
	code    int
}

type permissionAdminServiceHasPrivilegeTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.HasPrivilegeIn
	expected permissionAdminServiceHasPrivilegeExpected
}

type permissionAdminServiceGetAdminGroupPrivilegeExpected struct {
	success bool
	code    int
	group   string
}

type permissionAdminServiceGetAdminGroupPrivilegeTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn
	expected permissionAdminServiceGetAdminGroupPrivilegeExpected
}

type permissionAdminServiceCheckAdminGroupPrivilegeExistsExpected struct {
	success bool
	exists  bool
	code    int
}

type permissionAdminServiceCheckAdminGroupPrivilegeExistsTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn
	expected permissionAdminServiceCheckAdminGroupPrivilegeExistsExpected
}

type permissionAdminServiceFilterPrivilegeGroupExpected struct {
	success   bool
	code      int
	groups    int
	totalRows int
}

type permissionAdminServiceFilterPrivilegeGroupTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.FilterPrivilegeGroupIn
	expected permissionAdminServiceFilterPrivilegeGroupExpected
}

type permissionAdminServiceFilterPrivilegeGroupPrivilegesExpected struct {
	success   bool
	code      int
	links     int
	totalRows int
}

type permissionAdminServiceFilterPrivilegeGroupPrivilegesTestCase struct {
	name     string
	setup    func(t *testing.T, f *permissionFixture) *permissionSvc.FilterPrivilegeGroupPrivilegesIn
	expected permissionAdminServiceFilterPrivilegeGroupPrivilegesExpected
}

func TestPermissionAdminService_ChangeAdminPrivilegeRole(t *testing.T) {
	t.Parallel()

	tests := []permissionAdminServiceChangeAdminPrivilegeRoleTestCase{
		{
			name: "assigns an existing privilege group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.ChangeAdminPrivilegeRoleIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVChangeAdminPrivilegeRole)
				seedPrivilegeGroup(t, f.db, "Operator")
				return &permissionSvc.ChangeAdminPrivilegeRoleIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetUserID: 7, NewPrivilegeGroup: "Operator",
				}
			},
			expected: permissionAdminServiceChangeAdminPrivilegeRoleExpected{success: true, group: "Operator"},
		},
		{
			name: "rejects callers without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.ChangeAdminPrivilegeRoleIn {
				return &permissionSvc.ChangeAdminPrivilegeRoleIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetUserID: 7, NewPrivilegeGroup: "Operator",
				}
			},
			expected: permissionAdminServiceChangeAdminPrivilegeRoleExpected{code: http.StatusForbidden},
		},
		{
			name: "rejects an unknown target group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.ChangeAdminPrivilegeRoleIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVChangeAdminPrivilegeRole)
				return &permissionSvc.ChangeAdminPrivilegeRoleIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetUserID: 7, NewPrivilegeGroup: "Missing",
				}
			},
			expected: permissionAdminServiceChangeAdminPrivilegeRoleExpected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.ChangeAdminPrivilegeRole(context.Background(), tt.setup(t, f))
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

	tests := []permissionAdminServiceInsertNewPrivilegeGroupTestCase{
		{
			name: "inserts group when authorized",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVInsertPrivilegeGroup)
				return &permissionSvc.InsertNewPrivilegeGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					PrivilegeGroup: "Operator", Description: "Operations team",
				}
			},
			expected: permissionAdminServiceInsertNewPrivilegeGroupExpected{success: true, groups: 1},
		},
		{
			name: "rejects caller without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn {
				return &permissionSvc.InsertNewPrivilegeGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					PrivilegeGroup: "Operator",
				}
			},
			expected: permissionAdminServiceInsertNewPrivilegeGroupExpected{code: http.StatusForbidden},
		},
		{
			name: "validates required group",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn {
				return &permissionSvc.InsertNewPrivilegeGroupIn{Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root"}
			},
			expected: permissionAdminServiceInsertNewPrivilegeGroupExpected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.InsertNewPrivilegeGroup(context.Background(), tt.setup(t, f))
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

	tests := []permissionAdminServiceAssignPrivilegeToGroupTestCase{
		{
			name: "assigns privilege to target group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.AssignPrivilegeToGroupIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVAssignPrivilegeToGroup)
				seedPrivilegeGroup(t, f.db, "Operator")
				return &permissionSvc.AssignPrivilegeToGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetPrivilegeGroup: "Operator", PrivilegeCode: "ReadReports",
				}
			},
			expected: permissionAdminServiceAssignPrivilegeToGroupExpected{success: true, links: 2},
		},
		{
			name: "rejects caller without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.AssignPrivilegeToGroupIn {
				return &permissionSvc.AssignPrivilegeToGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetPrivilegeGroup: "Operator", PrivilegeCode: "ReadReports",
				}
			},
			expected: permissionAdminServiceAssignPrivilegeToGroupExpected{code: http.StatusUnauthorized},
		},
		{
			name: "rejects missing target group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.AssignPrivilegeToGroupIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVAssignPrivilegeToGroup)
				return &permissionSvc.AssignPrivilegeToGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetPrivilegeGroup: "Missing", PrivilegeCode: "ReadReports",
				}
			},
			expected: permissionAdminServiceAssignPrivilegeToGroupExpected{code: http.StatusBadRequest, links: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.AssignPrivilegeToGroup(context.Background(), tt.setup(t, f))
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

	tests := []permissionAdminServiceHasPrivilegeTestCase{
		{
			name: "returns true for assigned privilege",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.HasPrivilegeIn {
				seedGroupPrivilege(t, f.db, "Operator", "ReadReports")
				return &permissionSvc.HasPrivilegeIn{Trace: testTrace(), PrivilegeGroup: "Operator", Privilege: "ReadReports"}
			},
			expected: permissionAdminServiceHasPrivilegeExpected{success: true, allowed: true},
		},
		{
			name: "returns false for unassigned privilege",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.HasPrivilegeIn {
				seedGroupPrivilege(t, f.db, "Operator", "ReadReports")
				return &permissionSvc.HasPrivilegeIn{Trace: testTrace(), PrivilegeGroup: "Operator", Privilege: "DeleteReports"}
			},
			expected: permissionAdminServiceHasPrivilegeExpected{success: true},
		},
		{
			name: "rejects empty privilege group",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.HasPrivilegeIn {
				return &permissionSvc.HasPrivilegeIn{Trace: testTrace(), Privilege: "ReadReports"}
			},
			expected: permissionAdminServiceHasPrivilegeExpected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.HasPrivilege(context.Background(), tt.setup(t, f))
			assertEqual(t, "HasPrivilege() success", tt.expected.success, got.Success)
			assertEqual(t, "HasPrivilege() allowed", tt.expected.allowed, got.IsAllowed)
			assertEqual(t, "HasPrivilege() error code", tt.expected.code, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_GetAdminGroupPrivilege(t *testing.T) {
	t.Parallel()

	tests := []permissionAdminServiceGetAdminGroupPrivilegeTestCase{
		{
			name: "returns mapped admin group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn {
				seedUserPrivilegeGroup(t, f.db, 9, "Root")
				return &permissionSvc.GetAdminGroupPrivilegeIn{Trace: testTrace(), UserAdminId: 9}
			},
			expected: permissionAdminServiceGetAdminGroupPrivilegeExpected{success: true, group: "Root"},
		},
		{
			name: "returns empty group when user has no mapping",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn {
				return &permissionSvc.GetAdminGroupPrivilegeIn{Trace: testTrace(), UserAdminId: 9}
			},
			expected: permissionAdminServiceGetAdminGroupPrivilegeExpected{success: true},
		},
		{
			name: "rejects zero user ID",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn {
				return &permissionSvc.GetAdminGroupPrivilegeIn{Trace: testTrace()}
			},
			expected: permissionAdminServiceGetAdminGroupPrivilegeExpected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.GetAdminGroupPrivilege(context.Background(), tt.setup(t, f))
			assertEqual(t, "GetAdminGroupPrivilege() success", tt.expected.success, got.Success)
			assertEqual(t, "GetAdminGroupPrivilege() group", tt.expected.group, got.AdminPrivilegeGroup)
			assertEqual(t, "GetAdminGroupPrivilege() error code", tt.expected.code, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_CheckAdminGroupPrivilegeExists(t *testing.T) {
	t.Parallel()

	tests := []permissionAdminServiceCheckAdminGroupPrivilegeExistsTestCase{
		{
			name: "returns true for existing group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVCheckAdminGroupPrivilegeExists)
				seedPrivilegeGroup(t, f.db, "Root")
				return &permissionSvc.CheckAdminGroupPrivilegeExistsIn{Trace: testTrace(), AdminPrivilegeGroup: "Root"}
			},
			expected: permissionAdminServiceCheckAdminGroupPrivilegeExistsExpected{success: true, exists: true},
		},
		{
			name: "returns false when group is missing",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn {
				seedGroupPrivilege(t, f.db, "Missing", permissionConst.PVCheckAdminGroupPrivilegeExists)
				return &permissionSvc.CheckAdminGroupPrivilegeExistsIn{Trace: testTrace(), AdminPrivilegeGroup: "Missing"}
			},
			expected: permissionAdminServiceCheckAdminGroupPrivilegeExistsExpected{success: true},
		},
		{
			name: "rejects caller without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn {
				return &permissionSvc.CheckAdminGroupPrivilegeExistsIn{Trace: testTrace(), AdminPrivilegeGroup: "Root"}
			},
			expected: permissionAdminServiceCheckAdminGroupPrivilegeExistsExpected{code: http.StatusForbidden},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.CheckAdminGroupPrivilegeExists(context.Background(), tt.setup(t, f))
			assertEqual(t, "CheckAdminGroupPrivilegeExists() success", tt.expected.success, got.Success)
			assertEqual(t, "CheckAdminGroupPrivilegeExists() exists", tt.expected.exists, got.IsExists)
			assertEqual(t, "CheckAdminGroupPrivilegeExists() error code", tt.expected.code, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroup(t *testing.T) {
	t.Parallel()

	tests := []permissionAdminServiceFilterPrivilegeGroupTestCase{
		{
			name: "returns groups with total count",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.FilterPrivilegeGroupIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVFilterPrivilegeGroup)
				seedPrivilegeGroup(t, f.db, "Operator")
				seedPrivilegeGroup(t, f.db, "Auditor")
				return &permissionSvc.FilterPrivilegeGroupIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1, Page: 1, Limit: 10,
				}
			},
			expected: permissionAdminServiceFilterPrivilegeGroupExpected{success: true, groups: 2, totalRows: 2},
		},
		{
			name: "rejects invalid pagination",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.FilterPrivilegeGroupIn {
				return &permissionSvc.FilterPrivilegeGroupIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1, Page: 0, Limit: 10,
				}
			},
			expected: permissionAdminServiceFilterPrivilegeGroupExpected{code: http.StatusBadRequest},
		},
		{
			name: "rejects callers without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.FilterPrivilegeGroupIn {
				return &permissionSvc.FilterPrivilegeGroupIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1, Page: 1, Limit: 10,
				}
			},
			expected: permissionAdminServiceFilterPrivilegeGroupExpected{code: http.StatusForbidden},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.FilterPrivilegeGroup(context.Background(), tt.setup(t, f))
			assertEqual(t, "FilterPrivilegeGroup() success", tt.expected.success, got.Success)
			assertEqual(t, "FilterPrivilegeGroup() error code", tt.expected.code, got.ErrorCode)
			assertEqual(t, "FilterPrivilegeGroup() group count", tt.expected.groups, len(got.List))
			assertEqual(t, "FilterPrivilegeGroup() total rows", tt.expected.totalRows, got.TotalRows)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroupPrivileges(t *testing.T) {
	t.Parallel()

	tests := []permissionAdminServiceFilterPrivilegeGroupPrivilegesTestCase{
		{
			name: "returns group privilege links",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.FilterPrivilegeGroupPrivilegesIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVFilterPrivilege)
				seedGroupPrivilege(t, f.db, "Operator", "ReadReports")
				seedGroupPrivilege(t, f.db, "Operator", "WriteReports")
				return &permissionSvc.FilterPrivilegeGroupPrivilegesIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1,
					PrivilegeGroup: "Operator", Page: 1, Limit: 10,
				}
			},
			expected: permissionAdminServiceFilterPrivilegeGroupPrivilegesExpected{success: true, links: 2, totalRows: 2},
		},
		{
			name: "rejects missing privilege group",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.FilterPrivilegeGroupPrivilegesIn {
				return &permissionSvc.FilterPrivilegeGroupPrivilegesIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1,
					PrivilegeGroup: "Operator", Page: 1, Limit: 10,
				}
			},
			expected: permissionAdminServiceFilterPrivilegeGroupPrivilegesExpected{code: http.StatusForbidden},
		},
		{
			name: "rejects invalid page size",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.FilterPrivilegeGroupPrivilegesIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVFilterPrivilege)
				return &permissionSvc.FilterPrivilegeGroupPrivilegesIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1,
					PrivilegeGroup: "Operator", Page: 1, Limit: 0,
				}
			},
			expected: permissionAdminServiceFilterPrivilegeGroupPrivilegesExpected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.FilterPrivilegeGroupPrivileges(context.Background(), tt.setup(t, f))
			assertEqual(t, "FilterPrivilegeGroupPrivileges() success", tt.expected.success, got.Success)
			assertEqual(t, "FilterPrivilegeGroupPrivileges() error code", tt.expected.code, got.ErrorCode)
			assertEqual(t, "FilterPrivilegeGroupPrivileges() link count", tt.expected.links, len(got.List))
			assertEqual(t, "FilterPrivilegeGroupPrivileges() total rows", tt.expected.totalRows, got.TotalRows)
		})
	}
}
