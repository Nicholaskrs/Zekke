package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"template-go/base/helpers"
	userSvc "template-go/modules/user/svc"
	"template-go/test"
	"template-go/test/fixture"
)

func TestUserService_LoginUser(t *testing.T) {
	t.Parallel()

	type input struct {
		Username string
		Password string
	}
	type expected struct {
		success    bool
		errorCode  int
		tokenCount int
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}

	passwordHash, err := helpers.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}

	tests := []testCase{
		{
			name: "authenticates user and creates refresh token",
			input: input{
				Username: "test-user",
				Password: "correct-password",
			},
			expected: expected{success: true, tokenCount: 1},
		},
		{
			name: "rejects unknown username",
			input: input{
				Username: "unknown",
				Password: "correct-password",
			},
			expected: expected{errorCode: http.StatusUnauthorized},
		},
		{
			name: "rejects incorrect password",
			input: input{
				Username: "test-user",
				Password: "incorrect-password",
			},
			expected: expected{errorCode: http.StatusUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{
				Username:     "test-user",
				Email:        "test-user@example.test",
				PasswordHash: passwordHash,
			})

			got := f.Service.LoginUser(context.Background(), &userSvc.LoginUserIn{
				Trace: fixture.TestTrace(), Username: tt.input.Username, Password: tt.input.Password,
			})
			test.AssertEqual(t, "LoginUser success", tt.expected.success, got.Success)
			test.AssertEqual(t, "LoginUser error code", tt.expected.errorCode, got.ErrorCode)
			tokens := f.UserQuery.FindRefreshTokens(user.ID)
			test.AssertEqual(t, "refresh token count", tt.expected.tokenCount, len(tokens))
			if tt.expected.success {
				test.AssertEqual(t, "LoginUser user ID", user.ID, got.UserID)
				test.AssertEqual(t, "LoginUser username", user.Username, got.Username)
				test.AssertEqual(t, "LoginUser role", user.Role, got.UserRole)
				test.AssertNotEmpty(t, "LoginUser access token", got.Token)
				test.AssertNotEmpty(t, "LoginUser refresh token", got.RefreshToken)
				test.AssertEqual(t, "stored refresh token hash", helpers.HashToken(got.RefreshToken), tokens[0].TokenHash)
			}
		})
	}
}

