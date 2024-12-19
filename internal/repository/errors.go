package repository

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInternalServer = errors.New("internal server error")
	ErrBadRequest     = errors.New("bad request")
)
