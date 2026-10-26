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

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *userAdminFixture) *userSvc.LoginUserAdminIn
		wantSuccess bool
		wantCode    int
		wantTokens  int64
	}{
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
			wantSuccess: true,
			wantTokens:  1,
		},
		{
			name: "rejects non-administrator account",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.LoginUserAdminIn {
				seedUser(t, f.db, withUsername("regular"), withRole("User"), withPasswordHash(hashPassword(t, "correct")))
				return &userSvc.LoginUserAdminIn{Trace: testTrace(), Username: "regular", Password: "correct"}
			},
			wantCode: http.StatusUnauthorized,
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
			wantCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.LoginUserAdmin(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("LoginUserAdmin() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var count int64
			if err := f.db.Model(&model.RefreshToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantTokens {
				t.Fatalf("persisted refresh token count = %d, want %d", count, tt.wantTokens)
			}
			if tt.wantSuccess && (got.Token == "" || got.RefreshToken == "") {
				t.Fatal("LoginUserAdmin() returned empty token")
			}
		})
	}
}

func TestUserAdminService_FilterUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T, f *userAdminFixture) *userSvc.AdminFilterUserIn
		wantSuccess   bool
		wantCode      int
		wantUsers     int
		wantTotalRows int
	}{
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
			wantSuccess:   true,
			wantUsers:     1,
			wantTotalRows: 1,
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
			wantCode: http.StatusForbidden,
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
			wantCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.FilterUser(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode ||
				len(got.Users) != tt.wantUsers || got.TotalRows != tt.wantTotalRows {
				t.Fatalf("FilterUser() = %+v, want success %t, users %d, total %d, and code %d",
					got, tt.wantSuccess, tt.wantUsers, tt.wantTotalRows, tt.wantCode)
			}
			if calls := f.permissionMock.HasPrivilegeCalls(); len(calls) != 1 {
				t.Fatalf("HasPrivilege calls = %d, want 1", len(calls))
			}
		})
	}
}

func TestUserAdminService_DeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		setup             func(t *testing.T, f *userAdminFixture) *userSvc.AdminDeleteUserIn
		wantSuccess       bool
		wantCode          int
		wantActiveUsers   int64
		wantPrivilegeCall int
	}{
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
			wantSuccess:       true,
			wantPrivilegeCall: 1,
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
			wantCode:          http.StatusForbidden,
			wantActiveUsers:   1,
			wantPrivilegeCall: 1,
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
			wantCode:          http.StatusServiceUnavailable,
			wantActiveUsers:   1,
			wantPrivilegeCall: 1,
		},
		{
			name: "returns not found for missing target",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminDeleteUserIn {
				f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
				}
				return &userSvc.AdminDeleteUserIn{Trace: testTrace(), AdminPrivilegeGroup: "Root", UserId: 42}
			},
			wantCode:          http.StatusNotFound,
			wantPrivilegeCall: 1,
		},
		{
			name: "rejects zero target ID without checking permission",
			setup: func(t *testing.T, _ *userAdminFixture) *userSvc.AdminDeleteUserIn {
				return &userSvc.AdminDeleteUserIn{Trace: testTrace(), AdminPrivilegeGroup: "Root"}
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.DeleteUser(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("DeleteUser() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var count int64
			if err := f.db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantActiveUsers {
				t.Fatalf("active user count = %d, want %d", count, tt.wantActiveUsers)
			}
			if tt.wantSuccess {
				var token model.RefreshToken
				if err := f.db.Where("token_hash = ?", helpers.HashToken("active-user-token")).First(&token).Error; err != nil {
					t.Fatal(err)
				}
				if token.RevokedAt == nil {
					t.Fatal("refresh token was not revoked with deleted user")
				}
			}
			if calls := f.permissionMock.HasPrivilegeCalls(); len(calls) != tt.wantPrivilegeCall {
				t.Fatalf("HasPrivilege calls = %d, want %d", len(calls), tt.wantPrivilegeCall)
			}
		})
	}
}

