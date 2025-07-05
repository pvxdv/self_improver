package config

import (
	"fmt"
	"github.com/pvxdv/self_improver/internal/config/telegram"
	"reflect"
	"strings"

	"github.com/pvxdv/self_improver/internal/config/app"
	"github.com/pvxdv/self_improver/internal/config/http"
	"github.com/pvxdv/self_improver/internal/config/storage"
)

type Config struct {
	App      *app.Config
	HTTP     *http.Config
	Storage  *storage.Config
	Telegram *telegram.Config
}

type Loader interface {
	Load() (*Config, error)
}

type Validator interface {
	Validate() []error
}

func LoadAndValidate(loader Loader) (*Config, error) {
	cfg, err := loader.Load()
	if err != nil {
		return nil, err
	}

	errorsBySection := cfg.collectValidationErrors()
	if len(errorsBySection) > 0 {
		return nil, formatValidationError(errorsBySection)
	}

	return cfg, nil
}

func (c *Config) collectValidationErrors() map[string][]error {
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()

	result := make(map[string][]error)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.IsNil() {
			continue
		}

		if validator, ok := fieldVal.Interface().(Validator); ok {
			if errs := validator.Validate(); len(errs) > 0 {
				result[field.Name] = errs
			}
		}
	}

	return result
}

func formatValidationError(errors map[string][]error) error {
	var formatted []string

	for section, errs := range errors {
		formatted = append(formatted, fmt.Sprintf("→ %s:", section))
		for _, err := range errs {
			formatted = append(formatted, fmt.Sprintf("    - %s", err.Error()))
		}
		formatted = append(formatted, "")
	}

	return fmt.Errorf("\n%s", strings.Join(formatted, "\n"))
}
