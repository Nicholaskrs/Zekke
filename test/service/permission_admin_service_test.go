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

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *permissionFixture) *permissionSvc.ChangeAdminPrivilegeRoleIn
		wantSuccess bool
		wantCode    int
		wantGroup   string
	}{
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
			wantSuccess: true,
			wantGroup:   "Operator",
		},
		{
			name: "rejects callers without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.ChangeAdminPrivilegeRoleIn {
				return &permissionSvc.ChangeAdminPrivilegeRoleIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetUserID: 7, NewPrivilegeGroup: "Operator",
				}
			},
			wantCode: http.StatusForbidden,
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
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.ChangeAdminPrivilegeRole(context.Background(), tt.setup(t, f))
			assertEqual(t, "ChangeAdminPrivilegeRole() success", tt.wantSuccess, got.Success)
			assertEqual(t, "ChangeAdminPrivilegeRole() error code", tt.wantCode, got.ErrorCode)
			var mapping model.UserPrivilegeGroup
			err := f.db.Where("user_id = ?", 7).First(&mapping).Error
			if tt.wantSuccess {
				if err != nil {
					t.Fatal(err)
				}
				if mapping.PrivilegeGroupCode != tt.wantGroup {
					t.Fatalf("assigned group = %q, want %q", mapping.PrivilegeGroupCode, tt.wantGroup)
				}
			} else if err == nil {
				t.Fatal("failed operation unexpectedly persisted a user privilege group")
			}
		})
	}
}

func TestPermissionAdminService_InsertNewPrivilegeGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn
		wantSuccess bool
		wantCode    int
		wantGroups  int64
	}{
		{
			name: "inserts group when authorized",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVInsertPrivilegeGroup)
				return &permissionSvc.InsertNewPrivilegeGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					PrivilegeGroup: "Operator", Description: "Operations team",
				}
			},
			wantSuccess: true,
			wantGroups:  1,
		},
		{
			name: "rejects caller without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn {
				return &permissionSvc.InsertNewPrivilegeGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					PrivilegeGroup: "Operator",
				}
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "validates required group",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.InsertNewPrivilegeGroupIn {
				return &permissionSvc.InsertNewPrivilegeGroupIn{Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root"}
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.InsertNewPrivilegeGroup(context.Background(), tt.setup(t, f))
			assertEqual(t, "InsertNewPrivilegeGroup() success", tt.wantSuccess, got.Success)
			assertEqual(t, "InsertNewPrivilegeGroup() error code", tt.wantCode, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.PrivilegeGroup{}).Where("code = ?", "Operator").Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantGroups {
				t.Fatalf("persisted group count = %d, want %d", count, tt.wantGroups)
			}
		})
	}
}

func TestPermissionAdminService_AssignPrivilegeToGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *permissionFixture) *permissionSvc.AssignPrivilegeToGroupIn
		wantSuccess bool
		wantCode    int
		wantLinks   int64
	}{
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
			wantSuccess: true,
			wantLinks:   2,
		},
		{
			name: "rejects caller without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.AssignPrivilegeToGroupIn {
				return &permissionSvc.AssignPrivilegeToGroupIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					TargetPrivilegeGroup: "Operator", PrivilegeCode: "ReadReports",
				}
			},
			wantCode: http.StatusUnauthorized,
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
			wantCode:  http.StatusBadRequest,
			wantLinks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.AssignPrivilegeToGroup(context.Background(), tt.setup(t, f))
			assertEqual(t, "AssignPrivilegeToGroup() success", tt.wantSuccess, got.Success)
			assertEqual(t, "AssignPrivilegeToGroup() error code", tt.wantCode, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.PrivilegeGroupPrivilege{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantLinks {
				t.Fatalf("persisted group privilege count = %d, want %d", count, tt.wantLinks)
			}
		})
	}
}

