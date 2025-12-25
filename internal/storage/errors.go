package storage

import "errors"

var (
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
)
