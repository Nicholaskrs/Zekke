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

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *userFixture) *userSvc.LoginUserIn
		wantSuccess bool
		wantCode    int
	}{
		{
			name: "issues access and refresh tokens for valid credentials",
			setup: func(t *testing.T, f *userFixture) *userSvc.LoginUserIn {
				passwordHash := hashPassword(t, "valid-password")
				seedUser(t, f.db, withUsername("login-user"), withPasswordHash(passwordHash))
				return &userSvc.LoginUserIn{Trace: testTrace(), Username: "login-user", Password: "valid-password"}
			},
			wantSuccess: true,
		},
		{
			name: "rejects unknown username",
			setup: func(t *testing.T, _ *userFixture) *userSvc.LoginUserIn {
				return &userSvc.LoginUserIn{Trace: testTrace(), Username: "missing", Password: "valid-password"}
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "rejects incorrect password",
			setup: func(t *testing.T, f *userFixture) *userSvc.LoginUserIn {
				seedUser(t, f.db, withUsername("login-user"), withPasswordHash(hashPassword(t, "valid-password")))
				return &userSvc.LoginUserIn{Trace: testTrace(), Username: "login-user", Password: "incorrect"}
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "does not allow administrator accounts to use public login",
			setup: func(t *testing.T, f *userFixture) *userSvc.LoginUserIn {
				seedUser(t, f.db, withUsername("admin"), withRole("Admin"), withPasswordHash(hashPassword(t, "valid-password")))
				return &userSvc.LoginUserIn{Trace: testTrace(), Username: "admin", Password: "valid-password"}
			},
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			got := f.service.LoginUser(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("LoginUser() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var count int64
			if err := f.db.Model(&model.RefreshToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if tt.wantSuccess {
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

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *userFixture) *userSvc.UserRegisterIn
		wantSuccess bool
		wantCode    int
		wantUsers   int64
	}{
		{
			name: "registers user with existing role",
			setup: func(t *testing.T, f *userFixture) *userSvc.UserRegisterIn {
				seedRole(t, f.db, "Member")
				return &userSvc.UserRegisterIn{
					Trace: testTrace(), Username: "new-user", Email: "new-user@example.test",
					Password: "password", FullName: "New User", UserRole: "Member",
				}
			},
			wantSuccess: true,
			wantUsers:   1,
		},
		{
			name: "rejects missing role",
			setup: func(t *testing.T, _ *userFixture) *userSvc.UserRegisterIn {
				return &userSvc.UserRegisterIn{
					Trace: testTrace(), Username: "new-user", Email: "new-user@example.test",
					Password: "password", FullName: "New User", UserRole: "Missing",
				}
			},
			wantCode: http.StatusBadRequest,
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
			wantCode:  http.StatusUnprocessableEntity,
			wantUsers: 1,
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
			wantCode:  http.StatusInternalServerError,
			wantUsers: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			got := f.service.Register(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("Register() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var count int64
			if err := f.db.Model(&model.User{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantUsers {
				t.Fatalf("persisted user count = %d, want %d", count, tt.wantUsers)
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *userFixture) uint
		wantSuccess bool
		wantCode    int
	}{
		{
			name: "returns existing user",
			setup: func(t *testing.T, f *userFixture) uint {
				return seedUser(t, f.db).ID
			},
			wantSuccess: true,
		},
		{
			name: "returns not found for missing user",
			setup: func(t *testing.T, _ *userFixture) uint {
				return 99
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			id := tt.setup(t, f)
			got := f.service.GetUser(context.Background(), &userSvc.GetUserIn{Trace: testTrace(), UserID: id})
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("GetUser() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			if tt.wantSuccess && got.User.ID != id {
				t.Errorf("GetUser() user ID = %d, want %d", got.User.ID, id)
			}
		})
	}
}

func TestUserService_InsertFcmToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *userFixture) *userSvc.InsertFcmTokenIn
		wantSuccess bool
		wantCode    int
		wantTokens  int64
	}{
		{
			name: "inserts token for existing user",
			setup: func(t *testing.T, f *userFixture) *userSvc.InsertFcmTokenIn {
				user := seedUser(t, f.db)
				return &userSvc.InsertFcmTokenIn{Trace: testTrace(), UserID: user.ID, Token: "fcm-token"}
			},
			wantSuccess: true,
			wantTokens:  1,
		},
		{
			name: "rejects missing user",
			setup: func(t *testing.T, _ *userFixture) *userSvc.InsertFcmTokenIn {
				return &userSvc.InsertFcmTokenIn{Trace: testTrace(), UserID: 42, Token: "fcm-token"}
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			got := f.service.InsertFcmToken(context.Background(), tt.setup(t, f))
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("InsertFcmToken() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			var count int64
			if err := f.db.Model(&model.FcmToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantTokens {
				t.Fatalf("persisted FCM token count = %d, want %d", count, tt.wantTokens)
			}
		})
	}
}

func TestUserService_DeleteFcmTokenBulk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T, f *userFixture) []string
		wantSuccess   bool
		wantRemaining int64
	}{
		{
			name: "deletes matching tokens only",
			setup: func(t *testing.T, f *userFixture) []string {
				user := seedUser(t, f.db)
				seedFCMToken(t, f.db, user.ID, "remove-me")
				seedFCMToken(t, f.db, user.ID, "keep-me")
				return []string{"remove-me"}
			},
			wantSuccess:   true,
			wantRemaining: 1,
		},
		{
			name:        "succeeds with empty token set",
			setup:       func(t *testing.T, _ *userFixture) []string { return nil },
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			input := &userSvc.DeleteFcmTokenBulkIn{Trace: testTrace(), Tokens: tt.setup(t, f)}
			got := f.service.DeleteFcmTokenBulk(context.Background(), input)
			if got.Success != tt.wantSuccess {
				t.Fatalf("DeleteFcmTokenBulk() = %+v, want success %t", got, tt.wantSuccess)
			}
			var count int64
			if err := f.db.Model(&model.FcmToken{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tt.wantRemaining {
				t.Fatalf("remaining FCM token count = %d, want %d", count, tt.wantRemaining)
			}
		})
	}
}

func TestUserService_GetUserFcmToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T, f *userFixture) uint
		wantSuccess    bool
		wantCode       int
		wantTokenCount int
	}{
		{
			name: "returns a user's tokens",
			setup: func(t *testing.T, f *userFixture) uint {
				user := seedUser(t, f.db)
				seedFCMToken(t, f.db, user.ID, "token-1")
				return user.ID
			},
			wantSuccess:    true,
			wantTokenCount: 1,
		},
		{
			name: "returns an empty list when user has no tokens",
			setup: func(t *testing.T, f *userFixture) uint {
				return seedUser(t, f.db).ID
			},
			wantSuccess: true,
		},
		{
			name:     "rejects missing user",
			setup:    func(t *testing.T, _ *userFixture) uint { return 42 },
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
			got := f.service.GetUserFcmToken(context.Background(), &userSvc.GetUserFcmTokenIn{
				Trace: testTrace(), UserID: tt.setup(t, f),
			})
			if got.Success != tt.wantSuccess || got.ErrorCode != tt.wantCode {
				t.Fatalf("GetUserFcmToken() = %+v, want success %t and code %d", got, tt.wantSuccess, tt.wantCode)
			}
			if len(got.FcmTokens) != tt.wantTokenCount {
				t.Fatalf("returned FCM token count = %d, want %d", len(got.FcmTokens), tt.wantTokenCount)
			}
		})
	}
}

func TestUserService_Refresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, f *userFixture) *userSvc.RefreshIn
		wantSuccess bool
		wantCode    int
		wantTokens  int64
	}{
		{
			name: "rotates a valid refresh token",
			setup: func(t *testing.T, f *userFixture) *userSvc.RefreshIn {
				user := seedUser(t, f.db)
				seedRefreshToken(t, f.db, user.ID, "old-token", "family", time.Now().Add(time.Hour), nil)
				return &userSvc.RefreshIn{Trace: testTrace(), UserId: int(user.ID), RefreshToken: "old-token"}
			},
			wantSuccess: true,
			wantTokens:  2,
		},
		{
			name: "rejects missing refresh token",
			setup: func(t *testing.T, f *userFixture) *userSvc.RefreshIn {
				user := seedUser(t, f.db)
				return &userSvc.RefreshIn{Trace: testTrace(), UserId: int(user.ID), RefreshToken: "missing"}
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "rejects expired refresh token",
			setup: func(t *testing.T, f *userFixture) *userSvc.RefreshIn {
				user := seedUser(t, f.db)
				seedRefreshToken(t, f.db, user.ID, "expired", "family", time.Now().Add(-time.Hour), nil)
				return &userSvc.RefreshIn{Trace: testTrace(), UserId: int(user.ID), RefreshToken: "expired"}
			},
			wantCode:   http.StatusForbidden,
			wantTokens: 1,
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
			wantCode:   http.StatusForbidden,
			wantTokens: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUserFixture(t)
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
			if tt.wantSuccess {
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