func TestUserService_Register(t *testing.T) {
	t.Parallel()

	type input struct {
		Username string
		Email    string
		Password string
		FullName string
		UserRole string
	}
	type expected struct {
		success       bool
		errorCode     int
		userCountDiff int64
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}

	// +----------+-------------------------+
	// | Username | Email                   |
	// +----------+-------------------------+
	// | existing | existing@example.test   |
	// +----------+-------------------------+
	tests := []testCase{
		{
			name: "registers user",
			input: input{
				Username: "new-user",
				Email:    "new-user@example.test",
				Password: "correct-password",
				FullName: "New User",
				UserRole: "User",
			},
			expected: expected{success: true, userCountDiff: 1},
		},
		{
			name: "rejects unknown role",
			input: input{
				Username: "new-user",
				Email:    "new-user@example.test",
				Password: "correct-password",
				FullName: "New User",
				UserRole: "Unknown",
			},
			expected: expected{errorCode: http.StatusBadRequest},
		},
		{
			name: "rejects duplicate username",
			input: input{
				Username: "existing",
				Email:    "new-user@example.test",
				Password: "correct-password",
				FullName: "New User",
				UserRole: "User",
			},
			expected: expected{errorCode: http.StatusUnprocessableEntity},
		},
		{
			name: "rejects duplicate email",
			input: input{
				Username: "new-user",
				Email:    "existing@example.test",
				Password: "correct-password",
				FullName: "New User",
				UserRole: "User",
			},
			expected: expected{errorCode: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			f.SeedUser(t, fixture.UserRow{
				Username: "existing",
				Email:    "existing@example.test",
			})
			f.SeedRole(t, "User")
			initialCount := f.UserQuery.CountUsers()

			got := f.Service.Register(context.Background(), &userSvc.UserRegisterIn{
				Trace:    fixture.TestTrace(),
				Username: tt.input.Username,
				Email:    tt.input.Email,
				Password: tt.input.Password,
				FullName: tt.input.FullName,
				UserRole: tt.input.UserRole,
			})
			test.AssertEqual(t, "Register success", tt.expected.success, got.Success)
			test.AssertEqual(t, "Register error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "Register user count", initialCount+tt.expected.userCountDiff, f.UserQuery.CountUsers())

			if !tt.expected.success {
				return
			}
			registered := f.UserQuery.FindByUsername(tt.input.Username)
			test.AssertEqual(t, "registered email", tt.input.Email, registered.Email)
			test.AssertEqual(t, "registered full name", tt.input.FullName, registered.FullName)
			test.AssertEqual(t, "registered role", tt.input.UserRole, registered.Role)
			test.AssertTrue(t, "registered password hash", helpers.VerifyPassword(tt.input.Password, registered.Password))
			test.AssertNotEmpty(t, "registered external ID", registered.ExternalID)
			test.AssertNil(t, "registered deleted at", registered.DeletedAt)
			test.AssertTrue(t, "registered timestamps", registered.Timestamp != nil)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	t.Parallel()

	type input struct {
		UserID uint
	}
	type expected struct {
		success   bool
		errorCode int
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}

	tests := []testCase{
		{
			name:     "returns user",
			expected: expected{success: true},
		},
		{
			name:     "returns not found for unknown user",
			input:    input{UserID: 99999},
			expected: expected{errorCode: http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{
				Username: "lookup-user",
				Email:    "lookup-user@example.test",
			})
			userID := tt.input.UserID
			if userID == 0 {
				userID = user.ID
			}

			got := f.Service.GetUser(context.Background(), &userSvc.GetUserIn{Trace: fixture.TestTrace(), UserID: userID})
			test.AssertEqual(t, "GetUser success", tt.expected.success, got.Success)
			test.AssertEqual(t, "GetUser error code", tt.expected.errorCode, got.ErrorCode)
			if tt.expected.success {
				test.AssertEqual(t, "GetUser ID", user.ID, got.User.ID)
				test.AssertEqual(t, "GetUser username", user.Username, got.User.Username)
				test.AssertEqual(t, "GetUser email", user.Email, got.User.Email)
			} else {
				test.AssertNil(t, "GetUser result", got.User)
			}
		})
	}
}

func TestUserService_InsertFcmToken(t *testing.T) {
	t.Parallel()

	type input struct {
		UserID uint
		Token  string
	}
	type expected struct {
		success   bool
		errorCode int
		tokenRows int
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}

	tests := []testCase{
		{
			name: "inserts token for user",
			input: input{
				Token: "device-token",
			},
			expected: expected{success: true, tokenRows: 1},
		},
		{
			name: "rejects unknown user",
			input: input{
				UserID: 99999,
				Token:  "device-token",
			},
			expected: expected{errorCode: http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{})
			userID := tt.input.UserID
			if userID == 0 {
				userID = user.ID
			}

			got := f.Service.InsertFcmToken(context.Background(), &userSvc.InsertFcmTokenIn{
				Trace:  fixture.TestTrace(),
				UserID: userID,
				Token:  tt.input.Token,
			})
			test.AssertEqual(t, "InsertFcmToken success", tt.expected.success, got.Success)
			test.AssertEqual(t, "InsertFcmToken error code", tt.expected.errorCode, got.ErrorCode)
			tokens := f.UserQuery.FindFcmTokens(user.ID)
			test.AssertEqual(t, "FCM token rows", tt.expected.tokenRows, len(tokens))
			if tt.expected.success {
				test.AssertEqual(t, "inserted FCM token", tt.input.Token, tokens[0].FcmToken)
				test.AssertEqual(t, "FCM token user ID", user.ID, tokens[0].UserID)
				test.AssertTrue(t, "FCM token timestamps", tokens[0].Timestamp != nil)
			}
		})
	}
}

func TestUserService_DeleteFcmTokenBulk(t *testing.T) {
	t.Parallel()

	type input struct {
		Tokens []string
	}
	type expected struct {
		success         bool
		errorCode       int
		remainingTokens []string
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}

	tests := []testCase{
		{
			name:  "deletes requested tokens",
			input: input{Tokens: []string{"remove-token"}},
			expected: expected{
				success:         true,
				remainingTokens: []string{"keep-token"},
			},
		},
		{
			name:  "accepts empty token list without changing data",
			input: input{Tokens: []string{}},
			expected: expected{
				success:         true,
				remainingTokens: []string{"remove-token", "keep-token"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{})
			f.SeedFcmToken(t, fixture.FcmTokenRow{UserID: user.ID, Token: "remove-token"})
			f.SeedFcmToken(t, fixture.FcmTokenRow{UserID: user.ID, Token: "keep-token"})

			got := f.Service.DeleteFcmTokenBulk(context.Background(), &userSvc.DeleteFcmTokenBulkIn{
				Trace:  fixture.TestTrace(),
				Tokens: tt.input.Tokens,
			})
			test.AssertEqual(t, "DeleteFcmTokenBulk success", tt.expected.success, got.Success)
			test.AssertEqual(t, "DeleteFcmTokenBulk error code", tt.expected.errorCode, got.ErrorCode)
			tokens := f.UserQuery.FindFcmTokens(user.ID)
			test.AssertEqual(t, "remaining FCM token count", len(tt.expected.remainingTokens), len(tokens))
			for i, token := range tt.expected.remainingTokens {
				test.AssertEqual(t, "remaining FCM token", token, tokens[i].FcmToken)
			}
		})
	}
}

func TestUserService_GetUserFcmToken(t *testing.T) {
	t.Parallel()

	type input struct {
		UserID uint
	}
	type expected struct {
		success   bool
		errorCode int
		tokenRows int
	}
	type testCase struct {
		name     string
		input    input
		expected expected
	}

	tests := []testCase{
		{
			name:     "returns user tokens",
			expected: expected{success: true, tokenRows: 1},
		},
		{
			name:     "returns empty token list",
			expected: expected{success: true},
		},
		{
			name:     "returns not found for unknown user",
			input:    input{UserID: 99999},
			expected: expected{errorCode: http.StatusNotFound},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{})
			if i == 0 {
				f.SeedFcmToken(t, fixture.FcmTokenRow{UserID: user.ID, Token: "device-token"})
			}
			userID := tt.input.UserID
			if userID == 0 {
				userID = user.ID
			}

			got := f.Service.GetUserFcmToken(context.Background(), &userSvc.GetUserFcmTokenIn{Trace: fixture.TestTrace(), UserID: userID})
			test.AssertEqual(t, "GetUserFcmToken success", tt.expected.success, got.Success)
			test.AssertEqual(t, "GetUserFcmToken error code", tt.expected.errorCode, got.ErrorCode)
			test.AssertEqual(t, "GetUserFcmToken token count", tt.expected.tokenRows, len(got.FcmTokens))
			if tt.expected.success && tt.expected.tokenRows > 0 {
				test.AssertEqual(t, "GetUserFcmToken token", "device-token", got.FcmTokens[0].FcmToken)
			}
		})
	}
}