func TestPermissionAdminService_HasPrivilege(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *permissionFixture) *permissionSvc.HasPrivilegeIn
		wantSuccess bool
		wantAllowed bool
		wantCode    int
	}{
		{
			name: "returns true for assigned privilege",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.HasPrivilegeIn {
				seedGroupPrivilege(t, f.db, "Operator", "ReadReports")
				return &permissionSvc.HasPrivilegeIn{Trace: testTrace(), PrivilegeGroup: "Operator", Privilege: "ReadReports"}
			},
			wantSuccess: true,
			wantAllowed: true,
		},
		{
			name: "returns false for unassigned privilege",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.HasPrivilegeIn {
				seedGroupPrivilege(t, f.db, "Operator", "ReadReports")
				return &permissionSvc.HasPrivilegeIn{Trace: testTrace(), PrivilegeGroup: "Operator", Privilege: "DeleteReports"}
			},
			wantSuccess: true,
		},
		{
			name: "rejects empty privilege group",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.HasPrivilegeIn {
				return &permissionSvc.HasPrivilegeIn{Trace: testTrace(), Privilege: "ReadReports"}
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.HasPrivilege(context.Background(), tt.setup(t, f))
			assertEqual(t, "HasPrivilege() success", tt.wantSuccess, got.Success)
			assertEqual(t, "HasPrivilege() allowed", tt.wantAllowed, got.IsAllowed)
			assertEqual(t, "HasPrivilege() error code", tt.wantCode, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_GetAdminGroupPrivilege(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn
		wantSuccess bool
		wantCode    int
		wantGroup   string
	}{
		{
			name: "returns mapped admin group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn {
				seedUserPrivilegeGroup(t, f.db, 9, "Root")
				return &permissionSvc.GetAdminGroupPrivilegeIn{Trace: testTrace(), UserAdminId: 9}
			},
			wantSuccess: true,
			wantGroup:   "Root",
		},
		{
			name: "returns empty group when user has no mapping",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn {
				return &permissionSvc.GetAdminGroupPrivilegeIn{Trace: testTrace(), UserAdminId: 9}
			},
			wantSuccess: true,
		},
		{
			name: "rejects zero user ID",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.GetAdminGroupPrivilegeIn {
				return &permissionSvc.GetAdminGroupPrivilegeIn{Trace: testTrace()}
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.GetAdminGroupPrivilege(context.Background(), tt.setup(t, f))
			assertEqual(t, "GetAdminGroupPrivilege() success", tt.wantSuccess, got.Success)
			assertEqual(t, "GetAdminGroupPrivilege() group", tt.wantGroup, got.AdminPrivilegeGroup)
			assertEqual(t, "GetAdminGroupPrivilege() error code", tt.wantCode, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_CheckAdminGroupPrivilegeExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn
		wantSuccess bool
		wantExists  bool
		wantCode    int
	}{
		{
			name: "returns true for existing group",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn {
				seedGroupPrivilege(t, f.db, "Root", permissionConst.PVCheckAdminGroupPrivilegeExists)
				seedPrivilegeGroup(t, f.db, "Root")
				return &permissionSvc.CheckAdminGroupPrivilegeExistsIn{Trace: testTrace(), AdminPrivilegeGroup: "Root"}
			},
			wantSuccess: true,
			wantExists:  true,
		},
		{
			name: "returns false when group is missing",
			setup: func(t *testing.T, f *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn {
				seedGroupPrivilege(t, f.db, "Missing", permissionConst.PVCheckAdminGroupPrivilegeExists)
				return &permissionSvc.CheckAdminGroupPrivilegeExistsIn{Trace: testTrace(), AdminPrivilegeGroup: "Missing"}
			},
			wantSuccess: true,
		},
		{
			name: "rejects caller without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.CheckAdminGroupPrivilegeExistsIn {
				return &permissionSvc.CheckAdminGroupPrivilegeExistsIn{Trace: testTrace(), AdminPrivilegeGroup: "Root"}
			},
			wantCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.CheckAdminGroupPrivilegeExists(context.Background(), tt.setup(t, f))
			assertEqual(t, "CheckAdminGroupPrivilegeExists() success", tt.wantSuccess, got.Success)
			assertEqual(t, "CheckAdminGroupPrivilegeExists() exists", tt.wantExists, got.IsExists)
			assertEqual(t, "CheckAdminGroupPrivilegeExists() error code", tt.wantCode, got.ErrorCode)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T, f *permissionFixture) *permissionSvc.FilterPrivilegeGroupIn
		wantSuccess   bool
		wantCode      int
		wantGroups    int
		wantTotalRows int
	}{
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
			wantSuccess:   true,
			wantGroups:    2,
			wantTotalRows: 2,
		},
		{
			name: "rejects invalid pagination",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.FilterPrivilegeGroupIn {
				return &permissionSvc.FilterPrivilegeGroupIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1, Page: 0, Limit: 10,
				}
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "rejects callers without privilege",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.FilterPrivilegeGroupIn {
				return &permissionSvc.FilterPrivilegeGroupIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1, Page: 1, Limit: 10,
				}
			},
			wantCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.FilterPrivilegeGroup(context.Background(), tt.setup(t, f))
			assertEqual(t, "FilterPrivilegeGroup() success", tt.wantSuccess, got.Success)
			assertEqual(t, "FilterPrivilegeGroup() error code", tt.wantCode, got.ErrorCode)
			assertEqual(t, "FilterPrivilegeGroup() group count", tt.wantGroups, len(got.List))
			assertEqual(t, "FilterPrivilegeGroup() total rows", tt.wantTotalRows, got.TotalRows)
		})
	}
}

func TestPermissionAdminService_FilterPrivilegeGroupPrivileges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T, f *permissionFixture) *permissionSvc.FilterPrivilegeGroupPrivilegesIn
		wantSuccess   bool
		wantCode      int
		wantLinks     int
		wantTotalRows int
	}{
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
			wantSuccess:   true,
			wantLinks:     2,
			wantTotalRows: 2,
		},
		{
			name: "rejects missing privilege group",
			setup: func(t *testing.T, _ *permissionFixture) *permissionSvc.FilterPrivilegeGroupPrivilegesIn {
				return &permissionSvc.FilterPrivilegeGroupPrivilegesIn{
					Trace: testTrace(), AdminPrivilegeGroup: "Root", AdminUserId: 1,
					PrivilegeGroup: "Operator", Page: 1, Limit: 10,
				}
			},
			wantCode: http.StatusForbidden,
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
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPermissionFixture(t)
			got := f.service.FilterPrivilegeGroupPrivileges(context.Background(), tt.setup(t, f))
			assertEqual(t, "FilterPrivilegeGroupPrivileges() success", tt.wantSuccess, got.Success)
			assertEqual(t, "FilterPrivilegeGroupPrivileges() error code", tt.wantCode, got.ErrorCode)
			assertEqual(t, "FilterPrivilegeGroupPrivileges() link count", tt.wantLinks, len(got.List))
			assertEqual(t, "FilterPrivilegeGroupPrivileges() total rows", tt.wantTotalRows, got.TotalRows)
		})
	}
}
