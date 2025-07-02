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

	errEmptyPort          = "field 'Port' is empty"
	errInvalidTimeout     = "field 'Timeout': must be >= 10ms && <= 5s"
	errInvalidIdleTimeout = "field 'IdleTimeout': must be >= 60s && <= 300s"
)

type Config struct {
	Port        string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"{Port: %s, Timeout: %s, IdleTimeout: %s}",
		c.Port,
		c.Timeout,
		c.IdleTimeout)
}

func (c *Config) Validate() []error {
	errs := make([]error, 0)

	if c.Port == "" {
		errs = append(errs, fmt.Errorf(errEmptyPort))
	}
	if c.Timeout < MinTimeout || c.Timeout > MaxTimeout {
		errs = append(errs, fmt.Errorf(errInvalidTimeout))
	}
	if c.IdleTimeout < MinIdleTimeout || c.IdleTimeout > MaxIdleTimeout {
		errs = append(errs, fmt.Errorf(errInvalidIdleTimeout))
	}

	return errs
}
