package repo

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrDuplicate           = errors.New("duplicate resource")
	ErrForeignKeyViolation = errors.New("referenced resource does not exist")
)
