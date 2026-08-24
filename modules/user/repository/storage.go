package user

import (
	"context"

	"gorm.io/gorm"
)

type UserStorage struct {
	db *gorm.DB
}

// NewUserStorage creates a new instance of UserRepository
func NewUserStorage(db *gorm.DB) *UserStorage {
	return &UserStorage{db: db}
}

// NewUserRepositoryRead creates a new UserRepository with the provided context.
// This repository is used only for read operations.
func (u *UserStorage) NewUserRepositoryRead(ctx context.Context) *UserRepositoryRead {
	return &UserRepositoryRead{transaction: u.db.WithContext(ctx)}
}

// NewUserRepositoryWrite creates a new UserRepository with a new database transaction.
// The transaction is always started when this method is called.
func (u *UserStorage) NewUserRepositoryWrite(ctx context.Context) *UserRepositoryWrite {
	return &UserRepositoryWrite{
		UserRepositoryRead: u.NewUserRepositoryRead(ctx),
		transaction:        u.db.WithContext(ctx).Begin(),
	}
}
