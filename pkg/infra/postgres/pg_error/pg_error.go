package pg_error

import "errors"

var (
	ErrNotFound = errors.New("not found")
)