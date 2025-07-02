package storage

import (
	"fmt"
	"strconv"
)

const (
	ErrEmptyHost   = "field 'Host' is empty"
	ErrEmptyPort   = "field 'Port' is empty"
	ErrInvalidPort = "field 'Port': must be a valid port number (1024-65535)"
	ErrEmptyUser   = "field 'User' is empty"
	ErrEmptyDBName = "field 'Name' is empty"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (c *Config) String() string {
	const mask = "xxxxx"

	user := c.User
	if user != "" {
		user = mask
	}

	pass := c.Password
	if pass != "" {
		pass = mask
	}

	return fmt.Sprintf(
		"{Host: %s, Port: %s, User: %s, Password: %s, Name: %s}",
		c.Host,
		c.Port,
		user,
		pass,
		c.Name)
}

func (c *Config) Validate() []error {
	errs := make([]error, 0)

	if c.Host == "" {
		errs = append(errs, fmt.Errorf(ErrEmptyHost))
	}

	if c.Port == "" {
		errs = append(errs, fmt.Errorf(ErrEmptyPort))
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1024 || port > 65535 {
		errs = append(errs, fmt.Errorf(ErrInvalidPort))
	}

	if c.User == "" {
		errs = append(errs, fmt.Errorf(ErrEmptyUser))
	}

	if c.Name == "" {
		errs = append(errs, fmt.Errorf(ErrEmptyDBName))
	}

	return errs
}
