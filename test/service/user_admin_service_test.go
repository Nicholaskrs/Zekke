package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	dto "template-go/base/base"
	"template-go/base/helpers"
	"template-go/data/model"
	permissionSvc "template-go/modules/permission/svc"
	userSvc "template-go/modules/user/svc"
)

func TestUserAdminService_LoginUserAdmin(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		tokens  int64
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, f *userAdminFixture) *userSvc.LoginUserAdminIn
		expected expected
	}

	tests := []testCase{
		{
			name: "issues tokens for valid administrator credentials",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.LoginUserAdminIn {
				seedUser(t, f.db, withUsername("admin"), withRole("Admin"), withPasswordHash(hashPassword(t, "correct")))
				f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
					return &permissionSvc.GetAdminGroupPrivilegeOut{
						BaseOut: dto.BaseOut{Success: true}, AdminPrivilegeGroup: "Root",
					}
				}
				return &userSvc.LoginUserAdminIn{Trace: testTrace(), Username: "admin", Password: "correct"}
			},
			expected: expected{success: true, tokens: 1},
		},
		{
			name: "rejects non-administrator account",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.LoginUserAdminIn {
				seedUser(t, f.db, withUsername("regular"), withRole("User"), withPasswordHash(hashPassword(t, "correct")))
				return &userSvc.LoginUserAdminIn{Trace: testTrace(), Username: "regular", Password: "correct"}
			},
			expected: expected{code: http.StatusUnauthorized},
		},
		{
			name: "returns permission service failure",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.LoginUserAdminIn {
				seedUser(t, f.db, withUsername("admin"), withRole("Admin"), withPasswordHash(hashPassword(t, "correct")))
				f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
					return &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission unavailable"}}
				}
				return &userSvc.LoginUserAdminIn{Trace: testTrace(), Username: "admin", Password: "correct"}
			},
			expected: expected{code: http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.LoginUserAdmin(context.Background(), tt.setup(t, f))
			assertEqual(t, "LoginUserAdmin() success", tt.expected.success, got.Success)
			assertEqual(t, "LoginUserAdmin() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.RefreshToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.tokens {
				t.Fatalf("persisted refresh token count = %d, want %d", count, tt.expected.tokens)
			}
			if tt.expected.success && (got.Token == "" || got.RefreshToken == "") {
				t.Fatal("LoginUserAdmin() returned empty token")
			}
		})
	}
}

func TestUserAdminService_FilterUser(t *testing.T) {
	t.Parallel()

	type expected struct {
		success   bool
		code      int
		users     int
		totalRows int
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, f *userAdminFixture) *userSvc.AdminFilterUserIn
		expected expected
	}

	tests := []testCase{
		{
			name: "returns active users and count",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminFilterUserIn {
				seedUser(t, f.db, withUsername("active"))
				deletedAt := time.Now()
				seed := seedUser(t, f.db, withUsername("deleted"))
				if err := f.db.Model(seed).Update("deleted_at", deletedAt).Error; err != nil {
					t.Fatal(err)
				}
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
				}
				return &userSvc.AdminFilterUserIn{
					Trace: testTrace(), PaginateIn: dto.PaginateIn{Page: 1, Limit: 10},
					AdminUserId: 1, AdminPrivilegeGroup: "Root",
				}
			},
			expected: expected{success: true, users: 1, totalRows: 1},
		},
		{
			name: "rejects denied permission",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminFilterUserIn {
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}}
				}
				return &userSvc.AdminFilterUserIn{
					Trace: testTrace(), PaginateIn: dto.PaginateIn{Page: 1, Limit: 10},
					AdminUserId: 1, AdminPrivilegeGroup: "Root",
				}
			},
			expected: expected{code: http.StatusForbidden},
		},
		{
			name: "propagates permission service error",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminFilterUserIn {
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission failure"}}
				}
				return &userSvc.AdminFilterUserIn{
					Trace: testTrace(), PaginateIn: dto.PaginateIn{Page: 1, Limit: 10},
					AdminUserId: 1, AdminPrivilegeGroup: "Root",
				}
			},
			expected: expected{code: http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.FilterUser(context.Background(), tt.setup(t, f))
			assertEqual(t, "FilterUser() success", tt.expected.success, got.Success)
			assertEqual(t, "FilterUser() error code", tt.expected.code, got.ErrorCode)
			assertEqual(t, "FilterUser() user count", tt.expected.users, len(got.Users))
			assertEqual(t, "FilterUser() total rows", tt.expected.totalRows, got.TotalRows)
			if calls := f.permissionMock.HasPrivilegeCalls(); len(calls) != 1 {
				t.Fatalf("HasPrivilege calls = %d, want 1", len(calls))
			}
		})
	}
}

