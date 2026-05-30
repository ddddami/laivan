package repo

import "errors"

var (
	ErrNotFound = errors.New("resource not found")
	ErrDuplicate = errors.New("duplicate resource")
)
