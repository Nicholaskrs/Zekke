package user

import (
	"template-go/base/helpers"
	"template-go/data/model"

	"gorm.io/gorm"
)

type UserRepositoryRead struct {
	transaction *gorm.DB
}

// FindUserByUsername retrieves a user by their username.
func (r *UserRepositoryRead) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.transaction.Where("username = ?", username).Where("role != ?", "Admin").Where("deleted_at IS NULL").First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// FindUserAdminByUsername retrieves a admin by their username.
func (r *UserRepositoryRead) FindUserAdminByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.transaction.Where("username = ?", username).Where("role = ?", "Admin").Where("deleted_at IS NULL").First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// CheckUsernameExists check if username is available.
func (r *UserRepositoryRead) CheckUsernameExists(username string) (bool, error) {
	var count int64
	if err := r.transaction.
		Model(&model.User{}).
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		Count(&count).Error; err != nil {
		return false, helpers.Wrap(err)
	}
	return count > 0, nil
}

// CheckEmailExists check if username is available.
func (r *UserRepositoryRead) CheckEmailExists(email string) (bool, error) {
	var count int64
	if err := r.transaction.
		Model(&model.User{}).
		Where("email = ?", email).
		Where("deleted_at IS NULL").
		Count(&count).Error; err != nil {
		return false, helpers.Wrap(err)
	}
	return count > 0, nil
}

// FindUserByID retrieves a user by their userId.
func (r *UserRepositoryRead) FindUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.transaction.Where("id = ?", id).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// CheckUserExistsByUsername returns bool as flag if user is exists based on username.
func (r *UserRepositoryRead) CheckUserExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.transaction.Model(&model.User{}).Where("username = ?", username).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return false, helpers.Wrap(err)
	}
	return count > 0, nil
}

// FindUserByExternalID retrieves a user by their externalID.
func (r *UserRepositoryRead) FindUserByExternalID(externalID string) (*model.User, error) {
	var user model.User
	if err := r.transaction.Where("external_id = ?", externalID).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// FilterUser retrieves a user based on given filter. It returns model user and error.
func (r *UserRepositoryRead) FilterUser(filterUser FilterUser) ([]*model.User, error) {
	var users []*model.User

	query := r.transaction

	// Where Condition.
	query = r.applyFilterUser(filterUser, query)

	// Find model
	if err := query.
		Offset(filterUser.GetOffset()).
		Limit(filterUser.Limit).
		Find(&users).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return users, nil
}

// FilterUserCount retrieves a row count based on given filter.
func (r *UserRepositoryRead) FilterUserCount(filterUser FilterUser) (int, error) {
	var totalRows int64

	query := r.transaction

	// Where Condition.
	query = r.applyFilterUser(filterUser, query)

	// Find total rows
	if err := query.Model(model.User{}).
		Count(&totalRows).Error; err != nil {
		return 0, helpers.Wrap(err)
	}

	return int(totalRows), nil
}

func (r *UserRepositoryRead) applyFilterUser(filterUser FilterUser, query *gorm.DB) *gorm.DB {
	if filterUser.Role != "" {
		query = query.Where("role = ?", filterUser.Role)
	}

	if filterUser.IsDeleted == false {
		query = query.Where("deleted_at IS NULL")
	}
	return query
}

// GetAllUser retrieves a all users.
func (r *UserRepositoryRead) GetAllUser() ([]uint, error) {
	var userIDs []uint
	if err := r.transaction.Model(&model.User{}).Select("id").Find(&userIDs).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return userIDs, nil
}

func (r *UserRepositoryRead) GetUserFcmToken(userId uint) ([]*model.FcmToken, error) {
	var fcmTokens []*model.FcmToken
	if err := r.transaction.Where("user_id = ?", userId).Find(&fcmTokens).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return fcmTokens, nil
}

// CheckRoleExists returns bool as flag if role exists.
func (r *UserRepositoryRead) CheckRoleExists(role string) (bool, error) {
	var count int64
	if err := r.transaction.Model(&model.Role{}).Where("code = ?", role).Count(&count).Error; err != nil {
		return false, helpers.Wrap(err)
	}
	return count > 0, nil
}

func (r *UserRepositoryRead) GetAdminUserPrivilegeGroup(userId uint) (string, error) {
	var privilegeGroup string
	if err := r.transaction.Model(&model.UserPrivilegeGroup{}).
		Where("user_id = ?", userId).
		Select("privilege_group_code").
		Scan(&privilegeGroup).Error; err != nil {
		return "", helpers.Wrap(err)
	}
	return privilegeGroup, nil
}