func TestUserAdminService_DeleteUser(t *testing.T) {
	t.Parallel()

	type expected struct {
		success       bool
		code          int
		activeUsers   int64
		privilegeCall int
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, f *userAdminFixture) *userSvc.AdminDeleteUserIn
		expected expected
	}

	tests := []testCase{
		{
			name: "soft-deletes target user when authorized",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminDeleteUserIn {
				user := seedUser(t, f.db)
				seedRefreshToken(t, f.db, user.ID, "active-user-token", "user-session", time.Now().Add(time.Hour), nil)
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
				}
				return &userSvc.AdminDeleteUserIn{Trace: testTrace(), AdminPrivilegeGroup: "Root", UserId: int(user.ID)}
			},
			expected: expected{success: true, privilegeCall: 1},
		},
		{
			name: "does not modify users when denied",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminDeleteUserIn {
				user := seedUser(t, f.db)
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}}
				}
				return &userSvc.AdminDeleteUserIn{Trace: testTrace(), AdminPrivilegeGroup: "Root", UserId: int(user.ID)}
			},
			expected: expected{code: http.StatusForbidden, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "propagates permission dependency failure",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminDeleteUserIn {
				user := seedUser(t, f.db)
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission unavailable"}}
				}
				return &userSvc.AdminDeleteUserIn{Trace: testTrace(), AdminPrivilegeGroup: "Root", UserId: int(user.ID)}
			},
			expected: expected{code: http.StatusServiceUnavailable, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "returns not found for missing target",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminDeleteUserIn {
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
				}
				return &userSvc.AdminDeleteUserIn{Trace: testTrace(), AdminPrivilegeGroup: "Root", UserId: 42}
			},
			expected: expected{code: http.StatusNotFound, privilegeCall: 1},
		},
		{
			name: "rejects zero target ID without checking permission",
			setup: func(t *testing.T, _ *userAdminFixture) *userSvc.AdminDeleteUserIn {
				return &userSvc.AdminDeleteUserIn{Trace: testTrace(), AdminPrivilegeGroup: "Root"}
			},
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.DeleteUser(context.Background(), tt.setup(t, f))
			assertEqual(t, "DeleteUser() success", tt.expected.success, got.Success)
			assertEqual(t, "DeleteUser() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.activeUsers {
				t.Fatalf("active user count = %d, want %d", count, tt.expected.activeUsers)
			}
			if tt.expected.success {
				var token model.RefreshToken
				if err := f.db.Where("token_hash = ?", helpers.HashToken("active-user-token")).First(&token).Error; err != nil {
					t.Fatal(err)
				}
				if token.RevokedAt == nil {
					t.Fatal("refresh token was not revoked with deleted user")
				}
			}
			if calls := f.permissionMock.HasPrivilegeCalls(); len(calls) != tt.expected.privilegeCall {
				t.Fatalf("HasPrivilege calls = %d, want %d", len(calls), tt.expected.privilegeCall)
			}
		})
	}
}

func TestUserAdminService_Refresh(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		tokens  int64
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn
		expected expected
	}

	tests := []testCase{
		{
			name: "rotates administrator refresh token",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn {
				user := seedUser(t, f.db, withRole("Admin"))
				seedUserPrivilegeGroup(t, f.db, user.ID, "Root")
				seedRefreshToken(t, f.db, user.ID, "old-admin-token", "admin-family", time.Now().Add(time.Hour), nil)
				return &userSvc.AdminRefreshIn{Trace: testTrace(), AdminUserId: int(user.ID), RefreshToken: "old-admin-token"}
			},
			expected: expected{success: true, tokens: 2},
		},
		{
			name: "returns not found for unknown token",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn {
				user := seedUser(t, f.db, withRole("Admin"))
				return &userSvc.AdminRefreshIn{Trace: testTrace(), AdminUserId: int(user.ID), RefreshToken: "missing"}
			},
			expected: expected{code: http.StatusNotFound},
		},
		{
			name: "rejects expired token",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn {
				user := seedUser(t, f.db, withRole("Admin"))
				seedRefreshToken(t, f.db, user.ID, "expired-admin", "admin-family", time.Now().Add(-time.Hour), nil)
				return &userSvc.AdminRefreshIn{Trace: testTrace(), AdminUserId: int(user.ID), RefreshToken: "expired-admin"}
			},
			expected: expected{code: http.StatusForbidden, tokens: 1},
		},
		{
			name: "revokes family when used token is reused",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn {
				user := seedUser(t, f.db, withRole("Admin"))
				usedAt := time.Now().Add(-time.Minute)
				seedRefreshToken(t, f.db, user.ID, "used-admin", "admin-family", time.Now().Add(time.Hour), &usedAt)
				seedRefreshToken(t, f.db, user.ID, "sibling-admin", "admin-family", time.Now().Add(time.Hour), nil)
				return &userSvc.AdminRefreshIn{Trace: testTrace(), AdminUserId: int(user.ID), RefreshToken: "used-admin"}
			},
			expected: expected{code: http.StatusForbidden, tokens: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.Refresh(context.Background(), tt.setup(t, f))
			assertEqual(t, "Refresh() success", tt.expected.success, got.Success)
			assertEqual(t, "Refresh() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.RefreshToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.tokens {
				t.Fatalf("refresh token count = %d, want %d", count, tt.expected.tokens)
			}
			if tt.name == "revokes family when used token is reused" {
				var revoked int64
				if err := f.db.Model(&model.RefreshToken{}).Where("revoked_at IS NOT NULL").Count(&revoked).Error; err != nil {
					t.Fatal(err)
				}
				if revoked != 2 {
					t.Fatalf("revoked refresh token count = %d, want 2", revoked)
				}
			}
		})
	}
}

