package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"template-go/base/helpers"
	"template-go/data/model"
	userSvc "template-go/modules/user/svc"
)

func TestUserService_LoginUser(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		username       string
		password       string
		expectedCode   int
		expectedResult int // 0 = success, 1 = unauthorized
	}

	type seedData struct {
		username string
		password string
		role     string
	}

	// Seed data:
	//
	// +------------+-----------------+----------+
	// | Username   | Password        | Role     |
	// +------------+-----------------+----------+
	// | login-user | valid-password  | User     |
	// | admin      | valid-password  | Admin    |
	// +------------+-----------------+----------+
	users := []seedData{
		{username: "login-user", password: "valid-password", role: "User"},
		{username: "admin", password: "valid-password", role: "Admin"},
	}

	seedSharedData := func(t *testing.T, f *userFixture) {
		passwordHash := hashPassword(t, "valid-password")
		for _, u := range users {
			seedUserWithData(t, f.db, userSeedData{
				username:     u.username,
				passwordHash: passwordHash,
				role:         u.role,
			})
		}
	}

	tests := []testCase{
		{
			name:           "issues access and refresh tokens for valid credentials",
			username:       "login-user",
			password:       "valid-password",
			expectedResult: 0,
		},
		{
			name:           "rejects unknown username",
			username:       "missing",
			password:       "valid-password",
			expectedCode:   http.StatusUnauthorized,
			expectedResult: 1,
		},
		{
			name:           "rejects incorrect password",
			username:       "login-user",
			password:       "incorrect",
			expectedCode:   http.StatusUnauthorized,
			expectedResult: 1,
		},
		{
			name:           "does not allow administrator accounts to use public login",
			username:       "admin",
			password:       "valid-password",
			expectedCode:   http.StatusUnauthorized,
			expectedResult: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			seedSharedData(t, f)
			got := f.service.LoginUser(context.Background(), &userSvc.LoginUserIn{
				Trace:    testTrace(),
				Username: tt.username,
				Password: tt.password,
			})
			assertEqual(t, "LoginUser() error code", tt.expectedCode, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.RefreshToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if tt.expectedResult == 0 { // success
				if got.Success == false {
					t.Fatal("LoginUser() expected success")
				}
				if got.Token == "" || got.RefreshToken == "" {
					t.Fatal("LoginUser() returned an empty access or refresh token")
				}
				if count != 1 {
					t.Fatalf("persisted refresh token count = %d, want 1", count)
				}
			} else { // unauthorized
				if got.Success == true {
					t.Fatal("LoginUser() expected failure")
				}
				if count != 0 {
					t.Fatalf("persisted refresh token count = %d after rejected login, want 0", count)
				}
			}
		})
	}
}

