package app

import "fmt"

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"

	errEmptyEnv   = "field 'Env' is empty"
	errInvalidEnv = "field 'Env': must be one of: local, dev, prod"
	errProdDebug  = "debug mode is not allowed in production"
)

type Config struct {
	Env   string // Application environment (local, dev, prod)
	Debug bool   // Debug mode flag
}

func (c *Config) Validate() []error {
	errs := make([]error, 0)

	if c.Env == "" {
		errs = append(errs, fmt.Errorf(errEmptyEnv))
	}

	if !isValidEnv(c.Env) {
		errs = append(errs, fmt.Errorf(errInvalidEnv))
	}

	if c.Env == envProd && c.Debug {
		errs = append(errs, fmt.Errorf(errProdDebug))
	}

	return errs
}

func isValidEnv(env string) bool {
	validEnvs := map[string]bool{envLocal: true, envProd: true, envDev: true}
	return validEnvs[env]
}
