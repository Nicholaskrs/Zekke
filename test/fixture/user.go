package fixture

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"template-go/base/helpers"
	"template-go/data/model"
	userRepo "template-go/modules/user/repository"
	userSvc "template-go/modules/user/svc"
	"template-go/test/mocks"
)

// =====================================================================
// Fixtures
// =====================================================================

type UserFixture struct {
	db                *gorm.DB
	Service           userSvc.UserService
	AdminService      userSvc.UserAdminService
	PermissionService *mocks.PermissionAdminServiceMock
	UserQuery         *UserQuery
}

func NewUserFixture(t *testing.T) *UserFixture {
	t.Helper()
	db, cfg := NewTestDatabase(t)
	storage := userRepo.NewUserStorage(db)
	permissionService := &mocks.PermissionAdminServiceMock{}
	return &UserFixture{
		db:                db,
		Service:           userSvc.NewUserService(*storage, testConfig(cfg)),
		AdminService:      userSvc.NewUserAdminService(*storage, permissionService, testConfig(cfg)),
		PermissionService: permissionService,
		UserQuery:         NewUserQuery(t, db),
	}
}

// =====================================================================
// Users
// =====================================================================

// UserRow describes one user. Empty fields fall back to defaults.
type UserRow struct {
	Username     string
	Email        string
	Role         string
	PasswordHash string
	DeletedAt    *time.Time // set to soft-delete the user
}

// SeedUser creates one user.
//
//	seedUser(t, db, userRow{})                  // all defaults
//	seedUser(t, db, userRow{Role: "Admin"})
func (u *UserFixture) SeedUser(t *testing.T, row UserRow) *model.User {
	t.Helper()

	user := &model.User{
		ExternalID: uuid.NewString(),
		Username:   orDefault(row.Username, "user-"+shortID()),
		Email:      orDefault(row.Email, "user-"+shortID()+"@example.test"),
		Password:   orDefault(row.PasswordHash, "unused"),
		FullName:   "Test User",
		Role:       orDefault(row.Role, "User"),
		DeletedAt:  row.DeletedAt,
	}
	mustCreate(t, u.db, user, "user")

	return user
}

// SeedUsers creates one user per row, in order, and returns them in the same
// order so tests can refer to them by index.
func (u *UserFixture) SeedUsers(t *testing.T, rows []UserRow) []*model.User {
	t.Helper()

	users := make([]*model.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, u.SeedUser(t, row))
	}
	return users
}

// =====================================================================
// Refresh tokens
// =====================================================================

// RefreshTokenRow describes one refresh token. Empty fields fall back to
// defaults: random token and family, expiring in one hour, not used.
type RefreshTokenRow struct {
	UserID   uint
	Token    string // raw token; it is hashed before being stored
	Family   string
	ExpireAt time.Time
	UsedAt   *time.Time
}

// SeedRefreshToken creates one refresh token.
func (u *UserFixture) SeedRefreshToken(t *testing.T, row RefreshTokenRow) *model.RefreshToken {
	t.Helper()
	currTime, err := helpers.TimeNow()
	if err != nil {
		currTime = time.Now()
	}

	expireAt := row.ExpireAt
	if expireAt.IsZero() {
		expireAt = currTime.Add(time.Hour)
	}

	token := &model.RefreshToken{
		UserID:    row.UserID,
		FamilyID:  orDefault(row.Family, uuid.NewString()),
		TokenHash: helpers.HashToken(orDefault(row.Token, uuid.NewString())),
		ExpireAt:  &expireAt,
		UsedAt:    row.UsedAt,
		Timestamp: &model.Timestamp{
			CreatedTs:     currTime,
			LastUpdatedTs: currTime,
		},
	}

	mustCreate(t, u.db, token, "refresh token")
	return token
}

// SeedRefreshTokens creates a batch of refresh tokens, in order.
func (u *UserFixture) SeedRefreshTokens(t *testing.T, rows []RefreshTokenRow) []*model.RefreshToken {
	t.Helper()

	tokens := make([]*model.RefreshToken, 0, len(rows))
	for _, row := range rows {
		tokens = append(tokens, u.SeedRefreshToken(t, row))
	}
	return tokens
}
