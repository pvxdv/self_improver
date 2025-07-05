package telegram

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
				Password: "evmepvmwvqvqvrbtn",
				Token:    "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			},
			wantErrors: nil,
		},
		{
			name: "invalid config",
			config: Config{
				Token:    "",
				Password: "",
			},
			wantErrors: []string{
				ErrEmptyToken, ErrEmptyPass,
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