func TestUserService_Refresh(t *testing.T) {
	t.Parallel()

	type input struct {
		UserID       int
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
			name: "rotates refresh token",
			input: input{
				RefreshToken: "valid-token",
			},
			expected: expected{success: true, tokenRows: 2},
		},
		{
			name: "rejects unknown token",
			input: input{
				RefreshToken: "missing-token",
			},
			expected:  expected{errorCode: http.StatusNotFound, tokenRows: 1},
			seedToken: "existing-token",
		},
		{
			name: "rejects expired token",
			input: input{
				RefreshToken: "expired-token",
			},
			expected: expected{errorCode: http.StatusForbidden, tokenRows: 1},
			expired:  true,
		},
		{
			name: "revokes family when used token is reused",
			input: input{
				RefreshToken: "used-token",
			},
			expected: expected{errorCode: http.StatusForbidden, tokenRows: 1, revoked: true},
			used:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			user := f.SeedUser(t, fixture.UserRow{})
			userID := tt.input.UserID
			if userID == 0 {
				userID = int(user.ID)
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
			seedToken := tt.seedToken
			if seedToken == "" {
				seedToken = tt.input.RefreshToken
			}
			f.SeedRefreshToken(t, fixture.RefreshTokenRow{
				UserID:   user.ID,
				Token:    seedToken,
				Family:   "token-family",
				ExpireAt: expireAt,
				UsedAt:   usedAt,
			})

			got := f.Service.Refresh(context.Background(), &userSvc.RefreshIn{
				Trace:        fixture.TestTrace(),
				UserId:       userID,
				RefreshToken: tt.input.RefreshToken,
			})
			test.AssertEqual(t, "Refresh success", tt.expected.success, got.Success)
			test.AssertEqual(t, "Refresh error code", tt.expected.errorCode, got.ErrorCode)
			tokens := f.UserQuery.FindRefreshTokens(user.ID)
			test.AssertEqual(t, "Refresh token row count", tt.expected.tokenRows, len(tokens))
			if tt.expected.success {
				test.AssertNotEmpty(t, "Refresh access token", got.AccessToken)
				test.AssertNotEmpty(t, "Refresh new token", got.RefreshToken)
				test.AssertEqual(t, "Refresh new token hash", helpers.HashToken(got.RefreshToken), tokens[1].TokenHash)
				test.AssertTrue(t, "Refresh old token used at", tokens[0].UsedAt != nil)
			}
			if tt.expected.revoked {
				test.AssertTrue(t, "Refresh reused token revoked at", tokens[0].RevokedAt != nil)
			}
		})
	}
}