func TestUserService_Register(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		users   int64
	}

	type testCase struct {
		name     string
		username string
		email    string
		role     string
		expected expected
	}

	// Seed data:
	//
	// +--------+-----------------------+----------+
	// | Role   | Username               | Email    |
	// +--------+-----------------------+----------+
	// | Member | taken                 | taken@example.test |
	// +--------+-----------------------+----------+
	seedSharedData := func(t *testing.T, f *userFixture) {
		seedRole(t, f.db, "Member")
		seedUser(t, f.db, withUsername("taken"), withEmail("taken@example.test"))
	}

	// Test cases:
	//
	// +--------------------------+----------------+----------+----------------------+---------+-------+
	// | Scenario                 | Username       | Email    | Role                 | Success | Users |
	// +--------------------------+----------------+----------+----------------------+---------+-------+
	// | existing role            | new-user       | new      | Member               | true    | 2     |
	// | missing role             | new-user       | new      | Missing              | false   | 1     |
	// | duplicate username       | taken          | new      | Member               | false   | 1     |
	// | duplicate email          | new-user       | taken    | Member               | false   | 1     |
	// +--------------------------+----------------+----------+----------------------+---------+-------+
	tests := []testCase{
		{
			name: "registers user with existing role", username: "new-user", email: "new-user@example.test", role: "Member",
			expected: expected{success: true, users: 2},
		},
		{
			name: "rejects missing role", username: "new-user", email: "new-user@example.test", role: "Missing",
			expected: expected{code: http.StatusBadRequest, users: 1},
		},
		{
			name: "rejects duplicate username", username: "taken", email: "new-user@example.test", role: "Member",
			expected: expected{code: http.StatusUnprocessableEntity, users: 1},
		},
		{
			name: "rejects duplicate email without persisting a user", username: "new-user", email: "taken@example.test", role: "Member",
			expected: expected{code: http.StatusInternalServerError, users: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			seedSharedData(t, f)
			got := f.service.Register(context.Background(), &userSvc.UserRegisterIn{
				Trace: testTrace(), Username: tt.username, Email: tt.email,
				Password: "password", FullName: "New User", UserRole: tt.role,
			})
			assertEqual(t, "Register() success", tt.expected.success, got.Success)
			assertEqual(t, "Register() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.User{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.users {
				t.Fatalf("persisted user count = %d, want %d", count, tt.expected.users)
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
	}

	type testCase struct {
		name          string
		useSeededUser bool
		expected      expected
	}

	// Seed data: one user is created via seedUser(); its ID is reused
	// across test cases, since GORM primary keys never use ID 0.
	var seededID uint
	seedSharedData := func(t *testing.T, f *userFixture) {
		seededID = seedUser(t, f.db).ID
	}

	// Test cases:
	//
	// +---------------------+----------------+---------+
	// | Scenario            | Condition      | Code    |
	// +---------------------+----------------+---------+
	// | existing user       | seeded user    | 200     |
	// | missing user         | unknown ID     | 404     |
	// +---------------------+----------------+---------+
	tests := []testCase{
		{name: "returns existing user", useSeededUser: true, expected: expected{success: true}},
		{name: "returns not found for missing user", useSeededUser: false, expected: expected{code: http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			seedSharedData(t, f)
			id := uint(0)
			if tt.useSeededUser {
				id = seededID
			}
			got := f.service.GetUser(context.Background(), &userSvc.GetUserIn{Trace: testTrace(), UserID: id})
			assertEqual(t, "GetUser() success", tt.expected.success, got.Success)
			assertEqual(t, "GetUser() error code", tt.expected.code, got.ErrorCode)
			if tt.expected.success {
				assertEqual(t, "GetUser() user ID", id, got.User.ID)
			}
		})
	}
}

func TestUserService_InsertFcmToken(t *testing.T) {
	t.Parallel()

	type expected struct {
		success bool
		code    int
		tokens  int64
	}

	type testCase struct {
		name          string
		useSeededUser bool
		expected      expected
	}

	// Seed data: one user is created via seedUser(); its ID is reused
	// across test cases.
	var seededID uint
	seedSharedData := func(t *testing.T, f *userFixture) {
		seededID = seedUser(t, f.db).ID
	}

	// Test cases:
	//
	// +---------------------+----------------+---------+--------+
	// | Scenario            | Condition      | Code    | Tokens |
	// +---------------------+----------------+---------+--------+
	// | existing user       | seeded user    | 200     | 1      |
	// | missing user         | unknown ID     | 404     | 0      |
	// +---------------------+----------------+---------+--------+
	tests := []testCase{
		{name: "inserts token for existing user", useSeededUser: true, expected: expected{success: true, tokens: 1}},
		{name: "rejects missing user", useSeededUser: false, expected: expected{code: http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			seedSharedData(t, f)
			id := uint(0)
			if tt.useSeededUser {
				id = seededID
			}
			got := f.service.InsertFcmToken(context.Background(), &userSvc.InsertFcmTokenIn{Trace: testTrace(), UserID: id, Token: "fcm-token"})
			assertEqual(t, "InsertFcmToken() success", tt.expected.success, got.Success)
			assertEqual(t, "InsertFcmToken() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.FcmToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.tokens {
				t.Fatalf("persisted FCM token count = %d, want %d", count, tt.expected.tokens)
			}
		})
	}
}

func TestUserService_DeleteFcmTokenBulk(t *testing.T) {
	t.Parallel()

	type expected struct {
		success   bool
		remaining int64
	}

	type testCase struct {
		name     string
		tokens   []string
		expected expected
	}

	// Seed data: one user with two FCM tokens.
	//
	// +------------+
	// | Token      |
	// +------------+
	// | remove-me  |
	// | keep-me    |
	// +------------+
	seedSharedData := func(t *testing.T, f *userFixture) {
		user := seedUser(t, f.db)
		seedFCMTokens(t, f.db,
			fcmTokenRow{UserID: user.ID, Token: "remove-me"},
			fcmTokenRow{UserID: user.ID, Token: "keep-me"},
		)
	}

	// Test cases:
	//
	// +--------------------------+---------------------------+-----------+
	// | Scenario                 | Tokens requested           | Remaining |
	// +--------------------------+---------------------------+-----------+
	// | matching tokens          | remove-me                 | 1         |
	// | empty token set          | (none)                     | 2         |
	// +--------------------------+---------------------------+-----------+
	tests := []testCase{
		{name: "deletes matching tokens only", tokens: []string{"remove-me"}, expected: expected{success: true, remaining: 1}},
		{name: "succeeds with empty token set", tokens: nil, expected: expected{success: true, remaining: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			seedSharedData(t, f)
			got := f.service.DeleteFcmTokenBulk(context.Background(), &userSvc.DeleteFcmTokenBulkIn{Trace: testTrace(), Tokens: tt.tokens})
			assertEqual(t, "DeleteFcmTokenBulk() success", tt.expected.success, got.Success)
			var count int64
			if err := f.db.Model(&model.FcmToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.expected.remaining {
				t.Fatalf("remaining FCM token count = %d, want %d", count, tt.expected.remaining)
			}
		})
	}
}

func TestUserService_GetUserFcmToken(t *testing.T) {
	t.Parallel()

	type expected struct {
		success    bool
		code       int
		tokenCount int
	}

	type testCase struct {
		name      string
		userIndex int // -1 = missing user (sentinel ID), 0/1 = index into seeded users
		expected  expected
	}

	// Seed data: two users.
	//
	// +-------+------------+
	// | Index | FCM Tokens |
	// +-------+------------+
	// | 0     | token-1    |
	// | 1     | (none)     |
	// +-------+------------+
	var seededIDs [2]uint
	seedSharedData := func(t *testing.T, f *userFixture) {
		withToken := seedUser(t, f.db)
		seedFCMToken(t, f.db, withToken.ID, "token-1")
		withoutToken := seedUser(t, f.db)
		seededIDs = [2]uint{withToken.ID, withoutToken.ID}
	}

	// Test cases:
	//
	// +---------------------+------------------------+--------+-------+
	// | Scenario            | User                     | Code   | Count |
	// +---------------------+------------------------+--------+-------+
	// | has tokens          | index 0 (token-1)        | 200    | 1     |
	// | no tokens           | index 1 (no tokens)      | 200    | 0     |
	// | missing user        | sentinel ID              | 404    | 0     |
	// +---------------------+------------------------+--------+-------+
	tests := []testCase{
		{name: "returns a user's tokens", userIndex: 0, expected: expected{success: true, tokenCount: 1}},
		{name: "returns an empty list when user has no tokens", userIndex: 1, expected: expected{success: true}},
		{name: "rejects missing user", userIndex: -1, expected: expected{code: http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			seedSharedData(t, f)
			id := uint(0)
			if tt.userIndex >= 0 {
				id = seededIDs[tt.userIndex]
			}
			got := f.service.GetUserFcmToken(context.Background(), &userSvc.GetUserFcmTokenIn{Trace: testTrace(), UserID: id})
			assertEqual(t, "GetUserFcmToken() success", tt.expected.success, got.Success)
			assertEqual(t, "GetUserFcmToken() error code", tt.expected.code, got.ErrorCode)
			if len(got.FcmTokens) != tt.expected.tokenCount {
				t.Fatalf("returned FCM token count = %d, want %d", len(got.FcmTokens), tt.expected.tokenCount)
			}
		})
	}
}

func TestUserService_Refresh(t *testing.T) {
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

	// Seed data: one user with four refresh tokens in distinct states.
	// "reused" and "sibling" share a family so reuse-revocation can be
	// observed; other tokens use distinct families to avoid cross-case
	// revocation interference.
	//
	// +-------------+----------------+------------+--------+
	// | Raw Token   | Family         | Expires     | Used   |
	// +-------------+----------------+------------+--------+
	// | old-token   | family-valid   | +1h         | no     |
	// | expired     | family-expired | -1h         | no     |
	// | reused      | family-reused  | +1h         | yes    |
	// | sibling     | family-reused  | +1h         | no     |
	// +-------------+----------------+------------+--------+
	var seededUserID uint
	seedSharedData := func(t *testing.T, f *userFixture) {
		// Users table: one user.
		user := seedUser(t, f.db)
		seededUserID = user.ID

		// RefreshTokens table (see above):
		usedAt := time.Now().Add(-time.Minute)
		seedRefreshTokens(t, f.db,
			refreshTokenRow{UserID: user.ID, Token: "old-token", Family: "family-valid", ExpireAt: time.Now().Add(time.Hour)},
			refreshTokenRow{UserID: user.ID, Token: "expired", Family: "family-expired", ExpireAt: time.Now().Add(-time.Hour)},
			refreshTokenRow{UserID: user.ID, Token: "reused", Family: "family-reused", ExpireAt: time.Now().Add(time.Hour), UsedAt: &usedAt},
			refreshTokenRow{UserID: user.ID, Token: "sibling", Family: "family-reused", ExpireAt: time.Now().Add(time.Hour)},
		)
	}

	// Test cases:
	//
	// +--------------------------------+---------------------------+--------+--------+
	// | Scenario                       | Refresh Token               | Code   | Tokens |
	// +--------------------------------+---------------------------+--------+--------+
	// | valid refresh token            | old-token                   | 200    | 5      |
	// | missing refresh token          | missing                     | 404    | 4      |
	// | expired refresh token          | expired                     | 403    | 4      |
	// | reused consumed token          | reused                      | 403    | 4      |
	// +--------------------------------+---------------------------+--------+--------+
	tests := []testCase{
		{name: "rotates a valid refresh token", refreshToken: "old-token", expected: expected{success: true, tokens: 5}},
		{name: "rejects missing refresh token", refreshToken: "missing", expected: expected{code: http.StatusNotFound, tokens: 4}},
		{name: "rejects expired refresh token", refreshToken: "expired", expected: expected{code: http.StatusForbidden, tokens: 4}},
		{name: "revokes a token family when a consumed token is reused", refreshToken: "reused", expected: expected{code: http.StatusForbidden, tokens: 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			seedSharedData(t, f)
			got := f.service.Refresh(context.Background(), &userSvc.RefreshIn{
				Trace: testTrace(), UserId: int(seededUserID), RefreshToken: tt.refreshToken,
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
			if tt.expected.success {
				if got.RefreshToken == "" || got.AccessToken == "" {
					t.Fatal("Refresh() returned an empty token")
				}
				var old model.RefreshToken
				if err := f.db.Where("token_hash = ?", helpers.HashToken("old-token")).First(&old).Error; err != nil {
					t.Fatal(err)
				}
				if old.UsedAt == nil {
					t.Fatal("consumed refresh token was not marked as used")
				}
			}
			if tt.name == "revokes a token family when a consumed token is reused" {
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

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := helpers.HashPassword(password)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	return hash
}
