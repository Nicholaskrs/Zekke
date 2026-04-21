package user

import (
	"context"
	"template-go/base/helpers"
	"template-go/data/model"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

type UserStorage struct {
	db *gorm.DB
}

// NewUserStorage creates a new instance of UserRepository
func NewUserStorage(db *gorm.DB) *UserStorage {
	return &UserStorage{db: db}
}

type UserRepository struct {
	transaction *gorm.DB
}

func (u *UserStorage) BeginTx(ctx context.Context) *UserRepository {
	return &UserRepository{transaction: u.db.WithContext(ctx).Begin()}
}

// Commit is used to commit database changes.
func (repo *UserRepository) Commit() error {
	transaction := repo.transaction.Commit()
	if transaction.Error != nil {
		return transaction.Error
	}
	return nil
}

// Rollback is used to rollback database changes.
func (repo *UserRepository) Rollback() {
	repo.transaction.Rollback()
}

func (repo *UserRepository) LockUser(userId uint) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Clauses(
		clause.Locking{
			Strength: "UPDATE",
		},
	).Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err, true)
	}
	return &user, nil
}

// FindUserByUsername retrieves a user by their username
func (repo *UserRepository) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err, true)
	}
	return &user, nil
}

// FindUserByID retrieves a user by their userId
func (repo *UserRepository) FindUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err, true)
	}
	return &user, nil
}

// CheckUserExistsByUsername returns bool as flag if user is exists based on username
func (repo *UserRepository) CheckUserExistsByUsername(username string) (bool, error) {
	var count int64
	if err := repo.transaction.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, helpers.Wrap(err, true)
	}
	return count > 0, nil
}

// CreateUser used to insert new row data. It returns inserted ID and error
func (repo *UserRepository) CreateUser(param *model.User) (uint, error) {
	if err := repo.transaction.Create(&param).Error; err != nil {
		return 0, helpers.Wrap(err, true)
	}
	return param.ID, nil
}

func (repo *UserRepository) UpdateUser(param *model.User) error {
	return repo.transaction.Model(param).Where("id = ?", param.ID).Updates(param).Error
}

// FindUserByExternalID retrieves a user by their externalID
func (repo *UserRepository) FindUserByExternalID(externalID string) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Where("external_id = ?", externalID).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err, true)
	}
	return &user, nil
}

// FilterUser retrieves a user based on given filter. It returns model user and error
func (repo *UserRepository) FilterUser(filterUser FilterUser, page int, limit int) ([]*model.User, error) {
	var users []*model.User

	query := repo.transaction

	// Where Condition.
	if filterUser.Role != "" {
		query = query.Where("role = ?", filterUser.Role)
	}

	// Find model
	if err := query.
		Offset(helpers.GetOffset(page, limit)).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, helpers.Wrap(err, true)
	}
	return users, nil
}

// FilterUserCount retrieves a row count based on given filter.
func (repo *UserRepository) FilterUserCount(filterUser FilterUser) (int64, error) {
	var totalRows int64

	query := repo.transaction

	// Where Condition.
	if filterUser.Role != "" {
		query = query.Where("role = ?", filterUser.Role)
	}

	// Find total rows
	if err := query.Model(model.User{}).
		Count(&totalRows).Error; err != nil {
		return 0, helpers.Wrap(err, true)
	}

	return totalRows, nil
}

// GetAllUser retrieves a all users.
func (repo *UserRepository) GetAllUser() ([]uint, error) {
	var userIDs []uint
	if err := repo.transaction.Model(&model.User{}).Select("id").Find(&userIDs).Error; err != nil {
		return nil, helpers.Wrap(err, true)
	}
	return userIDs, nil
}

func (repo *UserRepository) CreateFcmToken(param *model.FcmToken) error {
	if err := repo.transaction.Create(param).Error; err != nil {
		return helpers.Wrap(err, true)
	}
	return nil
}

func (repo *UserRepository) DeleteFcmTokenBulk(tokens []string) error {
	if err := repo.transaction.Where("fcm_token IN (?)", tokens).Delete(&model.FcmToken{}).Error; err != nil {
		return helpers.Wrap(err, true)
	}
	return nil
}

func (repo *UserRepository) GetUserFcmToken(userId uint) ([]*model.FcmToken, error) {
	var fcmTokens []*model.FcmToken
	if err := repo.transaction.Where("user_id = ?", userId).Find(&fcmTokens).Error; err != nil {
		return nil, helpers.Wrap(err, true)
	}
	return fcmTokens, nil
}
