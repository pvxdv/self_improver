package http

import (
	"fmt"
	"time"
)

const (
	MinTimeout     = 10 * time.Millisecond
	MaxTimeout     = 5 * time.Second
	MinIdleTimeout = 60 * time.Second
	MaxIdleTimeout = 300 * time.Second

	errEmptyAddress       = "field 'Address' is empty"
	errInvalidTimeout     = "field 'Timeout': must be >= 10ms && <= 5s"
	errInvalidIdleTimeout = "field 'IdleTimeout': must be >= 60s && <= 300s"
)

type Config struct {
	Address     string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

func (c *Config) Validate() []error {
	errs := make([]error, 0)

	if c.Address == "" {
		errs = append(errs, fmt.Errorf(errEmptyAddress))
	}
	if c.Timeout < MinTimeout || c.Timeout > MaxTimeout {
		errs = append(errs, fmt.Errorf(errInvalidTimeout))
	}
	if c.IdleTimeout < MinIdleTimeout || c.IdleTimeout > MaxIdleTimeout {
		errs = append(errs, fmt.Errorf(errInvalidIdleTimeout))
	}

	return errs
}
