package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	dto "template-go/base/base"
	"template-go/base/helpers"
	permissionSvc "template-go/modules/permission/svc"
	userPermission "template-go/modules/user/permission"
	userSvc "template-go/modules/user/svc"
	"template-go/test"
	"template-go/test/fixture"
)

func TestUserAdminService_LoginUserAdmin(t *testing.T) {
	t.Parallel()

	type input struct {
		Username string
		Password string
	}
	type expected struct {
		success        bool
		errorCode      int
		privilegeCalls int
		tokenRows      int
	}
	type testCase struct {
		name               string
		input              input
		expected           expected
		privilegeGroupMock *permissionSvc.GetAdminGroupPrivilegeOut
	}

	defaultPrivilegeGroupMock := func() *permissionSvc.GetAdminGroupPrivilegeOut {
		return &permissionSvc.GetAdminGroupPrivilegeOut{
			BaseOut:             dto.BaseOut{Success: true},
			AdminPrivilegeGroup: "SuperAdmin",
		}
	}
	tests := []testCase{
		{
			name:               "authenticates admin",
			input:              input{Username: "admin", Password: "correct-password"},
			expected:           expected{success: true, privilegeCalls: 1, tokenRows: 1},
			privilegeGroupMock: defaultPrivilegeGroupMock(),
		},
		{
			name:               "rejects incorrect password",
			input:              input{Username: "admin", Password: "wrong-password"},
			expected:           expected{errorCode: http.StatusUnauthorized},
			privilegeGroupMock: defaultPrivilegeGroupMock(),
		},
		{
			name:     "returns dependency failure",
			input:    input{Username: "admin", Password: "correct-password"},
			expected: expected{errorCode: http.StatusInternalServerError, privilegeCalls: 1},
			privilegeGroupMock: fixture.Modify(defaultPrivilegeGroupMock, func(out *permissionSvc.GetAdminGroupPrivilegeOut) {
				out.Success = false
				out.ErrorCode = http.StatusInternalServerError
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			passwordHash, err := helpers.HashPassword("correct-password")
			if err != nil {
				t.Fatal(err)
			}
			admin := f.SeedUser(t, fixture.UserRow{Username: "admin", Role: "Admin", PasswordHash: passwordHash})
			f.PermissionService.GetAdminGroupPrivilegeFunc = func(ctx context.Context, in *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
				test.AssertEqual(t, "GetAdminGroupPrivilege.UserAdminId", int(admin.ID), in.UserAdminId)
				return tt.privilegeGroupMock
			}

			got := f.AdminService.LoginUserAdmin(context.Background(), &userSvc.LoginUserAdminIn{
				Trace: fixture.TestTrace(), Username: tt.input.Username, Password: tt.input.Password,
			})
			test.AssertEqual(t, "LoginUserAdmin success", tt.expected.success, got.Success)
			test.AssertEqual(t, "LoginUserAdmin error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "GetAdminGroupPrivilege calls", tt.expected.privilegeCalls, len(f.PermissionService.GetAdminGroupPrivilegeCalls()))
			test.AssertEqual(t, "refresh token rows", tt.expected.tokenRows, len(f.UserQuery.FindRefreshTokens(admin.ID)))
			if tt.expected.success {
				test.AssertEqual(t, "LoginUserAdmin user ID", admin.ID, got.UserID)
				test.AssertEqual(t, "LoginUserAdmin privilege group", "SuperAdmin", got.AdminGroupPrivilege)
				test.AssertNotEmpty(t, "LoginUserAdmin access token", got.Token)
				test.AssertNotEmpty(t, "LoginUserAdmin refresh token", got.RefreshToken)
			}
		})
	}
}

func TestUserAdminService_FilterUser(t *testing.T) {
	t.Parallel()

	type input struct {
		Page                int
		Limit               int
		AdminUserId         int
		AdminPrivilegeGroup string
	}
	type expected struct {
		success        bool
		errorCode      int
		privilegeCalls int
		userRows       int
		totalRows      int
	}
	type testCase struct {
		name             string
		input            input
		expected         expected
		hasPrivilegeMock *permissionSvc.HasPrivilegeOut
	}

	defaultHasPrivilegeMock := func() *permissionSvc.HasPrivilegeOut {
		return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
	}
	tests := []testCase{
		{
			name:             "filters active users",
			input:            input{Page: 1, Limit: 10, AdminUserId: 10, AdminPrivilegeGroup: "SuperAdmin"},
			expected:         expected{success: true, privilegeCalls: 1, userRows: 2, totalRows: 2},
			hasPrivilegeMock: defaultHasPrivilegeMock(),
		},
		{
			name:     "rejects caller without privilege",
			input:    input{Page: 1, Limit: 10, AdminUserId: 10, AdminPrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusForbidden, privilegeCalls: 1},
			hasPrivilegeMock: fixture.Modify(defaultHasPrivilegeMock, func(out *permissionSvc.HasPrivilegeOut) {
				out.IsAllowed = false
			}),
		},
		{
			name:     "returns privilege dependency failure",
			input:    input{Page: 1, Limit: 10, AdminUserId: 10, AdminPrivilegeGroup: "SuperAdmin"},
			expected: expected{errorCode: http.StatusInternalServerError, privilegeCalls: 1},
			hasPrivilegeMock: fixture.Modify(defaultHasPrivilegeMock, func(out *permissionSvc.HasPrivilegeOut) {
				out.Success = false
				out.ErrorCode = http.StatusInternalServerError
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			f.SeedUsers(t, []fixture.UserRow{{Username: "user-a"}, {Username: "user-b", Role: "Admin"}})
			f.PermissionService.HasPrivilegeFunc = func(ctx context.Context, in *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				test.AssertEqual(t, "HasPrivilege.PrivilegeGroup", tt.input.AdminPrivilegeGroup, in.PrivilegeGroup)
				test.AssertEqual(t, "HasPrivilege.Privilege", userPermission.PVFilterUser, in.Privilege)
				return tt.hasPrivilegeMock
			}

			got := f.AdminService.FilterUser(context.Background(), &userSvc.AdminFilterUserIn{
				Trace: fixture.TestTrace(), AdminUserId: tt.input.AdminUserId,
				AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup,
				PaginateIn:          dto.PaginateIn{Page: tt.input.Page, Limit: tt.input.Limit},
			})
			test.AssertEqual(t, "FilterUser success", tt.expected.success, got.Success)
			test.AssertEqual(t, "FilterUser error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "HasPrivilege calls", tt.expected.privilegeCalls, len(f.PermissionService.HasPrivilegeCalls()))
			test.AssertEqual(t, "FilterUser rows", tt.expected.userRows, len(got.Users))
			test.AssertEqual(t, "FilterUser total rows", tt.expected.totalRows, got.TotalRows)
		})
	}
}

func TestUserAdminService_DeleteUser(t *testing.T) {
	t.Parallel()

	type input struct {
		AdminUserId         int
		AdminPrivilegeGroup string
		UserId              int
	}
	type expected struct {
		success        bool
		errorCode      int
		privilegeCalls int
		deleted        bool
		revoked        bool
	}
	type testCase struct {
		name             string
		input            input
		expected         expected
		hasPrivilegeMock *permissionSvc.HasPrivilegeOut
	}

	defaultHasPrivilegeMock := func() *permissionSvc.HasPrivilegeOut {
		return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true}
	}
	tests := []testCase{
		{
			name:             "soft deletes user and revokes refresh tokens",
			input:            input{AdminUserId: 10, AdminPrivilegeGroup: "SuperAdmin"},
			expected:         expected{success: true, privilegeCalls: 1, deleted: true, revoked: true},
			hasPrivilegeMock: defaultHasPrivilegeMock(),
		},
		{
			name:             "rejects empty user ID",
			expected:         expected{errorCode: http.StatusBadRequest},
			hasPrivilegeMock: defaultHasPrivilegeMock(),
		},
		{
			name:             "returns not found for missing user",
			input:            input{AdminUserId: 10, AdminPrivilegeGroup: "SuperAdmin", UserId: 99999},
			expected:         expected{errorCode: http.StatusNotFound, privilegeCalls: 1},
			hasPrivilegeMock: defaultHasPrivilegeMock(),
		},
		{
			name:     "rejects caller without privilege",
			input:    input{AdminUserId: 10, AdminPrivilegeGroup: "Auditor"},
			expected: expected{errorCode: http.StatusForbidden, privilegeCalls: 1},
			hasPrivilegeMock: fixture.Modify(defaultHasPrivilegeMock, func(out *permissionSvc.HasPrivilegeOut) {
				out.IsAllowed = false
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{Username: "target-user"})
			f.SeedRefreshToken(t, fixture.RefreshTokenRow{UserID: user.ID, Token: "target-token"})
			userID := tt.input.UserId
			if userID == 0 && tt.expected.errorCode != http.StatusBadRequest {
				userID = int(user.ID)
			}
			f.PermissionService.HasPrivilegeFunc = func(ctx context.Context, in *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				test.AssertEqual(t, "HasPrivilege.PrivilegeGroup", tt.input.AdminPrivilegeGroup, in.PrivilegeGroup)
				test.AssertEqual(t, "HasPrivilege.Privilege", userPermission.PVDeleteUser, in.Privilege)
				return tt.hasPrivilegeMock
			}

			got := f.AdminService.DeleteUser(context.Background(), &userSvc.AdminDeleteUserIn{
				Trace: fixture.TestTrace(), AdminUserId: tt.input.AdminUserId,
				AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup, UserId: userID,
			})
			test.AssertEqual(t, "DeleteUser success", tt.expected.success, got.Success)
			test.AssertEqual(t, "DeleteUser error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "HasPrivilege calls", tt.expected.privilegeCalls, len(f.PermissionService.HasPrivilegeCalls()))
			actual := f.UserQuery.FindByID(user.ID)
			test.AssertEqual(t, "DeleteUser deleted state", tt.expected.deleted, actual.DeletedAt != nil)
			tokens := f.UserQuery.FindRefreshTokens(user.ID)
			test.AssertEqual(t, "DeleteUser token revoked", tt.expected.revoked, tokens[0].RevokedAt != nil)
		})
	}
}

func TestUserAdminService_Refresh(t *testing.T) {
	t.Parallel()

	type input struct {
		AdminUserId  int
		RefreshToken string
	}
	type expected struct {
		success   bool
		errorCode int
		tokenRows int
		revoked   bool
	}
	type testCase struct {
		name      string
		input     input
		expected  expected
		expired   bool
		used      bool
		seedToken string
	}

	tests := []testCase{
		{
			name:     "rotates admin refresh token",
			input:    input{RefreshToken: "admin-token"},
			expected: expected{success: true, tokenRows: 2},
		},
		{
			name:      "rejects unknown refresh token",
			input:     input{RefreshToken: "missing-token"},
			expected:  expected{errorCode: http.StatusNotFound, tokenRows: 1},
			seedToken: "existing-token",
		},
		{
			name:     "rejects expired refresh token",
			input:    input{RefreshToken: "expired-token"},
			expected: expected{errorCode: http.StatusForbidden, tokenRows: 1},
			expired:  true,
		},
		{
			name:     "revokes token family when token is reused",
			input:    input{RefreshToken: "used-token"},
			expected: expected{errorCode: http.StatusForbidden, tokenRows: 1, revoked: true},
			used:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{Username: "admin", Role: "Admin"})
			f.SeedUserPrivilegeGroup(t, user.ID, "SuperAdmin")
			seedToken := tt.seedToken
			if seedToken == "" {
				seedToken = tt.input.RefreshToken
			}
			expireAt := time.Now().Add(time.Hour)
			var usedAt *time.Time
			if tt.expired {
				expireAt = time.Now().Add(-time.Hour)
			}
			if tt.used {
				now := time.Now()
				usedAt = &now
			}
			f.SeedRefreshToken(t, fixture.RefreshTokenRow{
				UserID: user.ID, Token: seedToken, Family: "admin-family", ExpireAt: expireAt, UsedAt: usedAt,
			})
			adminID := tt.input.AdminUserId
			if adminID == 0 {
				adminID = int(user.ID)
			}

			got := f.AdminService.Refresh(context.Background(), &userSvc.AdminRefreshIn{
				Trace: fixture.TestTrace(), AdminUserId: adminID, RefreshToken: tt.input.RefreshToken,
			})
			test.AssertEqual(t, "Admin Refresh success", tt.expected.success, got.Success)
			test.AssertEqual(t, "Admin Refresh error code", tt.expected.errorCode, got.ErrorCode)
			tokens := f.UserQuery.FindRefreshTokens(user.ID)
			test.AssertEqual(t, "Admin Refresh token rows", tt.expected.tokenRows, len(tokens))
			if tt.expected.success {
				test.AssertNotEmpty(t, "Admin Refresh access token", got.AccessToken)
				test.AssertNotEmpty(t, "Admin Refresh token", got.RefreshToken)
				test.AssertEqual(t, "Admin Refresh new token hash", helpers.HashToken(got.RefreshToken), tokens[1].TokenHash)
				test.AssertTrue(t, "Admin Refresh old token used", tokens[0].UsedAt != nil)
			}
			if tt.expected.revoked {
				test.AssertTrue(t, "Admin Refresh token revoked", tokens[0].RevokedAt != nil)
			}
		})
	}
}

func TestUserAdminService_ChangeAdminPassword(t *testing.T) {
	t.Parallel()

	type input struct {
		UserAdminId int
		OldPassword string
		NewPassword string
	}
	type expected struct {
		success        bool
		errorCode      int
		privilegeCalls int
		tokenRows      int
		revoked        bool
	}
	type testCase struct {
		name               string
		input              input
		expected           expected
		privilegeGroupMock *permissionSvc.GetAdminGroupPrivilegeOut
	}

	defaultPrivilegeGroupMock := func() *permissionSvc.GetAdminGroupPrivilegeOut {
		return &permissionSvc.GetAdminGroupPrivilegeOut{
			BaseOut:             dto.BaseOut{Success: true},
			AdminPrivilegeGroup: "SuperAdmin",
		}
	}
	tests := []testCase{
		{
			name:               "changes password and rotates refresh tokens",
			input:              input{OldPassword: "old-password", NewPassword: "new-password"},
			expected:           expected{success: true, privilegeCalls: 1, tokenRows: 2, revoked: true},
			privilegeGroupMock: defaultPrivilegeGroupMock(),
		},
		{
			name:               "rejects empty admin ID",
			input:              input{OldPassword: "old-password", NewPassword: "new-password"},
			expected:           expected{errorCode: http.StatusBadRequest, tokenRows: 1},
			privilegeGroupMock: defaultPrivilegeGroupMock(),
		},
		{
			name:               "rejects incorrect old password",
			input:              input{OldPassword: "wrong-password", NewPassword: "new-password"},
			expected:           expected{errorCode: http.StatusBadRequest, tokenRows: 1},
			privilegeGroupMock: defaultPrivilegeGroupMock(),
		},
		{
			name:     "returns privilege dependency failure without partial update",
			input:    input{OldPassword: "old-password", NewPassword: "new-password"},
			expected: expected{errorCode: http.StatusInternalServerError, privilegeCalls: 1, tokenRows: 1},
			privilegeGroupMock: fixture.Modify(defaultPrivilegeGroupMock, func(out *permissionSvc.GetAdminGroupPrivilegeOut) {
				out.Success = false
				out.ErrorCode = http.StatusInternalServerError
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			passwordHash, err := helpers.HashPassword("old-password")
			if err != nil {
				t.Fatal(err)
			}
			admin := f.SeedUser(t, fixture.UserRow{Username: "admin", Role: "Admin", PasswordHash: passwordHash})
			f.SeedRefreshToken(t, fixture.RefreshTokenRow{UserID: admin.ID, Token: "old-token"})
			adminID := tt.input.UserAdminId
			if tt.name != "rejects empty admin ID" {
				adminID = int(admin.ID)
			}
			f.PermissionService.GetAdminGroupPrivilegeFunc = func(ctx context.Context, in *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
				test.AssertEqual(t, "GetAdminGroupPrivilege.UserAdminId", int(admin.ID), in.UserAdminId)
				return tt.privilegeGroupMock
			}

			got := f.AdminService.ChangeAdminPassword(context.Background(), &userSvc.AdminChangePasswordIn{
				Trace: fixture.TestTrace(), UserAdminId: adminID,
				OldPassword: tt.input.OldPassword, NewPassword: tt.input.NewPassword,
			})
			test.AssertEqual(t, "ChangeAdminPassword success", tt.expected.success, got.Success)
			test.AssertEqual(t, "ChangeAdminPassword error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "GetAdminGroupPrivilege calls", tt.expected.privilegeCalls, len(f.PermissionService.GetAdminGroupPrivilegeCalls()))
			actual := f.UserQuery.FindByID(admin.ID)
			test.AssertEqual(t, "ChangeAdminPassword password changed", tt.expected.success, helpers.VerifyPassword("new-password", actual.Password))
			tokens := f.UserQuery.FindRefreshTokens(admin.ID)
			test.AssertEqual(t, "ChangeAdminPassword token rows", tt.expected.tokenRows, len(tokens))
			test.AssertEqual(t, "ChangeAdminPassword revoked", tt.expected.revoked, tokens[0].RevokedAt != nil)
			if tt.expected.success {
				test.AssertNotEmpty(t, "ChangeAdminPassword returned refresh token", got.RefreshToken)
				test.AssertEqual(t, "ChangeAdminPassword returned token hash", helpers.HashToken(got.RefreshToken), tokens[1].TokenHash)
			}
		})
	}
}
