// internal/fitness/domain/errors.go
package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrValidation    = errors.New("validation failed")
	ErrAlreadyExists = errors.New("already exists")
	ErrAlreadyAchieved = errors.New("already achieved")
)
