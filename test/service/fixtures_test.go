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
	return newUserFixtureWithDB(db, cfg)
}

func newUserFixtureWithDB(db *gorm.DB, cfg config.Config) *userFixture {
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

// seedUser inserts a user with defaults that can be overridden by options.
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

// withUsername overrides the seeded user's username.
func withUsername(username string) userSeedOption {
	return func(user *model.User) { user.Username = username }
}

// withEmail overrides the seeded user's email.
func withEmail(email string) userSeedOption {
	return func(user *model.User) { user.Email = email }
}

// withRole overrides the seeded user's role.
func withRole(role string) userSeedOption {
	return func(user *model.User) { user.Role = role }
}

// withPasswordHash overrides the seeded user's password hash.
func withPasswordHash(passwordHash string) userSeedOption {
	return func(user *model.User) { user.Password = passwordHash }
}

// seedRole inserts a role with a description derived from its code.
func seedRole(t *testing.T, db *gorm.DB, code string) *model.Role {
	t.Helper()
	role := &model.Role{Code: code, Description: code + " role"}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return role
}

// seedPrivilegeGroup inserts a privilege group with a description derived from its code.
func seedPrivilegeGroup(t *testing.T, db *gorm.DB, code string) *model.PrivilegeGroup {
	t.Helper()
	group := &model.PrivilegeGroup{Code: code, Description: code + " group"}
	if err := db.Create(group).Error; err != nil {
		t.Fatalf("seed privilege group: %v", err)
	}
	return group
}

// seedGroupPrivilege inserts a privilege link for the specified group.
func seedGroupPrivilege(t *testing.T, db *gorm.DB, group, privilege string) *model.PrivilegeGroupPrivilege {
	t.Helper()
	link := &model.PrivilegeGroupPrivilege{PrivilegeGroupCode: group, Privilege: privilege}
	if err := db.Create(link).Error; err != nil {
		t.Fatalf("seed group privilege: %v", err)
	}
	return link
}

// seedUserPrivilegeGroup assigns a user to the specified privilege group.
func seedUserPrivilegeGroup(t *testing.T, db *gorm.DB, userID uint, group string) {
	t.Helper()
	if err := db.Create(&model.UserPrivilegeGroup{UserID: userID, PrivilegeGroupCode: group}).Error; err != nil {
		t.Fatalf("seed user privilege group: %v", err)
	}
}

// seedFCMToken inserts an FCM token for the specified user.
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

// seedRefreshToken inserts a refresh token after hashing its raw token value.
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
