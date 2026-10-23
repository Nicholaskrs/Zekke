package permission

import (
	"context"

	"gorm.io/gorm"
)

type PermissionStorage struct {
	db *gorm.DB
}

// NewPermissionStorage creates a new instance of PermissionStorage
func NewPermissionStorage(db *gorm.DB) *PermissionStorage {
	return &PermissionStorage{db: db}
}

// NewPermissionRepositoryRead creates a new UserRepository with the provided context.
// This repository is used only for read operations.
func (u *PermissionStorage) NewPermissionRepositoryRead(ctx context.Context) *PermissionRepositoryRead {
	return &PermissionRepositoryRead{transaction: u.db.WithContext(ctx)}
}

// NewPermissionRepositoryWrite creates a new PermissionRepository with a new database transaction.
// The transaction is always started when this method is called.
func (u *PermissionStorage) NewPermissionRepositoryWrite(ctx context.Context) *PermissionRepositoryWrite {
	return &PermissionRepositoryWrite{
		PermissionRepositoryRead: u.NewPermissionRepositoryRead(ctx),
		transaction:              u.db.WithContext(ctx).Begin(),
	}
}
