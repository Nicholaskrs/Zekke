package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"template-go/base/helpers"
	"template-go/data/model"
	notificationSvc "template-go/modules/notification/svc"
	permissionRepo "template-go/modules/permission/repository"
	permissionSvc "template-go/modules/permission/svc"
	userRepo "template-go/modules/user/repository"
	userSvc "template-go/modules/user/svc"
	"template-go/test/mocks"
	"template-go/util/config"
	"template-go/util/logtrace"
)

type userFixture struct {
	db      *gorm.DB
	service userSvc.UserService
}

func newUserFixture(t *testing.T) *userFixture {
	t.Helper()
	db, cfg := newTestDatabase(t)
	storage := userRepo.NewUserStorage(db)
	return &userFixture{
		db:      db,
		service: userSvc.NewUserService(*storage, testConfig(cfg)),
	}
}

type permissionFixture struct {
	db      *gorm.DB
	service permissionSvc.PermissionAdminService
}

func newPermissionFixture(t *testing.T) *permissionFixture {
	t.Helper()
	db, _ := newTestDatabase(t)
	storage := permissionRepo.NewPermissionStorage(db)
	return &permissionFixture{
		db:      db,
		service: permissionSvc.NewPermissionAdminService(*storage),
	}
}

type userAdminFixture struct {
	db             *gorm.DB
	service        userSvc.UserAdminService
	permissionMock *mocks.PermissionAdminServiceMock
}

func newUserAdminFixture(t *testing.T) *userAdminFixture {
	t.Helper()
	db, cfg := newTestDatabase(t)
	storage := userRepo.NewUserStorage(db)
	permissionMock := &mocks.PermissionAdminServiceMock{}
	return &userAdminFixture{
		db:             db,
		service:        userSvc.NewUserAdminService(*storage, permissionMock, testConfig(cfg)),
		permissionMock: permissionMock,
	}
}

type notificationFixture struct {
	service      notificationSvc.NotificationService
	userMock     *mocks.UserServiceMock
	firebaseMock *mocks.SenderMock
}

func newNotificationFixture(t *testing.T) *notificationFixture {
	t.Helper()
	userMock := &mocks.UserServiceMock{}
	firebaseMock := &mocks.SenderMock{}
	return &notificationFixture{
		service:      notificationSvc.NewNotificationService(firebaseMock, userMock),
		userMock:     userMock,
		firebaseMock: firebaseMock,
	}
}

func testConfig(cfg config.Config) config.Config {
	cfg.JwtSecret = "service-test-secret"
	cfg.JwtIssuer = "zekke-service-test"
	return cfg
}

func testTrace() *logtrace.LogTrace {
	return &logtrace.LogTrace{TraceId: uuid.NewString()}
}

type userSeedOption func(*model.User)