func TestUserAdminService_Refresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn
		wantSuccess bool
		wantCode    int
		wantTokens  int64
	}{
		{
			name: "rotates administrator refresh token",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn {
				user := seedUser(t, f.db, withRole("Admin"))
				seedUserPrivilegeGroup(t, f.db, user.ID, "Root")
				seedRefreshToken(t, f.db, user.ID, "old-admin-token", "admin-family", time.Now().Add(time.Hour), nil)
				return &userSvc.AdminRefreshIn{Trace: testTrace(), AdminUserId: int(user.ID), RefreshToken: "old-admin-token"}
			},
			wantSuccess: true,
			wantTokens:  2,
		},
		{
			name: "returns not found for unknown token",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn {
				user := seedUser(t, f.db, withRole("Admin"))
				return &userSvc.AdminRefreshIn{Trace: testTrace(), AdminUserId: int(user.ID), RefreshToken: "missing"}
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "rejects expired token",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminRefreshIn {
				user := seedUser(t, f.db, withRole("Admin"))
				seedRefreshToken(t, f.db, user.ID, "expired-admin", "admin-family", time.Now().Add(-time.Hour), nil)
				return &userSvc.AdminRefreshIn{Trace: testTrace(), AdminUserId: int(user.ID), RefreshToken: "expired-admin"}
			},
			wantCode:   http.StatusForbidden,
			wantTokens: 1,
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
			wantCode:   http.StatusForbidden,
			wantTokens: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.Refresh(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("Refresh() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var count int64
			if err := f.db.Model(&model.RefreshToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantTokens {
				t.Fatalf("refresh token count = %d, want %d", count, tt.wantTokens)
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

	tests := []struct {
		name               string
		setup              func(t *testing.T, f *userAdminFixture) *userSvc.AdminChangePasswordIn
		wantSuccess        bool
		wantCode           int
		wantStoredPassword string
	}{
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
			wantSuccess:        true,
			wantStoredPassword: "new-password",
		},
		{
			name: "rejects incorrect old password without changing data",
			setup: func(t *testing.T, f *userAdminFixture) *userSvc.AdminChangePasswordIn {
				user := seedUser(t, f.db, withRole("Admin"), withPasswordHash(hashPassword(t, "old-password")))
				return &userSvc.AdminChangePasswordIn{
					Trace: testTrace(), UserAdminId: int(user.ID), OldPassword: "wrong", NewPassword: "new-password",
				}
			},
			wantCode:           http.StatusBadRequest,
			wantStoredPassword: "old-password",
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
			wantCode:           http.StatusServiceUnavailable,
			wantStoredPassword: "old-password",
		},
		{
			name: "validates required ID",
			setup: func(t *testing.T, _ *userAdminFixture) *userSvc.AdminChangePasswordIn {
				return &userSvc.AdminChangePasswordIn{Trace: testTrace(), OldPassword: "old", NewPassword: "new"}
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			input := tt.setup(t, f)
			got := f.service.ChangeAdminPassword(context.Background(), input)
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("ChangeAdminPassword() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var user model.User
			err := f.db.First(&user, input.UserAdminId).Error
			if err == nil && tt.wantStoredPassword != "" && !helpers.VerifyPassword(tt.wantStoredPassword, user.Password) {
				t.Fatalf("stored password does not verify against expected password %q", tt.wantStoredPassword)
			}
			var refreshTokens []model.RefreshToken
			if err := f.db.Order("id").Find(&refreshTokens).Error; err != nil {
				t.Fatal(err)
			}
			if tt.wantSuccess {
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

	tests := []struct {
		name               string
		setup              func(t *testing.T, f *userAdminFixture) *userSvc.RegisterNewAdminIn
		wantSuccess        bool
		wantCode           int
		wantUsers          int64
		wantAssignCalls    int
		wantPrivilegeCalls int
	}{
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
			wantSuccess:        true,
			wantUsers:          1,
			wantAssignCalls:    1,
			wantPrivilegeCalls: 1,
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
			wantCode: http.StatusBadRequest,
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
			wantCode:           http.StatusForbidden,
			wantPrivilegeCalls: 1,
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
			wantCode:           http.StatusConflict,
			wantAssignCalls:    1,
			wantPrivilegeCalls: 1,
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
			wantCode:           http.StatusBadRequest,
			wantUsers:          1,
			wantPrivilegeCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			got := f.service.RegisterNewAdmin(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("RegisterNewAdmin() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var count int64
			if err := f.db.Model(&model.User{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantUsers {
				t.Fatalf("persisted user count = %d, want %d", count, tt.wantUsers)
			}
			if calls := f.permissionMock.ChangeAdminPrivilegeRoleCalls(); len(calls) != tt.wantAssignCalls {
				t.Fatalf("ChangeAdminPrivilegeRole calls = %d, want %d", len(calls), tt.wantAssignCalls)
			}
			if calls := f.permissionMock.HasPrivilegeCalls(); len(calls) != tt.wantPrivilegeCalls {
				t.Fatalf("HasPrivilege calls = %d, want %d", len(calls), tt.wantPrivilegeCalls)
			}
		})
	}
}
