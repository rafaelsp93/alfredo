package domain

import "errors"

var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrValidation indicates the request failed domain validation.
	ErrValidation = errors.New("validation failed")

	// ErrInternal indicates an unexpected internal error.
	ErrInternal = errors.New("internal error")

	// ErrAlreadyStopped is returned when a treatment has already been stopped.
	ErrAlreadyStopped = errors.New("already stopped")
)