func TestUserAdminService_ChangeAdminPassword(t *testing.T) {
	t.Parallel()

	type expected struct {
		success        bool
		code           int
		storedPassword string
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, f *userAdminFixture) *userSvc.AdminChangePasswordIn
		expected expected
	}

	tests := []testCase{
		{
			name: "updates password and issues a new refresh token",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminChangePasswordIn {
				user := seedUser(t, f.db, withRole("Admin"), withPasswordHash(hashPassword(t, "old-password")))
				seedRefreshToken(t, f.db, user.ID, "previous-token", "previous-family", time.Now().Add(time.Hour), nil)
				f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
					return &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, AdminPrivilegeGroup: "Root"}
				}
				return &userSvc.AdminChangePasswordIn{
					Trace: testTrace(), UserAdminId: int(user.ID), OldPassword: "old-password", NewPassword: "new-password",
				}
			},
			expected: expected{success: true, storedPassword: "new-password"},
		},
		{
			name: "rejects incorrect old password without changing data",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminChangePasswordIn {
				user := seedUser(t, f.db, withRole("Admin"), withPasswordHash(hashPassword(t, "old-password")))
				return &userSvc.AdminChangePasswordIn{
					Trace: testTrace(), UserAdminId: int(user.ID), OldPassword: "wrong", NewPassword: "new-password",
				}
			},
			expected: expected{code: http.StatusBadRequest, storedPassword: "old-password"},
		},
		{
			name: "rolls back password update if permission lookup fails",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminChangePasswordIn {
				user := seedUser(t, f.db, withRole("Admin"), withPasswordHash(hashPassword(t, "old-password")))
				f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
					return &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission unavailable"}}
				}
				return &userSvc.AdminChangePasswordIn{
					Trace: testTrace(), UserAdminId: int(user.ID), OldPassword: "old-password", NewPassword: "new-password",
				}
			},
			expected: expected{code: http.StatusServiceUnavailable, storedPassword: "old-password"},
		},
		{
			name: "validates required ID",
			setup: func(t *testing.T, _ *userAdminFixture) *userSvc.AdminChangePasswordIn {
				return &userSvc.AdminChangePasswordIn{Trace: testTrace(), OldPassword: "old", NewPassword: "new"}
			},
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			input := tt.setup(t, f)
			got := f.service.ChangeAdminPassword(context.Background(), input)
			assertEqual(t, "ChangeAdminPassword() success", tt.expected.success, got.Success)
			assertEqual(t, "ChangeAdminPassword() error code", tt.expected.code, got.ErrorCode)
			var user model.User
			err := f.db.First(&user, input.UserAdminId).Error
			if err == nil && tt.expected.storedPassword != "" && !helpers.VerifyPassword(tt.expected.storedPassword, user.Password) {
				t.Fatalf("stored password does not verify against expected password %q", tt.expected.storedPassword)
			}
			var refreshTokens []model.RefreshToken
			if err := f.db.Order("id").Find(&refreshTokens).Error; err != nil {
				t.Fatal(err)
			}
			if tt.expected.success {
				if len(refreshTokens) != 2 || refreshTokens[0].RevokedAt == nil || got.RefreshToken == "" {
					t.Fatalf("refresh token state = %+v; expected old token revoked and replacement issued", refreshTokens)
				}
			} else if len(refreshTokens) != 0 {
				t.Fatalf("failed password change modified refresh tokens: %+v", refreshTokens)
			}
		})
	}
}