func seedUser(t *testing.T, db *gorm.DB, opts ...userSeedOption) *model.User {
	t.Helper()
	user := &model.User{
		ExternalID: uuid.NewString(),
		Username:   "user-" + uuid.NewString()[:8],
		Email:      "user-" + uuid.NewString()[:8] + "@example.test",
		Password:   "unused",
		FullName:   "Test User",
		Role:       "User",
	}
	for _, opt := range opts {
		opt(user)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}

func withUsername(username string) userSeedOption {
	return func(user *model.User) { user.Username = username }
}

func withEmail(email string) userSeedOption {
	return func(user *model.User) { user.Email = email }
}

func withRole(role string) userSeedOption {
	return func(user *model.User) { user.Role = role }
}

func withPasswordHash(passwordHash string) userSeedOption {
	return func(user *model.User) { user.Password = passwordHash }
}

type userSeedData struct {
	username     string
	passwordHash string
	role         string
}

func seedUserWithData(t *testing.T, db *gorm.DB, data userSeedData) *model.User {
	t.Helper()
	opts := []userSeedOption{
		withUsername(data.username),
		withPasswordHash(data.passwordHash),
		withRole(data.role),
	}
	return seedUser(t, db, opts...)
}

// userRow describes a single seeded user row for seedUsers. Fields left at
// their zero value fall back to the defaults applied by seedUser.
type userRow struct {
	Username     string
	Email        string
	Role         string
	PasswordHash string
	DeletedAt    *time.Time
}

// seedUsers seeds one user row per entry, in order, and returns the created
// users in the same order so callers can reference them (e.g. by index) in
// their test cases.
func seedUsers(t *testing.T, db *gorm.DB, rows ...userRow) []*model.User {
	t.Helper()
	users := make([]*model.User, 0, len(rows))
	for _, row := range rows {
		var opts []userSeedOption
		if row.Username != "" {
			opts = append(opts, withUsername(row.Username))
		}
		if row.Email != "" {
			opts = append(opts, withEmail(row.Email))
		}
		if row.Role != "" {
			opts = append(opts, withRole(row.Role))
		}
		if row.PasswordHash != "" {
			opts = append(opts, withPasswordHash(row.PasswordHash))
		}
		user := seedUser(t, db, opts...)
		if row.DeletedAt != nil {
			if err := db.Model(user).Update("deleted_at", *row.DeletedAt).Error; err != nil {
				t.Fatalf("mark user deleted: %v", err)
			}
		}
		users = append(users, user)
	}
	return users
}

func seedRole(t *testing.T, db *gorm.DB, code string) *model.Role {
	t.Helper()
	role := &model.Role{Code: code, Description: code + " role"}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return role
}

func seedPrivilegeGroup(t *testing.T, db *gorm.DB, code string) *model.PrivilegeGroup {
	t.Helper()
	group := &model.PrivilegeGroup{Code: code, Description: code + " group"}
	if err := db.Create(group).Error; err != nil {
		t.Fatalf("seed privilege group: %v", err)
	}
	return group
}

func seedGroupPrivilege(t *testing.T, db *gorm.DB, group, privilege string) *model.PrivilegeGroupPrivilege {
	t.Helper()
	link := &model.PrivilegeGroupPrivilege{PrivilegeGroupCode: group, Privilege: privilege}
	if err := db.Create(link).Error; err != nil {
		t.Fatalf("seed group privilege: %v", err)
	}
	return link
}

func seedUserPrivilegeGroup(t *testing.T, db *gorm.DB, userID uint, group string) {
	t.Helper()
	if err := db.Create(&model.UserPrivilegeGroup{UserID: userID, PrivilegeGroupCode: group}).Error; err != nil {
		t.Fatalf("seed user privilege group: %v", err)
	}
}

func seedFCMToken(t *testing.T, db *gorm.DB, userID uint, token string) *model.FcmToken {
	t.Helper()
	now := time.Now()
	fcmToken := &model.FcmToken{
		UserID:   userID,
		FcmToken: token,
		Timestamp: &model.Timestamp{
			CreatedTs:     now,
			LastUpdatedTs: now,
		},
	}
	if err := db.Create(fcmToken).Error; err != nil {
		t.Fatalf("seed FCM token: %v", err)
	}
	return fcmToken
}

// fcmTokenRow describes a single seeded FCM token for seedFCMTokens.
type fcmTokenRow struct {
	UserID uint
	Token  string
}

// seedFCMTokens seeds one FCM token row per entry, in order, so a batch of
// tokens for one or more users can be declared as a single readable table.
func seedFCMTokens(t *testing.T, db *gorm.DB, rows ...fcmTokenRow) []*model.FcmToken {
	t.Helper()
	tokens := make([]*model.FcmToken, 0, len(rows))
	for _, row := range rows {
		tokens = append(tokens, seedFCMToken(t, db, row.UserID, row.Token))
	}
	return tokens
}

func seedRefreshToken(t *testing.T, db *gorm.DB, userID uint, rawToken, familyID string, expireAt time.Time, usedAt *time.Time) *model.RefreshToken {
	t.Helper()
	now := time.Now()
	token := &model.RefreshToken{
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: helpers.HashToken(rawToken),
		ExpireAt:  &expireAt,
		UsedAt:    usedAt,
		Timestamp: &model.Timestamp{
			CreatedTs:     now,
			LastUpdatedTs: now,
		},
	}
	if err := db.Create(token).Error; err != nil {
		t.Fatalf("seed refresh token: %v", err)
	}
	return token
}

// refreshTokenRow describes a single seeded refresh token for
// seedRefreshTokens.
type refreshTokenRow struct {
	UserID   uint
	Token    string
	Family   string
	ExpireAt time.Time
	UsedAt   *time.Time
}

// seedRefreshTokens seeds one refresh token row per entry, in order, so a
// family of related tokens can be declared as a single readable table
// instead of a sequence of individual seedRefreshToken calls.
func seedRefreshTokens(t *testing.T, db *gorm.DB, rows ...refreshTokenRow) []*model.RefreshToken {
	t.Helper()
	tokens := make([]*model.RefreshToken, 0, len(rows))
	for _, row := range rows {
		tokens = append(tokens, seedRefreshToken(t, db, row.UserID, row.Token, row.Family, row.ExpireAt, row.UsedAt))
	}
	return tokens
}
