package store

import "errors"

var (
	ErrNotFound = errors.New("store: resource not found")
	ErrConflict = errors.New("store: resource already exists")
)