func TestUserAdminService_RegisterNewAdmin(t *testing.T) {
	t.Parallel()

	type expected struct {
		success        bool
		code           int
		users          int64
		assignCalls    int
		privilegeCalls int
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, f *userAdminFixture) *userSvc.RegisterNewAdminIn
		expected expected
	}

	tests := []testCase{
		{
			name: "registers admin and assigns requested group",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.RegisterNewAdminIn {
				f.permissionMock.CheckAdminGroupPrivilegeExistsFunc = func(context.Context, *permissionSvc.CheckAdminGroupPrivilegeExistsIn) *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
					return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{BaseOut: dto.BaseOut{Success: true}, IsExists: true}
				}
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
				}
				f.permissionMock.ChangeAdminPrivilegeRoleFunc = func(context.Context, *permissionSvc.ChangeAdminPrivilegeRoleIn) *permissionSvc.ChangeAdminPrivilegeRoleOut {
					return &permissionSvc.ChangeAdminPrivilegeRoleOut{BaseOut: dto.BaseOut{Success: true}}
				}
				return &userSvc.RegisterNewAdminIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					Username: "new-admin", Email: "new-admin@example.test", Password: "password",
					FullName: "New Admin", PrivilegeGroup: "Auditor",
				}
			},
			expected: expected{success: true, users: 1, assignCalls: 1, privilegeCalls: 1},
		},
		{
			name: "rejects a privilege group that does not exist",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.RegisterNewAdminIn {
				f.permissionMock.CheckAdminGroupPrivilegeExistsFunc = func(context.Context, *permissionSvc.CheckAdminGroupPrivilegeExistsIn) *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
					return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{BaseOut: dto.BaseOut{Success: true}}
				}
				return &userSvc.RegisterNewAdminIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					Username: "new-admin", Email: "new-admin@example.test", Password: "password", PrivilegeGroup: "Missing",
				}
			},
			expected: expected{code: http.StatusBadRequest},
		},
		{
			name: "rejects caller without registration privilege",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.RegisterNewAdminIn {
				f.permissionMock.CheckAdminGroupPrivilegeExistsFunc = func(context.Context, *permissionSvc.CheckAdminGroupPrivilegeExistsIn) *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
					return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{BaseOut: dto.BaseOut{Success: true}, IsExists: true}
				}
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}}
				}
				return &userSvc.RegisterNewAdminIn{Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root", PrivilegeGroup: "Auditor"}
			},
			expected: expected{code: http.StatusForbidden, privilegeCalls: 1},
		},
		{
			name: "rolls back created user if group assignment fails",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.RegisterNewAdminIn {
				f.permissionMock.CheckAdminGroupPrivilegeExistsFunc = func(context.Context, *permissionSvc.CheckAdminGroupPrivilegeExistsIn) *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
					return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{BaseOut: dto.BaseOut{Success: true}, IsExists: true}
				}
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
				}
				f.permissionMock.ChangeAdminPrivilegeRoleFunc = func(context.Context, *permissionSvc.ChangeAdminPrivilegeRoleIn) *permissionSvc.ChangeAdminPrivilegeRoleOut {
					return &permissionSvc.ChangeAdminPrivilegeRoleOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusConflict, ErrorMessage: "group assignment failed"}}
				}
				return &userSvc.RegisterNewAdminIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					Username: "new-admin", Email: "new-admin@example.test", Password: "password", PrivilegeGroup: "Auditor",
				}
			},
			expected: expected{code: http.StatusConflict, assignCalls: 1, privilegeCalls: 1},
		},
		{
			name: "rejects duplicate username",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.RegisterNewAdminIn {
				seedUser(t, f.db, withUsername("taken"))
				f.permissionMock.CheckAdminGroupPrivilegeExistsFunc = func(context.Context, *permissionSvc.CheckAdminGroupPrivilegeExistsIn) *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
					return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{BaseOut: dto.BaseOut{Success: true}, IsExists: true}
				}
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
				}
				return &userSvc.RegisterNewAdminIn{
					Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
					Username: "taken", Email: "new-admin@example.test", Password: "password", PrivilegeGroup: "Auditor",
				}
			},
			expected: expected{code: http.StatusBadRequest, users: 1, privilegeCalls: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.RegisterNewAdmin(context.Background(), tt.setup(t, f))
			assertEqual(t, "RegisterNewAdmin() success", tt.expected.success, got.Success)
			assertEqual(t, "RegisterNewAdmin() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.User{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.users {
				t.Fatalf("persisted user count = %d, want %d", count, tt.expected.users)
			}
			if calls := f.permissionMock.ChangeAdminPrivilegeRoleCalls(); len(calls) != tt.expected.assignCalls {
				t.Fatalf("ChangeAdminPrivilegeRole calls = %d, want %d", len(calls), tt.expected.assignCalls)
			}
			if calls := f.permissionMock.HasPrivilegeCalls(); len(calls) != tt.expected.privilegeCalls {
				t.Fatalf("HasPrivilege calls = %d, want %d", len(calls), tt.expected.privilegeCalls)
			}
		})
	}
}
