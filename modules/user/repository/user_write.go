package user

import (
	"context"
	"template-go/base/helpers"
	"template-go/core/telemetry/trace"
	"template-go/data/model"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

type UserRepositoryWrite struct {
	*UserRepositoryRead
	transaction *gorm.DB
}

// Commit is used to commit database changes.
func (repo *UserRepositoryWrite) Commit(ctx context.Context) error {
	ctx, span := trace.Tracer("Zekke").Start(ctx, "UserCommitTx")
	defer span.End()
	transaction := repo.transaction.Commit()
	if transaction.Error != nil {
		return transaction.Error
	}
	return nil
}

// Rollback is used to rollback database changes.
func (repo *UserRepositoryWrite) Rollback(ctx context.Context) {
	ctx, span := trace.Tracer("Zekke").Start(ctx, "UserRollbackTx")
	defer span.End()
	repo.transaction.Rollback()
}

func (repo *UserRepositoryWrite) LockUser(userId uint) (*model.User, error) {
	var user model.User
	if err := repo.transaction.Clauses(
		clause.Locking{
			Strength: "UPDATE",
		},
	).Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// CreateUser used to insert new row data. It returns inserted ID and error
func (repo *UserRepositoryWrite) CreateUser(param *model.User) (uint, error) {
	if err := repo.transaction.Create(&param).Error; err != nil {
		return 0, helpers.Wrap(err)
	}
	return param.ID, nil
}

func (repo *UserRepositoryWrite) UpdateUser(param *model.User) error {
	return repo.transaction.Model(param).Where("id = ?", param.ID).Updates(param).Error
}

func (repo *UserRepositoryWrite) CreateFcmToken(param *model.FcmToken) error {
	if err := repo.transaction.Create(param).Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}

func (repo *UserRepositoryWrite) DeleteFcmTokenBulk(tokens []string) error {
	if err := repo.transaction.Where("fcm_token IN (?)", tokens).Delete(&model.FcmToken{}).Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}
