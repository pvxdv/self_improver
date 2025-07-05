package finance

import "errors"

var (
	ErrNegativeAmount     = errors.New("amount must be positive")
	ErrInvalidID          = errors.New("invalid id")
	ErrDescriptionTooLong = errors.New("description len must be less than 1000")
)
