package domain

import "errors"

var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrValidation indicates the request failed domain validation.
	ErrValidation = errors.New("validation failed")

	// ErrInternal indicates an unexpected internal error.
	ErrInternal = errors.New("internal error")
)
