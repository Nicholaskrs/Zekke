package user

import (
	"template-go/base/helpers"
	"template-go/data/model"

	"gorm.io/gorm"
)

type UserRepositoryRead struct {
	transaction *gorm.DB
}

// FindUserByUsername retrieves a user by their username
func (repo *UserRepositoryRead) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// FindUserByID retrieves a user by their userId
func (repo *UserRepositoryRead) FindUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// CheckUserExistsByUsername returns bool as flag if user is exists based on username
func (repo *UserRepositoryRead) CheckUserExistsByUsername(username string) (bool, error) {
	var count int64
	if err := repo.transaction.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, helpers.Wrap(err)
	}
	return count > 0, nil
}

// FindUserByExternalID retrieves a user by their externalID
func (repo *UserRepositoryRead) FindUserByExternalID(externalID string) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Where("external_id = ?", externalID).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// FilterUser retrieves a user based on given filter. It returns model user and error
func (repo *UserRepositoryRead) FilterUser(filterUser FilterUser, page int, limit int) ([]*model.User, error) {
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
		return nil, helpers.Wrap(err)
	}
	return users, nil
}

// FilterUserCount retrieves a row count based on given filter.
func (repo *UserRepositoryRead) FilterUserCount(filterUser FilterUser) (int64, error) {
	var totalRows int64

	query := repo.transaction

	// Where Condition.
	if filterUser.Role != "" {
		query = query.Where("role = ?", filterUser.Role)
	}

	// Find total rows
	if err := query.Model(model.User{}).
		Count(&totalRows).Error; err != nil {
		return 0, helpers.Wrap(err)
	}

	return totalRows, nil
}

// GetAllUser retrieves a all users.
func (repo *UserRepositoryRead) GetAllUser() ([]uint, error) {
	var userIDs []uint
	if err := repo.transaction.Model(&model.User{}).Select("id").Find(&userIDs).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return userIDs, nil
}

func (repo *UserRepositoryRead) GetUserFcmToken(userId uint) ([]*model.FcmToken, error) {
	var fcmTokens []*model.FcmToken
	if err := repo.transaction.Where("user_id = ?", userId).Find(&fcmTokens).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return fcmTokens, nil
}
