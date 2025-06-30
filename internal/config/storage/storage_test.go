package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		wantErrors []string
	}{
		{
			name: "valid config",
			config: Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "admin",
				Password: "secret",
				Name:     "mydb",
			},
			wantErrors: nil,
		},
		{
			name: "empty host",
			config: Config{
				Host: "",
				Port: "5432",
				User: "admin",
				Name: "mydb",
			},
			wantErrors: []string{ErrEmptyHost},
		},
		{
			name: "empty port",
			config: Config{
				Host: "localhost",
				Port: "",
				User: "admin",
				Name: "mydb",
			},
			wantErrors: []string{ErrEmptyPort, ErrInvalidPort},
		},
		{
			name: "invalid port format",
			config: Config{
				Host: "localhost",
				Port: "not-a-number",
				User: "admin",
				Name: "mydb",
			},
			wantErrors: []string{ErrInvalidPort},
		},
		{
			name: "port too low",
			config: Config{
				Host: "localhost",
				Port: "1023",
				User: "admin",
				Name: "mydb",
			},
			wantErrors: []string{ErrInvalidPort},
		},
		{
			name: "port too high",
			config: Config{
				Host: "localhost",
				Port: "65536",
				User: "admin",
				Name: "mydb",
			},
			wantErrors: []string{ErrInvalidPort},
		},
		{
			name: "empty user",
			config: Config{
				Host: "localhost",
				Port: "5432",
				User: "",
				Name: "mydb",
			},
			wantErrors: []string{ErrEmptyUser},
		},
		{
			name: "empty db name",
			config: Config{
				Host: "localhost",
				Port: "5432",
				User: "admin",
				Name: "",
			},
			wantErrors: []string{ErrEmptyDBName},
		},
		{
			name: "multiple errors",
			config: Config{
				Host: "",
				Port: "0",
				User: "",
				Name: "",
			},
			wantErrors: []string{
				ErrEmptyHost,
				ErrInvalidPort,
				ErrEmptyUser,
				ErrEmptyDBName,
			},
		},
		{
			name: "password can be empty",
			config: Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "admin",
				Password: "",
				Name:     "mydb",
			},
			wantErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Validate()

			if tt.wantErrors == nil {
				assert.Empty(t, errs)
				return
			}

			require.Equal(t, len(tt.wantErrors), len(errs), "unexpected number of errors")

			for i, wantErr := range tt.wantErrors {
				assert.ErrorContains(t, errs[i], wantErr)
			}
		})
	}
}

func TestPortBoundaries(t *testing.T) {
	// Test valid boundary ports
	validPorts := []string{"1024", "3000", "65535"}
	for _, port := range validPorts {
		t.Run(port, func(t *testing.T) {
			cfg := Config{
				Host: "localhost",
				Port: port,
				User: "user",
				Name: "db",
			}
			assert.Empty(t, cfg.Validate())
		})
	}

	// Test invalid boundary ports
	invalidPorts := []string{"1023", "65536"}
	for _, port := range invalidPorts {
		t.Run(port, func(t *testing.T) {
			cfg := Config{
				Host: "localhost",
				Port: port,
				User: "user",
				Name: "db",
			}
			errs := cfg.Validate()
			require.Len(t, errs, 1)
			assert.ErrorContains(t, errs[0], ErrInvalidPort)
		})
	}
}
