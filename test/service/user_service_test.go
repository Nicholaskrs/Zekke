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

	type expected struct {
		success bool
		code    int
	}
	type testCase struct {
		name     string
		input    *userSvc.LoginUserIn
		expected expected
	}

	db, cfg := newTestDatabase(t)
	f := newUserFixtureWithDB(db, cfg)
	passwordHash := hashPassword(t, "valid-password")
	loginUser := seedUser(t, db, withUsername("login-user"), withPasswordHash(passwordHash))
	seedUser(t, db, withUsername("admin"), withRole("Admin"), withPasswordHash(passwordHash))

	tests := []testCase{
		{
			name:     "issues access and refresh tokens for valid credentials",
			input:    &userSvc.LoginUserIn{Trace: testTrace(), Username: "login-user", Password: "valid-password"},
			expected: expected{success: true},
		},
		{
			name:     "rejects unknown username",
			input:    &userSvc.LoginUserIn{Trace: testTrace(), Username: "missing", Password: "valid-password"},
			expected: expected{code: http.StatusUnauthorized},
		},
		{
			name:     "rejects incorrect password",
			input:    &userSvc.LoginUserIn{Trace: testTrace(), Username: "login-user", Password: "incorrect"},
			expected: expected{code: http.StatusUnauthorized},
		},
		{
			name:     "does not allow administrator accounts to use public login",
			input:    &userSvc.LoginUserIn{Trace: testTrace(), Username: "admin", Password: "valid-password"},
			expected: expected{code: http.StatusUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				if err := f.db.Where("user_id = ?", loginUser.ID).Delete(&model.RefreshToken{}).Error; err != nil {
					t.Errorf("clean up refresh tokens: %v", err)
				}
			})
			got := f.service.LoginUser(context.Background(), tt.input)
			assertEqual(t, "LoginUser() success", tt.expected.success, got.Success)
			assertEqual(t, "LoginUser() error code", tt.expected.code, got.ErrorCode)
			var count int64
			if err := f.db.Model(&model.RefreshToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if tt.expected.success {
				if got.Token == "" || got.RefreshToken == "" {
					t.Fatal("LoginUser() returned an empty access or refresh token")
				}
				if count != 1 {
					t.Fatalf("persisted refresh token count = %d, want 1", count)
				}
			} else if count != 0 {
				t.Fatalf("persisted refresh token count = %d after rejected login, want 0", count)
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
		setup    func(t *testing.T, f *userFixture) *userSvc.UserRegisterIn
		expected expected
	}

	tests := []testCase{
		{
			name: "registers user with existing role",
			setup: func(t *testing.T, f *userFixture) *userSvc.UserRegisterIn {
				seedRole(t, f.db, "Member")
				return &userSvc.UserRegisterIn{
					Trace: testTrace(), Username: "new-user", Email: "new-user@example.test",
					Password: "password", FullName: "New User", UserRole: "Member",
				}
			},
			expected: expected{success: true, users: 1},
		},
		{
			name: "rejects missing role",
			setup: func(t *testing.T, _ *userFixture) *userSvc.UserRegisterIn {
				return &userSvc.UserRegisterIn{
					Trace: testTrace(), Username: "new-user", Email: "new-user@example.test",
					Password: "password", FullName: "New User", UserRole: "Missing",
				}
			},
			expected: expected{code: http.StatusBadRequest},
		},
		{
			name: "rejects duplicate username",
			setup: func(t *testing.T, f *userFixture) *userSvc.UserRegisterIn {
				seedRole(t, f.db, "Member")
				seedUser(t, f.db, withUsername("taken"))
				return &userSvc.UserRegisterIn{
					Trace: testTrace(), Username: "taken", Email: "new-user@example.test",
					Password: "password", FullName: "New User", UserRole: "Member",
				}
			},
			expected: expected{code: http.StatusUnprocessableEntity, users: 1},
		},
		{
			name: "rejects duplicate email without persisting a user",
			setup: func(t *testing.T, f *userFixture) *userSvc.UserRegisterIn {
				seedRole(t, f.db, "Member")
				seedUser(t, f.db, withEmail("taken@example.test"))
				return &userSvc.UserRegisterIn{
					Trace: testTrace(), Username: "new-user", Email: "taken@example.test",
					Password: "password", FullName: "New User", UserRole: "Member",
				}
			},
			expected: expected{code: http.StatusInternalServerError, users: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			got := f.service.Register(context.Background(), tt.setup(t, f))
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
		name     string
		input    *userSvc.GetUserIn
		expected expected
	}

	f := newUserFixture(t)
	user := seedUser(t, f.db)

	tests := []testCase{
		{
			name:     "returns existing user",
			input:    &userSvc.GetUserIn{Trace: testTrace(), UserID: user.ID},
			expected: expected{success: true},
		},
		{
			name:     "returns not found for missing user",
			input:    &userSvc.GetUserIn{Trace: testTrace(), UserID: 99},
			expected: expected{code: http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.service.GetUser(context.Background(), tt.input)
			assertEqual(t, "GetUser() success", tt.expected.success, got.Success)
			assertEqual(t, "GetUser() error code", tt.expected.code, got.ErrorCode)
			if tt.expected.success {
				assertEqual(t, "GetUser() user ID", tt.input.UserID, got.User.ID)
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
		name     string
		setup    func(t *testing.T, f *userFixture) *userSvc.InsertFcmTokenIn
		expected expected
	}

	tests := []testCase{
		{
			name: "inserts token for existing user",
			setup: func(t *testing.T, f *userFixture) *userSvc.InsertFcmTokenIn {
				user := seedUser(t, f.db)
				return &userSvc.InsertFcmTokenIn{Trace: testTrace(), UserID: user.ID, Token: "fcm-token"}
			},
			expected: expected{success: true, tokens: 1},
		},
		{
			name: "rejects missing user",
			setup: func(t *testing.T, _ *userFixture) *userSvc.InsertFcmTokenIn {
				return &userSvc.InsertFcmTokenIn{Trace: testTrace(), UserID: 42, Token: "fcm-token"}
			},
			expected: expected{code: http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			got := f.service.InsertFcmToken(context.Background(), tt.setup(t, f))
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
		setup    func(t *testing.T, f *userFixture) []string
		expected expected
	}

	tests := []testCase{
		{
			name: "deletes matching tokens only",
			setup: func(t *testing.T, f *userFixture) []string {
				user := seedUser(t, f.db)
				seedFCMToken(t, f.db, user.ID, "remove-me")
				seedFCMToken(t, f.db, user.ID, "keep-me")
				return []string{"remove-me"}
			},
			expected: expected{success: true, remaining: 1},
		},
		{
			name:  "succeeds with empty token set",
			setup: func(t *testing.T, _ *userFixture) []string { return nil }, expected: expected{success: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			input := &userSvc.DeleteFcmTokenBulkIn{Trace: testTrace(), Tokens: tt.setup(t, f)}
			got := f.service.DeleteFcmTokenBulk(context.Background(), input)
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
		name     string
		input    *userSvc.GetUserFcmTokenIn
		expected expected
	}

	f := newUserFixture(t)
	withToken := seedUser(t, f.db)
	withoutToken := seedUser(t, f.db)
	seedFCMToken(t, f.db, withToken.ID, "token-1")

	tests := []testCase{
		{
			name:     "returns a user's tokens",
			input:    &userSvc.GetUserFcmTokenIn{Trace: testTrace(), UserID: withToken.ID},
			expected: expected{success: true, tokenCount: 1},
		},
		{
			name:     "returns an empty list when user has no tokens",
			input:    &userSvc.GetUserFcmTokenIn{Trace: testTrace(), UserID: withoutToken.ID},
			expected: expected{success: true},
		},
		{
			name:     "rejects missing user",
			input:    &userSvc.GetUserFcmTokenIn{Trace: testTrace(), UserID: 42},
			expected: expected{code: http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.service.GetUserFcmToken(context.Background(), tt.input)
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
		name     string
		setup    func(t *testing.T, f *userFixture) *userSvc.RefreshIn
		expected expected
	}

	tests := []testCase{
		{
			name: "rotates a valid refresh token",
			setup: func(t *testing.T, f *userFixture) *userSvc.RefreshIn {
				user := seedUser(t, f.db)
				seedRefreshToken(t, f.db, user.ID, "old-token", "family", time.Now().Add(time.Hour), nil)
				return &userSvc.RefreshIn{Trace: testTrace(), UserId: int(user.ID), RefreshToken: "old-token"}
			},
			expected: expected{success: true, tokens: 2},
		},
		{
			name: "rejects missing refresh token",
			setup: func(t *testing.T, f *userFixture) *userSvc.RefreshIn {
				user := seedUser(t, f.db)
				return &userSvc.RefreshIn{Trace: testTrace(), UserId: int(user.ID), RefreshToken: "missing"}
			},
			expected: expected{code: http.StatusNotFound},
		},
		{
			name: "rejects expired refresh token",
			setup: func(t *testing.T, f *userFixture) *userSvc.RefreshIn {
				user := seedUser(t, f.db)
				seedRefreshToken(t, f.db, user.ID, "expired", "family", time.Now().Add(-time.Hour), nil)
				return &userSvc.RefreshIn{Trace: testTrace(), UserId: int(user.ID), RefreshToken: "expired"}
			},
			expected: expected{code: http.StatusForbidden, tokens: 1},
		},
		{
			name: "revokes a token family when a consumed token is reused",
			setup: func(t *testing.T, f *userFixture) *userSvc.RefreshIn {
				user := seedUser(t, f.db)
				usedAt := time.Now().Add(-time.Minute)
				seedRefreshToken(t, f.db, user.ID, "reused", "family", time.Now().Add(time.Hour), &usedAt)
				seedRefreshToken(t, f.db, user.ID, "sibling", "family", time.Now().Add(time.Hour), nil)
				return &userSvc.RefreshIn{Trace: testTrace(), UserId: int(user.ID), RefreshToken: "reused"}
			},
			expected: expected{code: http.StatusForbidden, tokens: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
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
