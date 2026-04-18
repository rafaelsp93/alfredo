package domain

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation failed")
)
