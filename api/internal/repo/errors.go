package repo

import "errors"

var (
	ErrNotFound               = errors.New("resource not found")
	ErrStaleUpdate            = errors.New("resource version is stale")
	ErrDuplicate              = errors.New("duplicate resource")
	ErrForeignKeyViolation    = errors.New("referenced resource does not exist")
	ErrUnitTypeNotFound       = errors.New("unit type not found")
	ErrAgentNotFound          = errors.New("agent not found")
	ErrAgentForbidden         = errors.New("agent is not authorized for this campus")
	ErrAlreadyLinked          = errors.New("user already has a linked agent")
	ErrApplicationConflict    = errors.New("application conflicts with an existing application")
	ErrApplicationResolved    = errors.New("application is no longer pending")
	ErrCampusForbidden        = errors.New("operator is not authorized for this campus")
	ErrPhoneConflict          = errors.New("agent phone number conflicts with an existing agent")
	ErrAgentLifecycleConflict = errors.New("agent lifecycle transition conflicts with the current status")
)
