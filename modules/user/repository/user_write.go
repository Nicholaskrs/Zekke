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
func (w *UserRepositoryWrite) Commit(ctx context.Context) error {
	ctx, span := trace.Tracer("Zekke").Start(ctx, "UserCommitTx")
	defer span.End()
	transaction := w.transaction.Commit()
	if transaction.Error != nil {
		return transaction.Error
	}
	return nil
}

// Rollback is used to rollback database changes.
func (w *UserRepositoryWrite) Rollback(ctx context.Context) {
	ctx, span := trace.Tracer("Zekke").Start(ctx, "UserRollbackTx")
	defer span.End()
	w.transaction.Rollback()
}

func (w *UserRepositoryWrite) LockUser(userId uint) (*model.User, error) {
	var user model.User
	if err := w.transaction.Clauses(
		clause.Locking{
			Strength: "UPDATE",
		},
	).Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &user, nil
}

// CreateUser used to insert new row data. It returns inserted ID and error
func (w *UserRepositoryWrite) CreateUser(param *model.User) (uint, error) {
	if err := w.transaction.Create(&param).Error; err != nil {
		return 0, helpers.Wrap(err)
	}
	return param.ID, nil
}

func (w *UserRepositoryWrite) UpdateUser(param *model.User) error {
	return w.transaction.Model(param).Where("id = ?", param.ID).Updates(param).Error
}

func (w *UserRepositoryWrite) CreateFcmToken(param *model.FcmToken) error {
	if err := w.transaction.Create(param).Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}

func (w *UserRepositoryWrite) DeleteFcmTokenBulk(tokens []string) error {
	if err := w.transaction.Where("fcm_token IN (?)", tokens).Delete(&model.FcmToken{}).Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}

func (w *UserRepositoryWrite) InsertRefreshToken(param *model.RefreshToken) error {
	if err := w.transaction.Create(param).Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}

func (w *UserRepositoryWrite) LockRefreshToken(userId int, tokenHash string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	if err := w.transaction.Clauses(
		clause.Locking{
			Strength: "UPDATE",
		},
	).
		Where("token_hash = ?", tokenHash).
		Where("user_id = ?", userId).
		First(&refreshToken).Error; err != nil {
		return nil, helpers.Wrap(err)
	}
	return &refreshToken, nil
}

func (w *UserRepositoryWrite) UseRefreshToken(id int) error {
	var refreshToken model.RefreshToken
	if err := w.transaction.
		Model(&refreshToken).
		Where("id = ?", id).
		Update("used_at", gorm.Expr("now()")).
		Update("last_updated_ts", gorm.Expr("now()")).
		Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}

func (w *UserRepositoryWrite) RevokeRefreshTokenByFamilyID(familyID string) error {
	var refreshToken model.RefreshToken
	if err := w.transaction.
		Model(&refreshToken).
		Where("family_id = ?", familyID).
		Update("revoked_at", gorm.Expr("now()")).
		Update("last_updated_ts", gorm.Expr("now()")).
		Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}

func (w *UserRepositoryWrite) RevokeRefreshTokenByUserID(userID uint) error {
	var refreshToken model.RefreshToken
	if err := w.transaction.
		Model(&refreshToken).
		Where("user_id = ?", userID).
		Update("revoked_at", gorm.Expr("now()")).
		Update("last_updated_ts", gorm.Expr("now()")).
		Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil
}
