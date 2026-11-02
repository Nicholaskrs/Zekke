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
		name                string
		username            string
		password            string
		permissionSuccess   bool
		permissionErrorCode int
		expected            expected
	}

	// Seed data: two users, an administrator and a regular user, both
	// with the same known password.
	//
	// +----------+-----------------+----------+
	// | Username | Password        | Role     |
	// +----------+-----------------+----------+
	// | admin    | correct         | Admin    |
	// | regular  | correct         | User     |
	// +----------+-----------------+----------+
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		passwordHash := hashPassword(t, "correct")
		seedUser(t, f.db, withUsername("admin"), withRole("Admin"), withPasswordHash(passwordHash))
		seedUser(t, f.db, withUsername("regular"), withRole("User"), withPasswordHash(passwordHash))
	}

	// Test cases:
	//
	// +---------------------------------+----------+----------+--------------------+---------+--------+
	// | Scenario                        | Username | Password | Permission Result  | Code    | Tokens |
	// +---------------------------------+----------+----------+--------------------+---------+--------+
	// | valid administrator credentials | admin    | correct  | success/"Root"     | success | 1      |
	// | non-administrator account       | regular  | correct  | (not called)       | 401     | 0      |
	// | permission service failure      | admin    | correct  | 503 error          | 503     | 0      |
	// +---------------------------------+----------+----------+--------------------+---------+--------+
	tests := []testCase{
		{
			name:     "issues tokens for valid administrator credentials",
			username: "admin", password: "correct", permissionSuccess: true,
			expected: expected{success: true, tokens: 1},
		},
		{
			name:     "rejects non-administrator account",
			username: "regular", password: "correct",
			expected: expected{code: http.StatusUnauthorized},
		},
		{
			name:     "returns permission service failure",
			username: "admin", password: "correct", permissionErrorCode: http.StatusServiceUnavailable,
			expected: expected{code: http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
				if tt.permissionSuccess {
					return &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, AdminPrivilegeGroup: "Root"}
				}
				return &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: tt.permissionErrorCode, ErrorMessage: "permission unavailable"}}
			}
			got := f.service.LoginUserAdmin(context.Background(), &userSvc.LoginUserAdminIn{
				Trace: testTrace(), Username: tt.username, Password: tt.password,
			})
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
		name                string
		permissionSuccess   bool
		permissionAllowed   bool
		permissionErrorCode int
		expected            expected
	}

	// Seed data: one active user and one soft-deleted user.
	//
	// +----------+------------+
	// | Username | Deleted    |
	// +----------+------------+
	// | active   | no         |
	// | deleted  | yes        |
	// +----------+------------+
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		seedUser(t, f.db, withUsername("active"))
		deletedAt := time.Now()
		deleted := seedUser(t, f.db, withUsername("deleted"))
		if err := f.db.Model(deleted).Update("deleted_at", deletedAt).Error; err != nil {
			t.Fatal(err)
		}
	}

	// Test cases:
	//
	// +--------------------------+--------------------+----------+---------+-------+-----------+
	// | Scenario                 | Permission Result   | Allowed  | Code    | Users | TotalRows |
	// +--------------------------+--------------------+----------+---------+-------+-----------+
	// | active users, allowed    | success             | true     | success | 1     | 1         |
	// | permission denied        | success             | false    | 403     | 0     | 0         |
	// | permission service error | 503 error           | -        | 503     | 0     | 0         |
	// +--------------------------+--------------------+----------+---------+-------+-----------+
	tests := []testCase{
		{
			name: "returns active users and count", permissionSuccess: true, permissionAllowed: true,
			expected: expected{success: true, users: 1, totalRows: 1},
		},
		{
			name: "rejects denied permission", permissionSuccess: true,
			expected: expected{code: http.StatusForbidden},
		},
		{
			name: "propagates permission service error", permissionErrorCode: http.StatusServiceUnavailable,
			expected: expected{code: http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				if tt.permissionSuccess {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: tt.permissionAllowed}
				}
				return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: tt.permissionErrorCode, ErrorMessage: "permission failure"}}
			}
			got := f.service.FilterUser(context.Background(), &userSvc.AdminFilterUserIn{
				Trace: testTrace(), PaginateIn: dto.PaginateIn{Page: 1, Limit: 10},
				AdminUserId: 1, AdminPrivilegeGroup: "Root",
			})
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
		name                string
		target              string // "seeded", "missing", "zero"
		permissionSuccess   bool
		permissionAllowed   bool
		permissionErrorCode int
		expected            expected
	}

	// Seed data: one user with one active refresh token.
	//
	// +--------------------+--------------+--------+
	// | Raw Token          | Family       | Expires|
	// +--------------------+--------------+--------+
	// | active-user-token  | user-session | +1h    |
	// +--------------------+--------------+--------+
	//
	// A sentinel ID (999999) that can never collide with a seeded row
	// is used for the "missing target" scenario.
	const missingUserID = 999999
	var seededUserID uint
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		user := seedUser(t, f.db)
		seededUserID = user.ID
		seedRefreshToken(t, f.db, user.ID, "active-user-token", "user-session", time.Now().Add(time.Hour), nil)
	}

	// Test cases:
	//
	// +---------------------------------------+---------+--------------------+----------+---------+-------------+----------+
	// | Scenario                               | Target  | Permission Result  | Allowed  | Code    | ActiveUsers | PrivCall |
	// +---------------------------------------+---------+--------------------+----------+---------+-------------+----------+
	// | authorized soft-delete                 | seeded  | success            | true     | success | 0           | 1        |
	// | denied permission                      | seeded  | success            | false    | 403     | 1           | 1        |
	// | permission dependency failure           | seeded  | 503 error          | -        | 503     | 1           | 1        |
	// | missing target user                    | missing | success            | true     | 404     | 1           | 1        |
	// | zero target ID (no permission check)   | zero    | -                  | -        | 400     | 1           | 0        |
	// +---------------------------------------+---------+--------------------+----------+---------+-------------+----------+
	tests := []testCase{
		{
			name: "soft-deletes target user when authorized", target: "seeded",
			permissionSuccess: true, permissionAllowed: true,
			expected: expected{success: true, activeUsers: 0, privilegeCall: 1},
		},
		{
			name: "does not modify users when denied", target: "seeded",
			permissionSuccess: true,
			expected:          expected{code: http.StatusForbidden, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "propagates permission dependency failure", target: "seeded",
			permissionErrorCode: http.StatusServiceUnavailable,
			expected:            expected{code: http.StatusServiceUnavailable, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "returns not found for missing target", target: "missing",
			permissionSuccess: true, permissionAllowed: true,
			expected: expected{code: http.StatusNotFound, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "rejects zero target ID without checking permission", target: "zero",
			expected: expected{code: http.StatusBadRequest, activeUsers: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				if tt.permissionSuccess {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: tt.permissionAllowed}
				}
				return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: tt.permissionErrorCode, ErrorMessage: "permission unavailable"}}
			}
			var userID int
			switch tt.target {
			case "seeded":
				userID = int(seededUserID)
			case "missing":
				userID = missingUserID
			case "zero":
				userID = 0
			}
			got := f.service.DeleteUser(context.Background(), &userSvc.AdminDeleteUserIn{
				Trace: testTrace(), AdminPrivilegeGroup: "Root", UserId: userID,
			})
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
		name         string
		refreshToken string
		expected     expected
	}

	// Seed data: one administrator user, assigned to the "Root"
	// privilege group, with four refresh tokens in distinct states.
	// "used-admin" and "sibling-admin" share a family so reuse-revocation
	// can be observed; other tokens use distinct families to avoid
	// cross-case revocation interference.
	//
	// +-----------------+----------------+--------+--------+
	// | Raw Token       | Family         | Expires| Used   |
	// +-----------------+----------------+--------+--------+
	// | old-admin-token | admin-family   | +1h    | no     |
	// | expired-admin   | expired-family | -1h    | no     |
	// | used-admin      | reused-family  | +1h    | yes    |
	// | sibling-admin   | reused-family  | +1h    | no     |
	// +-----------------+----------------+--------+--------+
	var seededUserID uint
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		user := seedUser(t, f.db, withRole("Admin"))
		seededUserID = user.ID
		seedUserPrivilegeGroup(t, f.db, user.ID, "Root")
		seedRefreshToken(t, f.db, user.ID, "old-admin-token", "admin-family", time.Now().Add(time.Hour), nil)
		seedRefreshToken(t, f.db, user.ID, "expired-admin", "expired-family", time.Now().Add(-time.Hour), nil)
		usedAt := time.Now().Add(-time.Minute)
		seedRefreshToken(t, f.db, user.ID, "used-admin", "reused-family", time.Now().Add(time.Hour), &usedAt)
		seedRefreshToken(t, f.db, user.ID, "sibling-admin", "reused-family", time.Now().Add(time.Hour), nil)
	}

	// Test cases:
	//
	// +----------------------------------+------------------+--------+--------+
	// | Scenario                         | Refresh Token    | Code   | Tokens |
	// +----------------------------------+------------------+--------+--------+
	// | rotate valid admin refresh token  | old-admin-token  | 200    | 5      |
	// | unknown token                     | missing          | 404    | 4      |
	// | expired token                      | expired-admin   | 403    | 4      |
	// | reused token (family revoked)      | used-admin      | 403    | 4      |
	// +----------------------------------+------------------+--------+--------+
	tests := []testCase{
		{name: "rotates administrator refresh token", refreshToken: "old-admin-token", expected: expected{success: true, tokens: 5}},
		{name: "returns not found for unknown token", refreshToken: "missing", expected: expected{code: http.StatusNotFound, tokens: 4}},
		{name: "rejects expired token", refreshToken: "expired-admin", expected: expected{code: http.StatusForbidden, tokens: 4}},
		{name: "revokes family when used token is reused", refreshToken: "used-admin", expected: expected{code: http.StatusForbidden, tokens: 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			got := f.service.Refresh(context.Background(), &userSvc.AdminRefreshIn{
				Trace: testTrace(), AdminUserId: int(seededUserID), RefreshToken: tt.refreshToken,
			})
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
		name                string
		useSeededID         bool
		oldPassword         string
		newPassword         string
		permissionSuccess   bool
		permissionErrorCode int
		expected            expected
	}

	// Seed data: one administrator user with a known old password and
	// one active refresh token.
	//
	// +-------+---------------+-----------------+------------------+
	// | Role  | Old Password  | Refresh Token   | Family           |
	// +-------+---------------+-----------------+------------------+
	// | Admin | old-password  | previous-token  | previous-family  |
	// +-------+---------------+-----------------+------------------+
	var seededUserID uint
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		user := seedUser(t, f.db, withRole("Admin"), withPasswordHash(hashPassword(t, "old-password")))
		seededUserID = user.ID
		seedRefreshToken(t, f.db, user.ID, "previous-token", "previous-family", time.Now().Add(time.Hour), nil)
	}

	// Test cases:
	//
	// +-----------------------------------------+-------------+-------------+--------------+--------------------+---------+-----------------+
	// | Scenario                                 | UseSeededID | OldPassword | NewPassword  | Permission Result  | Code    | StoredPassword  |
	// +-----------------------------------------+-------------+-------------+--------------+--------------------+---------+-----------------+
	// | valid old password                       | yes         | old-password| new-password | success            | success | new-password    |
	// | incorrect old password                   | yes         | wrong       | new-password | (not called)       | 400     | old-password    |
	// | permission lookup failure (rolls back)   | yes         | old-password| new-password | 503 error          | 503     | old-password    |
	// | missing required ID                      | no          | old         | new          | (not called)       | 400     |                 |
	// +-----------------------------------------+-------------+-------------+--------------+--------------------+---------+-----------------+
	tests := []testCase{
		{
			name: "updates password and issues a new refresh token", useSeededID: true,
			oldPassword: "old-password", newPassword: "new-password", permissionSuccess: true,
			expected: expected{success: true, storedPassword: "new-password"},
		},
		{
			name: "rejects incorrect old password without changing data", useSeededID: true,
			oldPassword: "wrong", newPassword: "new-password",
			expected: expected{code: http.StatusBadRequest, storedPassword: "old-password"},
		},
		{
			name: "rolls back password update if permission lookup fails", useSeededID: true,
			oldPassword: "old-password", newPassword: "new-password", permissionErrorCode: http.StatusServiceUnavailable,
			expected: expected{code: http.StatusServiceUnavailable, storedPassword: "old-password"},
		},
		{
			name: "validates required ID", useSeededID: false,
			oldPassword: "old", newPassword: "new",
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
				if tt.permissionSuccess {
					return &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, AdminPrivilegeGroup: "Root"}
				}
				return &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: tt.permissionErrorCode, ErrorMessage: "permission unavailable"}}
			}
			userAdminID := 0
			if tt.useSeededID {
				userAdminID = int(seededUserID)
			}
			input := &userSvc.AdminChangePasswordIn{
				Trace: testTrace(), UserAdminId: userAdminID, OldPassword: tt.oldPassword, NewPassword: tt.newPassword,
			}
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
			} else if len(refreshTokens) != 1 {
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
		name                string
		username            string
		privilegeGroup      string
		groupExists         bool
		groupExistsSuccess  bool
		hasPrivilegeSuccess bool
		hasPrivilegeAllowed bool
		assignSuccess       bool
		assignErrorCode     int
		expected            expected
	}

	// Seed data: one user with a username that conflicts with the
	// registration attempts below.
	//
	// +----------+
	// | Username |
	// +----------+
	// | taken    |
	// +----------+
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		seedUser(t, f.db, withUsername("taken"))
	}

	// Test cases:
	//
	// +----------------------------------------------+-----------+-------------+--------------+----------------+---------+-------+-------------------+
	// | Scenario                                      | Username  | GroupExists | HasPrivilege | AssignSuccess  | Code    | Users | Assign/Priv Calls |
	// +----------------------------------------------+-----------+-------------+--------------+----------------+---------+-------+-------------------+
	// | registers admin, assigns group                 | new-admin | true        | allowed      | true           | success | 2     | 1/1               |
	// | privilege group does not exist                  | new-admin | false       | -            | -              | 400     | 1     | 0/0               |
	// | caller lacks registration privilege              | new-admin | true        | denied       | -              | 403     | 1     | 0/1               |
	// | rolls back user on group assignment failure      | new-admin | true        | allowed      | false (409)    | 409     | 1     | 1/1               |
	// | duplicate username                               | taken     | true        | allowed      | -              | 400     | 1     | 0/1               |
	// +----------------------------------------------+-----------+-------------+--------------+----------------+---------+-------+-------------------+
	tests := []testCase{
		{
			name: "registers admin and assigns requested group", username: "new-admin", privilegeGroup: "Auditor",
			groupExistsSuccess: true, groupExists: true, hasPrivilegeSuccess: true, hasPrivilegeAllowed: true, assignSuccess: true,
			expected: expected{success: true, users: 2, assignCalls: 1, privilegeCalls: 1},
		},
		{
			name: "rejects a privilege group that does not exist", username: "new-admin", privilegeGroup: "Missing",
			groupExistsSuccess: true, groupExists: false,
			expected: expected{code: http.StatusBadRequest, users: 1},
		},
		{
			name: "rejects caller without registration privilege", username: "new-admin", privilegeGroup: "Auditor",
			groupExistsSuccess: true, groupExists: true, hasPrivilegeSuccess: true, hasPrivilegeAllowed: false,
			expected: expected{code: http.StatusForbidden, users: 1, privilegeCalls: 1},
		},
		{
			name: "rolls back created user if group assignment fails", username: "new-admin", privilegeGroup: "Auditor",
			groupExistsSuccess: true, groupExists: true, hasPrivilegeSuccess: true, hasPrivilegeAllowed: true,
			assignSuccess: false, assignErrorCode: http.StatusConflict,
			expected: expected{code: http.StatusConflict, users: 1, assignCalls: 1, privilegeCalls: 1},
		},
		{
			name: "rejects duplicate username", username: "taken", privilegeGroup: "Auditor",
			groupExistsSuccess: true, groupExists: true, hasPrivilegeSuccess: true, hasPrivilegeAllowed: true,
			expected: expected{code: http.StatusBadRequest, users: 1, privilegeCalls: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.CheckAdminGroupPrivilegeExistsFunc = func(context.Context, *permissionSvc.CheckAdminGroupPrivilegeExistsIn) *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
				if tt.groupExistsSuccess {
					return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{BaseOut: dto.BaseOut{Success: true}, IsExists: tt.groupExists}
				}
				return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "group lookup failed"}}
			}
			f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				if tt.hasPrivilegeSuccess {
					return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: tt.hasPrivilegeAllowed}
				}
				return &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission unavailable"}}
			}
			f.permissionMock.ChangeAdminPrivilegeRoleFunc = func(context.Context, *permissionSvc.ChangeAdminPrivilegeRoleIn) *permissionSvc.ChangeAdminPrivilegeRoleOut {
				if tt.assignSuccess {
					return &permissionSvc.ChangeAdminPrivilegeRoleOut{BaseOut: dto.BaseOut{Success: true}}
				}
				return &permissionSvc.ChangeAdminPrivilegeRoleOut{BaseOut: dto.BaseOut{ErrorCode: tt.assignErrorCode, ErrorMessage: "group assignment failed"}}
			}
			got := f.service.RegisterNewAdmin(context.Background(), &userSvc.RegisterNewAdminIn{
				Trace: testTrace(), UserAdminId: 1, AdminPrivilegeGroup: "Root",
				Username: tt.username, Email: tt.username + "@example.test", Password: "password",
				FullName: "New Admin", PrivilegeGroup: tt.privilegeGroup,
			})
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
