package permission

import (
	"context"
	"template-go/base/helpers"
	"template-go/core/telemetry/trace"
	"template-go/data/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PermissionRepositoryWrite struct {
	*PermissionRepositoryRead
	transaction *gorm.DB
}

// Commit is used to commit database changes.
func (w *PermissionRepositoryWrite) Commit(ctx context.Context) error {
	_, span := trace.Tracer("Zekke").Start(ctx, "PermissionCommitTx")
	defer span.End()
	if err := w.transaction.Commit().Error; err != nil {
		return err
	}
	return nil
}

// Rollback is used to rollback database changes.
func (w *PermissionRepositoryWrite) Rollback(ctx context.Context) {
	_, span := trace.Tracer("Zekke").Start(ctx, "PermissionRollbackTx")
	defer span.End()
	w.transaction.Rollback()
}

func (w *PermissionRepositoryWrite) UpsertUserPrivilegeGroup(privilegeGroup *model.UserPrivilegeGroup) error {
	if err := w.transaction.
		Model(&model.UserPrivilegeGroup{}).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"privilege_group_code"}),
			},
		).Create(&privilegeGroup).Error; err != nil {
		return helpers.Wrap(err)
	}

	return nil
}

func (w *PermissionRepositoryWrite) AssignPrivilegeToPrivilegeGroup(param *model.PrivilegeGroupPrivilege) error {
	if err := w.transaction.Create(&param).Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil

}

func (w *PermissionRepositoryWrite) InsertPrivilegeGroup(param *model.PrivilegeGroup) error {
	if err := w.transaction.Create(&param).Error; err != nil {
		return helpers.Wrap(err)
	}
	return nil

}
