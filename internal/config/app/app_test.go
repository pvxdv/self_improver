package app

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
			name: "valid local config with debug",
			config: Config{
				Env:   envLocal,
				Debug: true,
			},
			wantErrors: nil,
		},
		{
			name: "valid prod config without debug",
			config: Config{
				Env:   envProd,
				Debug: false,
			},
			wantErrors: nil,
		},
		{
			name: "empty env",
			config: Config{
				Env: "",
			},
			wantErrors: []string{
				errEmptyEnv,
				errInvalidEnv,
			},
		},
		{
			name: "invalid env value",
			config: Config{
				Env: "staging",
			},
			wantErrors: []string{
				errInvalidEnv,
			},
		},
		{
			name: "debug mode in production",
			config: Config{
				Env:   envProd,
				Debug: true,
			},
			wantErrors: []string{
				errProdDebug,
			},
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

func TestIsValidEnv(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{envLocal, true},
		{envDev, true},
		{envProd, true},
		{"staging", false},
		{"", false},
		{"LOCAL", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidEnv(tt.env))
		})
	}
}
