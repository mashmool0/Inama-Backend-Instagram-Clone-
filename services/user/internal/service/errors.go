package service

import "errors"

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidCursor   = errors.New("invalid cursor")
)

type InvalidArgumentError struct {
	message string
}

func (e InvalidArgumentError) Error() string {
	return e.message
}

func NewInvalidArgument(message string) error {
	return InvalidArgumentError{message: message}
}
