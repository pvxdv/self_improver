package telegram

import "fmt"

const (
	ErrEmptyToken = "field 'Token' is empty"
	ErrEmptyPass  = "field 'Password' is empty"
)

type Config struct {
	Token    string // Telegram bot connect token
	Password string
}

func (c *Config) String() string {
	const mask = "xxxxx"

	token := c.Token
	if token != "" {
		token = mask
	}

	pass := c.Password
	if token != "" {
		token = mask
	}

	return fmt.Sprintf(
		"{Token: %s, Password: %s}",
		c.Token, pass,
	)
}

func (c *Config) Validate() []error {
	errs := make([]error, 0)

	if c.Token == "" {
		errs = append(errs, fmt.Errorf(ErrEmptyToken))
	}

	if c.Password == "" {
		errs = append(errs, fmt.Errorf(ErrEmptyPass))
	}

	return errs
}
