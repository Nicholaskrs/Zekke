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
		name          string
		input         *userSvc.LoginUserAdminIn
		permissionOut *permissionSvc.GetAdminGroupPrivilegeOut
		expected      expected
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
	// +---------------------------------+----------+----------+------------------------------------------------+---------+--------+
	// | Scenario                        | Username | Password | GetAdminGroupPrivilegeOut                       | Code    | Tokens |
	// +---------------------------------+----------+----------+------------------------------------------------+---------+--------+
	// | valid administrator credentials | admin    | correct  | {Success: true, AdminPrivilegeGroup: "Root"}    | success | 1      |
	// | non-administrator account       | regular  | correct  | (not called)                                    | 401     | 0      |
	// | permission service failure      | admin    | correct  | {ErrorCode: 503}                                | 503     | 0      |
	// +---------------------------------+----------+----------+------------------------------------------------+---------+--------+
	tests := []testCase{
		{
			name:  "issues tokens for valid administrator credentials",
			input: &userSvc.LoginUserAdminIn{Username: "admin", Password: "correct"},
			permissionOut: &permissionSvc.GetAdminGroupPrivilegeOut{
				BaseOut: dto.BaseOut{Success: true}, AdminPrivilegeGroup: "Root",
			},
			expected: expected{success: true, tokens: 1},
		},
		{
			name:     "rejects non-administrator account",
			input:    &userSvc.LoginUserAdminIn{Username: "regular", Password: "correct"},
			expected: expected{code: http.StatusUnauthorized},
		},
		{
			name:  "returns permission service failure",
			input: &userSvc.LoginUserAdminIn{Username: "admin", Password: "correct"},
			permissionOut: &permissionSvc.GetAdminGroupPrivilegeOut{
				BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission unavailable"},
			},
			expected: expected{code: http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
				return tt.permissionOut
			}
			input := tt.input
			input.Trace = testTrace()
			got := f.service.LoginUserAdmin(context.Background(), input)
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
		name            string
		hasPrivilegeOut *permissionSvc.HasPrivilegeOut
		expected        expected
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
		deletedAt := time.Now()
		seedUsers(t, f.db,
			userRow{Username: "active"},
			userRow{Username: "deleted", DeletedAt: &deletedAt},
		)
	}

	// Test cases:
	//
	// +--------------------------+------------------------------------------------------------+---------+-------+-----------+
	// | Scenario                 | HasPrivilegeOut                                             | Code    | Users | TotalRows |
	// +--------------------------+------------------------------------------------------------+---------+-------+-----------+
	// | active users, allowed    | {Success: true, IsAllowed: true}                            | success | 1     | 1         |
	// | permission denied        | {Success: true, IsAllowed: false}                           | 403     | 0     | 0         |
	// | permission service error | {ErrorCode: 503}                                            | 503     | 0     | 0         |
	// +--------------------------+------------------------------------------------------------+---------+-------+-----------+
	tests := []testCase{
		{
			name:            "returns active users and count",
			hasPrivilegeOut: &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true},
			expected:        expected{success: true, users: 1, totalRows: 1},
		},
		{
			name:            "rejects denied permission",
			hasPrivilegeOut: &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: false},
			expected:        expected{code: http.StatusForbidden},
		},
		{
			name:            "propagates permission service error",
			hasPrivilegeOut: &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission failure"}},
			expected:        expected{code: http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				return tt.hasPrivilegeOut
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
		name            string
		userID          int
		hasPrivilegeOut *permissionSvc.HasPrivilegeOut
		expected        expected
	}

	// Seed data: one user with one active refresh token. The seeded
	// user is always the first row in a freshly migrated, isolated
	// test database, so its ID is deterministically 1.
	//
	// +----+--------------------+--------------+--------+
	// | ID | Raw Token          | Family       | Expires|
	// +----+--------------------+--------------+--------+
	// | 1  | active-user-token  | user-session | +1h    |
	// +----+--------------------+--------------+--------+
	//
	// A sentinel ID (999999) that can never collide with a seeded row
	// is used for the "missing target" scenario.
	const (
		seededUserID  = 1
		missingUserID = 999999
	)
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		user := seedUser(t, f.db)
		seedRefreshToken(t, f.db, user.ID, "active-user-token", "user-session", time.Now().Add(time.Hour), nil)
	}

	// Test cases:
	//
	// +---------------------------------------+---------------+----------------------------------------+---------+-------------+----------+
	// | Scenario                               | UserID        | HasPrivilegeOut                        | Code    | ActiveUsers | PrivCall |
	// +---------------------------------------+---------------+----------------------------------------+---------+-------------+----------+
	// | authorized soft-delete                 | seededUserID  | {Success: true, IsAllowed: true}       | success | 0           | 1        |
	// | denied permission                      | seededUserID  | {Success: true, IsAllowed: false}      | 403     | 1           | 1        |
	// | permission dependency failure           | seededUserID  | {ErrorCode: 503}                       | 503     | 1           | 1        |
	// | missing target user                    | missingUserID | {Success: true, IsAllowed: true}       | 404     | 1           | 1        |
	// | zero target ID (no permission check)   | 0             | (not called)                           | 400     | 1           | 0        |
	// +---------------------------------------+---------------+----------------------------------------+---------+-------------+----------+
	tests := []testCase{
		{
			name: "soft-deletes target user when authorized", userID: seededUserID,
			hasPrivilegeOut: &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true},
			expected:        expected{success: true, activeUsers: 0, privilegeCall: 1},
		},
		{
			name: "does not modify users when denied", userID: seededUserID,
			hasPrivilegeOut: &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: false},
			expected:        expected{code: http.StatusForbidden, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "propagates permission dependency failure", userID: seededUserID,
			hasPrivilegeOut: &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission unavailable"}},
			expected:        expected{code: http.StatusServiceUnavailable, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "returns not found for missing target", userID: missingUserID,
			hasPrivilegeOut: &permissionSvc.HasPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, IsAllowed: true},
			expected:        expected{code: http.StatusNotFound, activeUsers: 1, privilegeCall: 1},
		},
		{
			name: "rejects zero target ID without checking permission", userID: 0,
			expected: expected{code: http.StatusBadRequest, activeUsers: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.HasPrivilegeFunc = func(context.Context, *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				return tt.hasPrivilegeOut
			}
			got := f.service.DeleteUser(context.Background(), &userSvc.AdminDeleteUserIn{
				Trace: testTrace(), AdminPrivilegeGroup: "Root", UserId: tt.userID,
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
		// Users table:
		// +------+--------+
		// | Role | Count  |
		// +------+--------+
		// | Admin| 1      |
		// +------+--------+
		user := seedUser(t, f.db, withRole("Admin"))
		seededUserID = user.ID

		// UserPrivilegeGroups table:
		// +---------------+
		// | PrivilegeGroup|
		// +---------------+
		// | Root          |
		// +---------------+
		seedUserPrivilegeGroup(t, f.db, user.ID, "Root")

		// RefreshTokens table (see above):
		usedAt := time.Now().Add(-time.Minute)
		seedRefreshTokens(t, f.db,
			refreshTokenRow{UserID: user.ID, Token: "old-admin-token", Family: "admin-family", ExpireAt: time.Now().Add(time.Hour)},
			refreshTokenRow{UserID: user.ID, Token: "expired-admin", Family: "expired-family", ExpireAt: time.Now().Add(-time.Hour)},
			refreshTokenRow{UserID: user.ID, Token: "used-admin", Family: "reused-family", ExpireAt: time.Now().Add(time.Hour), UsedAt: &usedAt},
			refreshTokenRow{UserID: user.ID, Token: "sibling-admin", Family: "reused-family", ExpireAt: time.Now().Add(time.Hour)},
		)
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
		name          string
		input         *userSvc.AdminChangePasswordIn
		permissionOut *permissionSvc.GetAdminGroupPrivilegeOut
		expected      expected
	}

	// Seed data: one administrator user with a known old password and
	// one active refresh token. The seeded user is always the first
	// row in a freshly migrated, isolated test database, so its ID is
	// deterministically 1.
	//
	// +----+-------+---------------+-----------------+------------------+
	// | ID | Role  | Old Password  | Refresh Token   | Family           |
	// +----+-------+---------------+-----------------+------------------+
	// | 1  | Admin | old-password  | previous-token  | previous-family  |
	// +----+-------+---------------+-----------------+------------------+
	const seededAdminID = 1
	seedSharedData := func(t *testing.T, f *userAdminFixture) {
		user := seedUser(t, f.db, withRole("Admin"), withPasswordHash(hashPassword(t, "old-password")))
		seedRefreshToken(t, f.db, user.ID, "previous-token", "previous-family", time.Now().Add(time.Hour), nil)
	}

	// Test cases:
	//
	// +-----------------------------------------+---------------+------------------------------------------+----------------------------+---------+-----------------+
	// | Scenario                                 | UserAdminId   | Input (OldPassword/NewPassword)           | GetAdminGroupPrivilegeOut  | Code    | StoredPassword  |
	// +-----------------------------------------+---------------+------------------------------------------+----------------------------+---------+-----------------+
	// | valid old password                       | seededAdminID | old-password/new-password                 | {Success: true}            | success | new-password    |
	// | incorrect old password                   | seededAdminID | wrong/new-password                        | (not called)               | 400     | old-password    |
	// | permission lookup failure (rolls back)   | seededAdminID | old-password/new-password                 | {ErrorCode: 503}           | 503     | old-password    |
	// | missing required ID                      | 0             | old/new                                   | (not called)               | 400     |                 |
	// +-----------------------------------------+---------------+------------------------------------------+----------------------------+---------+-----------------+
	tests := []testCase{
		{
			name:          "updates password and issues a new refresh token",
			input:         &userSvc.AdminChangePasswordIn{UserAdminId: seededAdminID, OldPassword: "old-password", NewPassword: "new-password"},
			permissionOut: &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{Success: true}, AdminPrivilegeGroup: "Root"},
			expected:      expected{success: true, storedPassword: "new-password"},
		},
		{
			name:     "rejects incorrect old password without changing data",
			input:    &userSvc.AdminChangePasswordIn{UserAdminId: seededAdminID, OldPassword: "wrong", NewPassword: "new-password"},
			expected: expected{code: http.StatusBadRequest, storedPassword: "old-password"},
		},
		{
			name:          "rolls back password update if permission lookup fails",
			input:         &userSvc.AdminChangePasswordIn{UserAdminId: seededAdminID, OldPassword: "old-password", NewPassword: "new-password"},
			permissionOut: &permissionSvc.GetAdminGroupPrivilegeOut{BaseOut: dto.BaseOut{ErrorCode: http.StatusServiceUnavailable, ErrorMessage: "permission unavailable"}},
			expected:      expected{code: http.StatusServiceUnavailable, storedPassword: "old-password"},
		},
		{
			name:     "validates required ID",
			input:    &userSvc.AdminChangePasswordIn{OldPassword: "old", NewPassword: "new"},
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserAdminFixture(t)
			seedSharedData(t, f)
			f.permissionMock.GetAdminGroupPrivilegeFunc = func(context.Context, *permissionSvc.GetAdminGroupPrivilegeIn) *permissionSvc.GetAdminGroupPrivilegeOut {
				return tt.permissionOut
			}
			input := tt.input
			input.Trace = testTrace()
			got := f.service.ChangeAdminPassword(context.Background(), input)
			assertEqual(t, "ChangeAdminPassword() success", tt.expected.success, got.Success)
			assertEqual(t, "ChangeAdminPassword() error code", tt.expected.code, got.ErrorCode)

			var user model.User
			err := f.db.First(&user, input.UserAdminId).Error
			if err == nil && tt.expected.storedPassword != "" && !helpers.VerifyPassword(tt.expected.storedPassword, user.Password) {
				t.Fatalf("stored password does not verify against expected password %q", tt.expected.storedPassword)
			}

			var previousToken model.RefreshToken
			if err := f.db.Where("token_hash = ?", helpers.HashToken("previous-token")).First(&previousToken).Error; err != nil {
				t.Fatal(err)
			}

			if !tt.expected.success {
				// ensure the refresh token was not revoked and no replacement was issued
				if previousToken.RevokedAt != nil {
					t.Fatal("failed password change revoked the existing refresh token")
				}
				if got.RefreshToken != "" {
					t.Fatal("failed password change issued a replacement refresh token")
				}
				return
			}

			// ensure the old refresh token was revoked and a replacement was issued
			if previousToken.RevokedAt == nil {
				t.Fatal("refresh token was not revoked after password change")
			}
			if got.RefreshToken == "" {
				t.Fatal("no replacement refresh token issued after password change")
			}
			var newToken model.RefreshToken
			if err := f.db.Where("token_hash = ?", helpers.HashToken(got.RefreshToken)).First(&newToken).Error; err != nil {
				t.Fatalf("replacement refresh token was not persisted: %v", err)
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
