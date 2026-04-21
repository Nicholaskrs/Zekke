package helpers

import (
	"errors"

	"gorm.io/gorm"
)

// ErrIsNotFound used to check if the error is error not found from gorm.
func ErrIsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
