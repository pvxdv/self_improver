package storage

import "errors"

var (
	ErrTrendNotFound = errors.New("trend not found")
	ErrTrendExists   = errors.New("trend exists")
)
